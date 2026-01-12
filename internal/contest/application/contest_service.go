package application

import (
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/exception"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	"context"
	"errors"
	"log"
	"time"
)

type ContestService struct {
	repository            port.ContestDatabasePort
	memberRepository      port.ContestMemberDatabasePort
	applicationRepository port.ContestApplicationRedisPort
	oauth2Repository      oauth2Port.OAuth2DatabasePort
	eventPublisher        port.EventPublisherPort
}

func NewContestService(
	repository port.ContestDatabasePort,
	memberRepository port.ContestMemberDatabasePort,
	applicationRepository port.ContestApplicationRedisPort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	eventPublisher port.EventPublisherPort,
) *ContestService {
	return &ContestService{
		repository:            repository,
		memberRepository:      memberRepository,
		applicationRepository: applicationRepository,
		oauth2Repository:      oauth2Repository,
		eventPublisher:        eventPublisher,
	}
}

func (c *ContestService) SaveContest(req *dto.CreateContestRequest, userId int64) (*domain.Contest, *dto.DiscordLinkRequiredResponse, error) {
	// Check if user has linked Discord account
	discordAccount, err := c.oauth2Repository.FindDiscordAccountByUserId(userId)
	if err != nil {
		if errors.Is(err, exception.ErrDiscordUserCannotFound) {
			return nil, dto.NewDiscordLinkRequiredResponse("Discord account linking is required to create a contest"), exception.ErrDiscordLinkRequired
		}
		return nil, nil, err
	}

	contest := *domain.NewContestInstance(
		req.Title,
		req.Description,
		req.MaxTeamCount,
		req.TotalPoint,
		req.ContestType,
		req.StartedAt,
		req.EndedAt,
		req.AutoStart,
		req.DiscordGuildId,
		req.DiscordTextChannelId,
	)

	// Validate contest (including Discord fields)
	if err := contest.Validate(); err != nil {
		return nil, nil, err
	}

	savedContest, err := c.repository.Save(&contest)
	if err != nil {
		return nil, nil, err
	}

	// Save contest creator as leader
	contestMember := domain.NewContestMemberAsLeader(userId, savedContest.ContestID)
	if err := c.memberRepository.Save(contestMember); err != nil {
		// If member save fails, we should consider rolling back the contest creation
		// For now, we'll return the error
		return nil, nil, err
	}

	// Publish contest created event (async - failure doesn't affect contest creation)
	if savedContest.HasDiscordIntegration() {
		go c.publishContestCreatedEvent(context.Background(), savedContest, userId, discordAccount.DiscordId)
	}

	return savedContest, nil, nil
}

func (c *ContestService) GetContestById(id int64) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(id)

	if err != nil {
		return nil, err
	}

	return contest, nil
}

func (c *ContestService) GetAllContests(offset, limit int, sortReq *commonDto.SortRequest) ([]domain.Contest, int64, error) {
	contests, totalCount, err := c.repository.GetContests(offset, limit, sortReq)

	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}

func (c *ContestService) UpdateContest(id int64, req *dto.UpdateContestRequest) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(id)

	if err != nil {
		return nil, err
	}

	if !req.HasChanges() {
		return nil, exception.ErrContestNoChanges
	}

	if err = req.Validate(); err != nil {
		return nil, err
	}

	req.ApplyTo(contest)

	err = c.repository.UpdateContest(contest)

	if err != nil {
		return nil, err
	}

	return contest, nil
}

func (c *ContestService) DeleteContestById(id int64) error {
	return c.repository.DeleteContestById(id)
}

// checkLeaderPermission - Leader 권한 확인
func (c *ContestService) checkLeaderPermission(contestId, userId int64) error {
	member, err := c.memberRepository.GetByContestAndUser(contestId, userId)
	if err != nil {
		return exception.ErrInvalidAccess
	}
	if !member.IsLeader() {
		return exception.ErrContestAlreadyStarted
	}

	return nil
}

func (c *ContestService) StartContest(ctx context.Context, contestId, userId int64) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	if err := c.checkLeaderPermission(contestId, userId); err != nil {
		return nil, err
	}

	if contest.ContestStatus != domain.ContestStatusPending {
		return nil, exception.ErrContestNotPending
	}

	if !contest.CanStart() {
		return nil, exception.ErrContestCannotStart
	}

	acceptedUserIDs, err := c.applicationRepository.GetAcceptedApplications(ctx, contestId)
	if err != nil {
		return nil, err
	}

	if len(acceptedUserIDs) > 0 {
		members := make([]*domain.ContestMember, 0, len(acceptedUserIDs))
		for _, userID := range acceptedUserIDs {
			member := domain.NewContestMember(userID, contestId, domain.MemberTypeNormal, domain.LeaderTypeMember)
			members = append(members, member)
		}

		if err := c.memberRepository.SaveBatch(members); err != nil {
			return nil, err
		}
	}

	if err := contest.TransitionTo(domain.ContestStatusActive); err != nil {
		return nil, err
	}

	if err := c.repository.UpdateContest(contest); err != nil {
		return nil, err
	}

	if err := c.applicationRepository.ClearApplications(ctx, contestId); err != nil {
		log.Fatal(err)
	}

	return contest, nil
}

// publishContestCreatedEvent publishes an event when a new contest is created
func (c *ContestService) publishContestCreatedEvent(
	ctx context.Context,
	contest *domain.Contest,
	creatorUserId int64,
	creatorDiscordId string,
) {
	event := &port.ContestCreatedEvent{
		EventType:            port.EventTypeContestCreated,
		ContestID:            contest.ContestID,
		CreatorUserID:        creatorUserId,
		CreatorDiscordID:     creatorDiscordId,
		DiscordGuildID:       *contest.DiscordGuildId,
		DiscordTextChannelID: *contest.DiscordTextChannelId,
		ContestTitle:         contest.Title,
		Timestamp:            time.Now(),
		Data: map[string]interface{}{
			"contest_type":   contest.ContestType,
			"max_team_count": contest.MaxTeamCount,
			"started_at":     contest.StartedAt,
			"ended_at":       contest.EndedAt,
			"auto_start":     contest.AutoStart,
		},
	}

	if err := c.eventPublisher.PublishContestCreatedEvent(ctx, event); err != nil {
		// Log error but don't affect contest creation
		_ = err
	}
}
