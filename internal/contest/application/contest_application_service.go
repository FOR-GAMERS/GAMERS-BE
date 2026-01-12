package application

import (
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	"GAMERS-BE/internal/global/exception"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	"context"
	"errors"
	"time"
)

type ContestApplicationService struct {
	applicationRepo  port.ContestApplicationRedisPort
	contestRepo      port.ContestDatabasePort
	memberRepo       port.ContestMemberDatabasePort
	eventPublisher   port.EventPublisherPort
	oauth2Repository oauth2Port.OAuth2DatabasePort
}

func NewContestApplicationService(
	applicationRepo port.ContestApplicationRedisPort,
	contestRepo port.ContestDatabasePort,
	memberRepo port.ContestMemberDatabasePort,
	eventPublisher port.EventPublisherPort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
) *ContestApplicationService {
	return &ContestApplicationService{
		applicationRepo:  applicationRepo,
		contestRepo:      contestRepo,
		memberRepo:       memberRepo,
		eventPublisher:   eventPublisher,
		oauth2Repository: oauth2Repository,
	}
}

// RequestParticipate - Contest 참가 신청
func (s *ContestApplicationService) RequestParticipate(ctx context.Context, contestId, userId int64) (*dto.DiscordLinkRequiredResponse, error) {
	// Check if user has linked Discord account
	_, err := s.oauth2Repository.FindDiscordAccountByUserId(userId)
	if err != nil {
		if errors.Is(err, exception.ErrDiscordUserCannotFound) {
			return dto.NewDiscordLinkRequiredResponse("Discord account linking is required to apply for this contest"), exception.ErrDiscordLinkRequired
		}
		return nil, err
	}

	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	if contest.ContestStatus != "PENDING" {
		return nil, exception.ErrCannotAcceptApplication
	}

	if !contest.IsBeforeStartTime() {
		return nil, exception.ErrContestAlreadyStarted
	}

	ttl := time.Until(contest.StartedAt)
	if ttl < 0 {
		ttl = 24 * time.Hour
	}

	err = s.applicationRepo.RequestParticipate(ctx, contestId, userId, ttl)
	if err != nil {
		return nil, err
	}

	// 이벤트 발행 (비동기 - 실패해도 신청은 완료됨)
	go s.publishApplicationRequestedEvent(context.Background(), contest, userId)

	return nil, nil
}

func (s *ContestApplicationService) checkLeaderPermission(contestId, userId int64) error {
	member, err := s.memberRepo.GetByContestAndUser(contestId, userId)
	if err != nil {
		return exception.ErrInvalidAccess
	}
	if !member.IsLeader() {
		return exception.ErrContestAlreadyStarted
	}

	return nil
}

// AcceptApplication - 신청 승인 (Leader만 가능)
func (s *ContestApplicationService) AcceptApplication(ctx context.Context, contestId, userId, leaderUserId int64) error {
	// Contest 존재 확인
	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return err
	}

	// Contest가 PENDING 상태인지 확인
	if contest.ContestStatus != "PENDING" {
		return exception.ErrCannotAcceptApplication
	}

	// Leader 권한 확인
	if err := s.checkLeaderPermission(contestId, leaderUserId); err != nil {
		return err
	}

	// 신청 승인 (Redis)
	err = s.applicationRepo.AcceptRequest(ctx, contestId, userId, leaderUserId)
	if err != nil {
		return err
	}

	// DB에 멤버 추가 (Withdraw 기능을 위해 즉시 동기화)
	member := domain.NewContestMember(userId, contestId, domain.MemberTypeNormal, domain.LeaderTypeMember)
	if err := s.memberRepo.Save(member); err != nil {
		// DB 저장 실패 시 Redis 상태 롤백은 하지 않음 (최종적 일관성)
		// 추후 MigrateAcceptedApplicationsToDatabase에서 재시도됨
		_ = err
	}

	// 이벤트 발행 (비동기)
	go s.publishApplicationAcceptedEvent(context.Background(), contest, userId, leaderUserId)

	return nil
}

// RejectApplication - 신청 거절 (Leader만 가능)
func (s *ContestApplicationService) RejectApplication(ctx context.Context, contestId, userId, leaderUserId int64) error {
	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return err
	}

	// Contest가 PENDING 상태인지 확인
	if contest.ContestStatus != "PENDING" {
		return exception.ErrCannotAcceptApplication
	}

	// Leader 권한 확인
	if err := s.checkLeaderPermission(contestId, leaderUserId); err != nil {
		return err
	}

	// 신청 거절
	err = s.applicationRepo.RejectRequest(ctx, contestId, userId, leaderUserId)
	if err != nil {
		return err
	}

	// 이벤트 발행 (비동기)
	go s.publishApplicationRejectedEvent(context.Background(), contest, userId, leaderUserId)

	return nil
}

