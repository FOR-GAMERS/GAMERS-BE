package application

import (
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/exception"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"
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
	userQueryRepo    userQueryPort.UserQueryPort
}

func NewContestApplicationService(
	applicationRepo port.ContestApplicationRedisPort,
	contestRepo port.ContestDatabasePort,
	memberRepo port.ContestMemberDatabasePort,
	eventPublisher port.EventPublisherPort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
) *ContestApplicationService {
	return &ContestApplicationService{
		applicationRepo:  applicationRepo,
		contestRepo:      contestRepo,
		memberRepo:       memberRepo,
		eventPublisher:   eventPublisher,
		oauth2Repository: oauth2Repository,
		userQueryRepo:    userQueryRepo,
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

	// Fetch user info and create sender snapshot
	user, err := s.userQueryRepo.FindById(userId)
	if err != nil {
		return nil, err
	}

	senderSnapshot := &port.SenderSnapshot{
		UserID:   user.Id,
		Username: user.Username,
		Tag:      user.Tag,
		Avatar:   user.Avatar,
	}

	ttl := time.Until(contest.StartedAt)
	if ttl < 0 {
		ttl = 24 * time.Hour
	}

	err = s.applicationRepo.RequestParticipate(ctx, contestId, senderSnapshot, ttl)
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
		return exception.ErrPermissionDenied
	}

	return nil
}

// AcceptApplication - 신청 승인 (Leader만 가능)
func (s *ContestApplicationService) AcceptApplication(ctx context.Context, contestId, userId, leaderUserId int64) error {
	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return err
	}

	if contest.ContestStatus != "PENDING" {
		return exception.ErrCannotAcceptApplication
	}

	if err := s.checkLeaderPermission(contestId, leaderUserId); err != nil {
		return err
	}

	err = s.applicationRepo.AcceptRequest(ctx, contestId, userId, leaderUserId)
	if err != nil {
		return err
	}

	member := domain.NewContestMember(userId, contestId, domain.MemberTypeNormal, domain.LeaderTypeMember)
	if err := s.memberRepo.Save(member); err != nil {
		// DB 저장 실패 시 Redis 상태 롤백은 하지 않음 (최종적 일관성)
		// 추후 MigrateAcceptedApplicationsToDatabase에서 재시도됨
		_ = err
	}

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

// GetMyContestStatus - 내 대회 상태 조회 (리더인지, 멤버인지, 지원했는지 등)
func (s *ContestApplicationService) GetMyContestStatus(ctx context.Context, contestId, userId int64) (*dto.UserContestStatusResponse, error) {
	// Contest 존재 확인
	_, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	response := &dto.UserContestStatusResponse{
		IsLeader:   false,
		IsMember:   false,
		HasApplied: false,
	}

	// Check if user is a member of the contest
	member, err := s.memberRepo.GetByContestAndUser(contestId, userId)
	if err == nil && member != nil {
		response.IsMember = true
		response.IsLeader = member.IsLeader()
		memberType := string(member.MemberType)
		response.MemberType = &memberType
		return response, nil
	}

	// If not a member, check if user has applied
	hasApplied, err := s.applicationRepo.HasApplied(ctx, contestId, userId)
	if err != nil {
		// Ignore error and assume not applied
		return response, nil
	}

	if hasApplied {
		response.HasApplied = true
		// Get application status
		application, err := s.applicationRepo.GetApplication(ctx, contestId, userId)
		if err == nil && application != nil {
			response.ApplicationStatus = &application.Status
		}
	}

	return response, nil
}

// GetContestMembers - Contest 참여 멤버 목록 조회 (Pagination)
func (s *ContestApplicationService) GetContestMembers(
	ctx context.Context,
	contestId int64,
	pagination *commonDto.PaginationRequest,
	sort *commonDto.SortRequest,
) (*commonDto.PaginationResponse, error) {
	// Contest 존재 확인
	_, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	// Get members with pagination
	members, totalCount, err := s.memberRepo.GetMembersWithUserByContest(contestId, pagination, sort)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	memberResponses := dto.ToContestMemberResponses(members)

	return commonDto.NewPaginationResponse(
		memberResponses,
		pagination.Page,
		pagination.PageSize,
		totalCount,
	), nil
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

// ChangeMemberRole - 멤버 역할 변경 (Leader만 가능)
func (s *ContestApplicationService) ChangeMemberRole(contestId, targetUserId, leaderUserId int64, newMemberType domain.MemberType) (*dto.ChangeMemberRoleResponse, error) {
	// Contest 존재 확인
	contest, err := s.contestRepo.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	// Contest가 PENDING 상태인지 확인 (대회 시작 전에만 역할 변경 가능)
	if contest.ContestStatus != "PENDING" {
		return nil, exception.ErrContestNotPending
	}

	// Leader 권한 확인
	if err := s.checkLeaderPermission(contestId, leaderUserId); err != nil {
		return nil, err
	}

	// 대상 멤버 확인
	targetMember, err := s.memberRepo.GetByContestAndUser(contestId, targetUserId)
	if err != nil {
		return nil, err
	}

	// Leader의 역할은 변경 불가
	if targetMember.IsLeader() {
		return nil, exception.ErrCannotChangeLeaderRole
	}

	// 이미 같은 역할인 경우
	if targetMember.MemberType == newMemberType {
		return nil, exception.ErrAlreadySameMemberType
	}

	// 역할 변경
	if err := s.memberRepo.UpdateMemberType(contestId, targetUserId, newMemberType); err != nil {
		return nil, err
	}

	return &dto.ChangeMemberRoleResponse{
		UserID:     targetUserId,
		ContestID:  contestId,
		MemberType: newMemberType,
		LeaderType: targetMember.LeaderType,
	}, nil
}

// WithdrawFromContest - 대회 탈퇴 (멤버 본인만 가능, 리더는 불가)
func (s *ContestApplicationService) WithdrawFromContest(contestId, userId int64) error {
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
