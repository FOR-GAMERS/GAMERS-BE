package application

import (
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	gameDomain "GAMERS-BE/internal/game/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/exception"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	"context"
	"errors"
	"log"
	"time"
)

// TournamentGeneratorPort defines the interface for tournament generation
type TournamentGeneratorPort interface {
	GenerateTournamentBracket(contestID int64, maxTeamCount int, gameTeamType gameDomain.GameTeamType) ([]*gameDomain.Game, error)
}

type ContestService struct {
	repository            port.ContestDatabasePort
	memberRepository      port.ContestMemberDatabasePort
	applicationRepository port.ContestApplicationRedisPort
	oauth2Repository      oauth2Port.OAuth2DatabasePort
	eventPublisher        port.EventPublisherPort
	discordValidator      port.DiscordValidationPort
	tournamentGenerator   TournamentGeneratorPort
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

// NewContestServiceWithDiscord creates a new contest service with Discord validation
func NewContestServiceWithDiscord(
	repository port.ContestDatabasePort,
	memberRepository port.ContestMemberDatabasePort,
	applicationRepository port.ContestApplicationRedisPort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	eventPublisher port.EventPublisherPort,
	discordValidator port.DiscordValidationPort,
) *ContestService {
	return &ContestService{
		repository:            repository,
		memberRepository:      memberRepository,
		applicationRepository: applicationRepository,
		oauth2Repository:      oauth2Repository,
		eventPublisher:        eventPublisher,
		discordValidator:      discordValidator,
	}
}

// NewContestServiceFull creates a new contest service with all dependencies
func NewContestServiceFull(
	repository port.ContestDatabasePort,
	memberRepository port.ContestMemberDatabasePort,
	applicationRepository port.ContestApplicationRedisPort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	eventPublisher port.EventPublisherPort,
	discordValidator port.DiscordValidationPort,
	tournamentGenerator TournamentGeneratorPort,
) *ContestService {
	return &ContestService{
		repository:            repository,
		memberRepository:      memberRepository,
		applicationRepository: applicationRepository,
		oauth2Repository:      oauth2Repository,
		eventPublisher:        eventPublisher,
		discordValidator:      discordValidator,
		tournamentGenerator:   tournamentGenerator,
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

	// Validate Discord integration if Discord fields are provided
	if req.DiscordGuildId != nil && *req.DiscordGuildId != "" {
		if err := c.validateDiscordIntegration(req, discordAccount.DiscordId); err != nil {
			return nil, nil, err
		}
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
		req.GameType,
		req.GamePointTableId,
		req.TotalTeamMember,
		req.DiscordGuildId,
		req.DiscordTextChannelId,
		req.Thumbnail,
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

	// Generate tournament bracket for TOURNAMENT type contests
	if savedContest.ContestType == domain.ContestTypeTournament && c.tournamentGenerator != nil {
		if err := c.generateTournamentBracket(savedContest); err != nil {
			// Log error but don't fail contest creation
			log.Printf("Failed to generate tournament bracket for contest %d: %v", savedContest.ContestID, err)
		}
	}

	// Publish contest created event (async - failure doesn't affect contest creation)
	if savedContest.HasDiscordIntegration() {
		go c.publishContestCreatedEvent(context.Background(), savedContest, userId, discordAccount.DiscordId)
	}

	return savedContest, nil, nil
}

// validateDiscordIntegration validates the Discord guild and channel
func (c *ContestService) validateDiscordIntegration(req *dto.CreateContestRequest, userDiscordID string) error {
	if c.discordValidator == nil {
		// Discord validation is optional
		return nil
	}

	// Validate channel is required when guild is specified
	if req.DiscordTextChannelId == nil || *req.DiscordTextChannelId == "" {
		return exception.ErrDiscordChannelRequired
	}

	// Validate that bot and user are in the guild, and channel is valid
	return c.discordValidator.ValidateGuildForContest(
		*req.DiscordGuildId,
		*req.DiscordTextChannelId,
		userDiscordID,
	)
}

// generateTournamentBracket generates tournament games for a contest
func (c *ContestService) generateTournamentBracket(contest *domain.Contest) error {
	if contest.MaxTeamCount <= 0 {
		return nil // No teams, no bracket needed
	}

	// Default to HURUPA (5-member) team type if game type is Valorant
	gameTeamType := gameDomain.GameTeamTypeHurupa
	if contest.GameType != nil && *contest.GameType == gameDomain.GameTypeLOL {
		gameTeamType = gameDomain.GameTeamTypeHurupa // LOL also uses 5 members
	}

	_, err := c.tournamentGenerator.GenerateTournamentBracket(
		contest.ContestID,
		contest.MaxTeamCount,
		gameTeamType,
	)
	return err
}

func (c *ContestService) GetContestById(id int64) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(id)

	if err != nil {
		return nil, err
	}

	return contest, nil
}

func (c *ContestService) GetAllContests(offset, limit int, sortReq *commonDto.SortRequest, title *string) ([]domain.Contest, int64, error) {
	contests, totalCount, err := c.repository.GetContests(offset, limit, sortReq, title)

	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}

func (c *ContestService) GetMyContests(userId int64, pagination *commonDto.PaginationRequest, sortReq *commonDto.SortRequest, status *domain.ContestStatus) ([]*port.ContestWithMembership, int64, error) {
	contests, totalCount, err := c.memberRepository.GetContestsByUserId(userId, pagination, sortReq, status)

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

func (c *ContestService) StopContest(ctx context.Context, contestId, userId int64) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(contestId)
	if err != nil {
		return nil, err
	}

	if err := c.checkLeaderPermission(contestId, userId); err != nil {
		return nil, err
	}

	if !contest.CanStop() {
		return nil, exception.ErrContestNotActive
	}

	if err := contest.TransitionTo(domain.ContestStatusFinished); err != nil {
		return nil, err
	}

	contest.EndedAt = time.Now()

	if err := c.repository.UpdateContest(contest); err != nil {
		return nil, err
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

// GetDiscordGuilds returns all guilds the bot is in
func (c *ContestService) GetDiscordGuilds() ([]port.DiscordGuild, error) {
	if c.discordValidator == nil {
		return nil, exception.ErrDiscordAPIError
	}
	return c.discordValidator.GetBotGuilds()
}

// GetDiscordTextChannels returns all text channels in a guild
func (c *ContestService) GetDiscordTextChannels(guildID string) ([]port.DiscordChannel, error) {
	if c.discordValidator == nil {
		return nil, exception.ErrDiscordAPIError
	}
	return c.discordValidator.GetGuildTextChannels(guildID)
}