// GetPendingApplications - Pending 신청 목록 조회
func (s *ContestApplicationService) GetPendingApplications(ctx context.Context, contestId int64) ([]*port.ContestApplication, error) {
	// Contest 존재 확인
	_, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	return s.applicationRepo.GetPendingApplications(ctx, contestId)
}

// GetMyApplication - 내 신청 정보 조회
func (s *ContestApplicationService) GetMyApplication(ctx context.Context, contestId, userId int64) (*port.ContestApplication, error) {
	return s.applicationRepo.GetApplication(ctx, contestId, userId)
}

// CancelApplication - Cancel a pending application (by user themselves)
func (s *ContestApplicationService) CancelApplication(ctx context.Context, contestId, userId int64) error {
	// Verify contest exists
	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return err
	}

	// Verify contest is in PENDING status
	if contest.ContestStatus != "PENDING" {
		return exception.ErrCannotAcceptApplication
	}

	// Cancel the application in Redis
	err = s.applicationRepo.CancelApplication(ctx, contestId, userId)
	if err != nil {
		return err
	}

	// Publish event (async)
	go s.publishApplicationCancelledEvent(context.Background(), contest, userId)

	return nil
}

// WithdrawFromContest - 대회 탈퇴 (멤버 본인만 가능, 리더는 불가)
func (s *ContestApplicationService) WithdrawFromContest(ctx context.Context, contestId, userId int64) error {
	// Contest 존재 확인
	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return err
	}

	// Contest가 PENDING 상태인지 확인
	if contest.ContestStatus != "PENDING" {
		return exception.ErrCannotAcceptApplication
	}

	// 멤버 확인
	member, err := s.memberRepo.GetByContestAndUser(contestId, userId)
	if err != nil {
		return err
	}

	// 리더는 탈퇴 불가
	if member.IsLeader() {
		return exception.ErrLeaderCannotWithdraw
	}

	// 멤버 삭제
	if err := s.memberRepo.DeleteById(contestId, userId); err != nil {
		return err
	}

	// 이벤트 발행 (비동기)
	go s.publishMemberWithdrawnEvent(context.Background(), contest, userId)

	return nil
}

func (s *ContestApplicationService) MigrateAcceptedApplicationsToDatabase(ctx context.Context, contestId int64) error {
	acceptedUserIDs, err := s.applicationRepo.GetAcceptedApplications(ctx, contestId)
	if err != nil {
		return err
	}

	if len(acceptedUserIDs) == 0 {
		return s.applicationRepo.ClearApplications(ctx, contestId)
	}

	// Accept 시점에 이미 DB에 저장되지 않은 멤버만 필터링
	members := make([]*domain.ContestMember, 0, len(acceptedUserIDs))
	for _, userId := range acceptedUserIDs {
		// 이미 DB에 존재하는지 확인
		_, err := s.memberRepo.GetByContestAndUser(contestId, userId)
		if err == nil {
			// 이미 존재하면 건너뜀
			continue
		}
		member := domain.NewContestMember(userId, contestId, domain.MemberTypeNormal, domain.LeaderTypeMember)
		members = append(members, member)
	}

	// 새로 추가할 멤버가 있는 경우에만 저장
	if len(members) > 0 {
		if err := s.memberRepo.SaveBatch(members); err != nil {
			return err
		}
	}

	return s.applicationRepo.ClearApplications(ctx, contestId)
}

func (s *ContestApplicationService) getDiscordIdByUserId(userId int64) string {
	discordAccount, err := s.oauth2Repository.FindDiscordAccountByUserId(userId)
	if err != nil {
		return ""
	}
	return discordAccount.DiscordId
}

func getStringFromPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// publishApplicationRequestedEvent - 신청 요청 이벤트 발행
func (s *ContestApplicationService) publishApplicationRequestedEvent(
	ctx context.Context,
	contest *domain.Contest,
	userId int64,
) {
	discordUserId := s.getDiscordIdByUserId(userId)

	event := &port.ContestApplicationEvent{
		EventType:            port.EventTypeApplicationRequested,
		ContestID:            contest.ContestID,
		UserID:               userId,
		DiscordUserID:        discordUserId,
		DiscordGuildID:       getStringFromPtr(contest.DiscordGuildId),
		DiscordTextChannelID: getStringFromPtr(contest.DiscordTextChannelId),
		Data: map[string]interface{}{
			"contest_title": contest.Title,
			"status":        "PENDING",
			"requested_at":  time.Now(),
		},
	}

	if err := s.eventPublisher.PublishContestApplicationEvent(ctx, event); err != nil {
		_ = err
	}
}

// publishApplicationAcceptedEvent - 신청 승인 이벤트 발행
func (s *ContestApplicationService) publishApplicationAcceptedEvent(
	ctx context.Context,
	contest *domain.Contest,
	userId int64,
	processedBy int64,
) {
	discordUserId := s.getDiscordIdByUserId(userId)
	processedByDiscordId := s.getDiscordIdByUserId(processedBy)

	event := &port.ContestApplicationEvent{
		EventType:            port.EventTypeApplicationAccepted,
		ContestID:            contest.ContestID,
		UserID:               userId,
		DiscordUserID:        discordUserId,
		DiscordGuildID:       getStringFromPtr(contest.DiscordGuildId),
		DiscordTextChannelID: getStringFromPtr(contest.DiscordTextChannelId),
		Data: map[string]interface{}{
			"contest_title":           contest.Title,
			"status":                  "ACCEPTED",
			"processed_by":            processedBy,
			"processed_by_discord_id": processedByDiscordId,
			"processed_at":            time.Now(),
		},
	}

	if err := s.eventPublisher.PublishContestApplicationEvent(ctx, event); err != nil {
		// 로그만 남기고 에러는 무시
		_ = err
	}
}

// publishApplicationRejectedEvent - 신청 거절 이벤트 발행
func (s *ContestApplicationService) publishApplicationRejectedEvent(
	ctx context.Context,
	contest *domain.Contest,
	userId int64,
	processedBy int64,
) {
	discordUserId := s.getDiscordIdByUserId(userId)
	processedByDiscordId := s.getDiscordIdByUserId(processedBy)

	event := &port.ContestApplicationEvent{
		EventType:            port.EventTypeApplicationRejected,
		ContestID:            contest.ContestID,
		UserID:               userId,
		DiscordUserID:        discordUserId,
		DiscordGuildID:       getStringFromPtr(contest.DiscordGuildId),
		DiscordTextChannelID: getStringFromPtr(contest.DiscordTextChannelId),
		Data: map[string]interface{}{
			"contest_title":           contest.Title,
			"status":                  "REJECTED",
			"processed_by":            processedBy,
			"processed_by_discord_id": processedByDiscordId,
			"processed_at":            time.Now(),
		},
	}

	if err := s.eventPublisher.PublishContestApplicationEvent(ctx, event); err != nil {
		// 로그만 남기고 에러는 무시
		_ = err
	}
}

// publishMemberWithdrawnEvent - 멤버 탈퇴 이벤트 발행
func (s *ContestApplicationService) publishMemberWithdrawnEvent(
	ctx context.Context,
	contest *domain.Contest,
	userId int64,
) {
	discordUserId := s.getDiscordIdByUserId(userId)

	event := &port.ContestApplicationEvent{
		EventType:            port.EventTypeMemberWithdrawn,
		ContestID:            contest.ContestID,
		UserID:               userId,
		DiscordUserID:        discordUserId,
		DiscordGuildID:       getStringFromPtr(contest.DiscordGuildId),
		DiscordTextChannelID: getStringFromPtr(contest.DiscordTextChannelId),
		Data: map[string]interface{}{
			"contest_title": contest.Title,
			"status":        "WITHDRAWN",
			"withdrawn_at":  time.Now(),
		},
	}

	if err := s.eventPublisher.PublishContestApplicationEvent(ctx, event); err != nil {
		// 로그만 남기고 에러는 무시
		_ = err
	}
}

// publishApplicationCancelledEvent - Publish application cancelled event
func (s *ContestApplicationService) publishApplicationCancelledEvent(
	ctx context.Context,
	contest *domain.Contest,
	userId int64,
) {
	discordUserId := s.getDiscordIdByUserId(userId)

	event := &port.ContestApplicationEvent{
		EventType:            port.EventTypeApplicationCancelled,
		ContestID:            contest.ContestID,
		UserID:               userId,
		DiscordUserID:        discordUserId,
		DiscordGuildID:       getStringFromPtr(contest.DiscordGuildId),
		DiscordTextChannelID: getStringFromPtr(contest.DiscordTextChannelId),
		Data: map[string]interface{}{
			"contest_title": contest.Title,
			"status":        "CANCELLED",
			"cancelled_at":  time.Now(),
		},
	}

	if err := s.eventPublisher.PublishContestApplicationEvent(ctx, event); err != nil {
		_ = err
	}
}
