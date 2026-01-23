package application

import (
	contestPort "GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
	notificationPort "GAMERS-BE/internal/notification/application/port"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"
	"context"
	"log"
	"time"
)

const (
	DefaultTeamTTL   = 7 * 24 * time.Hour // 7 days
	DefaultInviteTTL = 24 * time.Hour     // 24 hours
)

type TeamService struct {
	gameRepository      port.GameDatabasePort
	teamDBRepository    port.TeamDatabasePort
	teamRedisRepo       port.TeamRedisPort
	contestRepository   contestPort.ContestDatabasePort
	oauth2Repository    oauth2Port.OAuth2DatabasePort
	userQueryRepo       userQueryPort.UserQueryPort
	eventPublisher      port.TeamEventPublisherPort
	notificationHandler notificationPort.NotificationHandlerPort
}

func NewTeamService(
	gameRepository port.GameDatabasePort,
	teamDBRepository port.TeamDatabasePort,
	teamRedisRepo port.TeamRedisPort,
	contestRepository contestPort.ContestDatabasePort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
	eventPublisher port.TeamEventPublisherPort,
) *TeamService {
	return &TeamService{
		gameRepository:    gameRepository,
		teamDBRepository:  teamDBRepository,
		teamRedisRepo:     teamRedisRepo,
		contestRepository: contestRepository,
		oauth2Repository:  oauth2Repository,
		userQueryRepo:     userQueryRepo,
		eventPublisher:    eventPublisher,
	}
}

// SetNotificationHandler sets the notification handler (to avoid circular dependency)
func (s *TeamService) SetNotificationHandler(handler notificationPort.NotificationHandlerPort) {
	s.notificationHandler = handler
}

// CreateTeamInCache creates a new team in Redis cache with the creator as leader
func (s *TeamService) CreateTeamInCache(ctx context.Context, gameID, leaderUserID int64, teamName *string) (*port.CachedTeam, error) {
	// Get game to get contest info and max members
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	// Get contest for Discord channel info
	contest, err := s.contestRepository.GetContestById(game.ContestID)
	if err != nil {
		return nil, err
	}

	// Get leader's user info
	user, err := s.userQueryRepo.FindById(leaderUserID)
	if err != nil {
		return nil, err
	}

	// Get leader's Discord info if available
	var discordID string
	discordAccount, err := s.oauth2Repository.FindDiscordAccountByUserId(leaderUserID)
	if err == nil && discordAccount != nil {
		discordID = discordAccount.DiscordId
	}

	// Get next team_id for this game
	teamID, err := s.teamDBRepository.GetNextTeamID(gameID)
	if err != nil {
		return nil, err
	}

	maxMembers := game.GameTeamType.GetMaxTeamMembers()

	// Create cached team
	cachedTeam := &port.CachedTeam{
		GameID:       gameID,
		TeamID:       teamID,
		TeamName:     teamName,
		ContestID:    game.ContestID,
		MaxMembers:   maxMembers,
		CurrentCount: 1,
		LeaderUserID: leaderUserID,
		CreatedAt:    time.Now(),
		IsFinalized:  false,
	}

	// Create leader member
	leader := &port.CachedTeamMember{
		UserID:     leaderUserID,
		GameID:     gameID,
		TeamID:     teamID,
		MemberType: port.TeamMemberTypeLeader,
		JoinedAt:   time.Now(),
		DiscordID:  discordID,
		Username:   user.Username,
		Tag:        user.Tag,
	}

	// Store in Redis
	if err := s.teamRedisRepo.CreateTeam(ctx, cachedTeam, leader, DefaultTeamTTL); err != nil {
		return nil, err
	}

	// If single player game, auto-finalize
	if maxMembers == 1 {
		if err := s.FinalizeTeam(ctx, gameID, leaderUserID); err != nil {
			return nil, err
		}
		cachedTeam.IsFinalized = true
	}

	// Publish member joined event if contest has Discord integration
	if contest.HasDiscordIntegration() {
		go s.publishMemberJoinedEvent(ctx, game, contest, leader, 1, maxMembers)
	}

	return cachedTeam, nil
}

// GetTeam returns team information from Redis cache
func (s *TeamService) GetTeam(ctx context.Context, gameID int64) (*dto.TeamResponse, error) {
	// Check if team is finalized (in DB)
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return s.getTeamFromDB(gameID)
	}

	// Get from cache
	cachedTeam, err := s.teamRedisRepo.GetTeam(ctx, gameID)
	if err != nil {
		// Fallback to DB
		return s.getTeamFromDB(gameID)
	}

	members, err := s.teamRedisRepo.GetAllMembers(ctx, gameID)
	if err != nil {
		return nil, err
	}

	return dto.ToCachedTeamResponse(cachedTeam, members), nil
}

func (s *TeamService) getTeamFromDB(gameID int64) (*dto.TeamResponse, error) {
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	teamWithMembers, err := s.teamDBRepository.GetTeamByGameID(gameID)
	if err != nil {
		return nil, err
	}

	return dto.ToTeamResponse(game, teamWithMembers.Team, teamWithMembers.Members), nil
}

// InviteMember sends an invitation to a user to join the team
func (s *TeamService) InviteMember(ctx context.Context, gameID, inviterUserID, inviteeUserID int64) (*port.TeamInvite, error) {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return nil, exception.ErrTeamAlreadyFinalized
	}

	// Get game info
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	if !game.IsPending() {
		return nil, exception.ErrCannotInviteToGame
	}

	// Check if inviter is a member
	isMember, err := s.teamRedisRepo.IsMember(ctx, gameID, inviterUserID)
	if err != nil || !isMember {
		return nil, exception.ErrNotTeamMember
	}

	// Check if invitee is already a member
	isInviteeMember, _ := s.teamRedisRepo.IsMember(ctx, gameID, inviteeUserID)
	if isInviteeMember {
		return nil, exception.ErrTeamMemberAlreadyExists
	}

	// Check if invitee already has a pending invite
	hasPending, _ := s.teamRedisRepo.HasPendingInvite(ctx, gameID, inviteeUserID)
	if hasPending {
		return nil, exception.ErrTeamMemberAlreadyExists
	}

	// Check team capacity
	memberCount, err := s.teamRedisRepo.GetMemberCount(ctx, gameID)
	if err != nil {
		return nil, err
	}

	maxMembers := game.GameTeamType.GetMaxTeamMembers()
	if memberCount >= maxMembers {
		return nil, exception.ErrTeamIsFull
	}

	// Get inviter's info
	inviter, err := s.userQueryRepo.FindById(inviterUserID)
	if err != nil {
		return nil, err
	}

	// Get invitee's info
	invitee, err := s.userQueryRepo.FindById(inviteeUserID)
	if err != nil {
		return nil, err
	}

	// Get Discord IDs
	var inviteeDiscordID string
	inviteeDiscord, err := s.oauth2Repository.FindDiscordAccountByUserId(inviteeUserID)
	if err == nil && inviteeDiscord != nil {
		inviteeDiscordID = inviteeDiscord.DiscordId
	}

	var inviterDiscordID string
	inviterDiscord, err := s.oauth2Repository.FindDiscordAccountByUserId(inviterUserID)
	if err == nil && inviterDiscord != nil {
		inviterDiscordID = inviterDiscord.DiscordId
	}

	// Create invite
	invite := &port.TeamInvite{
		GameID:      gameID,
		InviterID:   inviterUserID,
		InviteeID:   inviteeUserID,
		Status:      port.InviteStatusPending,
		InvitedAt:   time.Now(),
		InviterName: inviter.Username,
		InviteeName: invitee.Username,
		DiscordID:   inviteeDiscordID,
	}

	if err := s.teamRedisRepo.CreateInvite(ctx, invite, DefaultInviteTTL); err != nil {
		return nil, err
	}

	// Publish invite event via RabbitMQ
	contest, err := s.contestRepository.GetContestById(game.ContestID)
	if err == nil && contest.HasDiscordIntegration() {
		go s.publishInviteEvent(ctx, game, contest, inviterUserID, inviterDiscordID, inviter.Username, inviteeUserID, inviteeDiscordID, invitee.Username)
	}

	// Get team name for notification
	cachedTeam, _ := s.teamRedisRepo.GetTeam(ctx, gameID)
	teamName := ""
	if cachedTeam != nil && cachedTeam.TeamName != nil {
		teamName = *cachedTeam.TeamName
	}

	// Send SSE notification to invitee
	go s.sendTeamInviteReceivedNotification(inviteeUserID, inviter.Username, teamName, gameID, game.ContestID)

	return invite, nil
}

// AcceptInvite accepts a team invitation
func (s *TeamService) AcceptInvite(ctx context.Context, gameID, inviteeUserID int64) (*port.CachedTeamMember, error) {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return nil, exception.ErrTeamAlreadyFinalized
	}

	// Get game info
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	// Get team info to get teamID
	cachedTeam, err := s.teamRedisRepo.GetTeam(ctx, gameID)
	if err != nil {
		return nil, err
	}

	// Check team capacity before accepting
	memberCount, err := s.teamRedisRepo.GetMemberCount(ctx, gameID)
	if err != nil {
		return nil, err
	}

	maxMembers := game.GameTeamType.GetMaxTeamMembers()
	if memberCount >= maxMembers {
		return nil, exception.ErrTeamIsFull
	}

	// Accept the invite
	if err := s.teamRedisRepo.AcceptInvite(ctx, gameID, inviteeUserID); err != nil {
		return nil, err
	}

	// Get invitee's info
	invitee, err := s.userQueryRepo.FindById(inviteeUserID)
	if err != nil {
		return nil, err
	}

	// Get Discord ID
	var discordID string
	discordAccount, err := s.oauth2Repository.FindDiscordAccountByUserId(inviteeUserID)
	if err == nil && discordAccount != nil {
		discordID = discordAccount.DiscordId
	}

	// Add member to team
	member := &port.CachedTeamMember{
		UserID:     inviteeUserID,
		GameID:     gameID,
		TeamID:     cachedTeam.TeamID,
		MemberType: port.TeamMemberTypeMember,
		JoinedAt:   time.Now(),
		DiscordID:  discordID,
		Username:   invitee.Username,
		Tag:        invitee.Tag,
	}

	if err := s.teamRedisRepo.AddMember(ctx, member, DefaultTeamTTL); err != nil {
		return nil, err
	}

	// Remove invite from pending
	_ = s.teamRedisRepo.CancelInvite(ctx, gameID, inviteeUserID)

	// Publish member joined event
	contest, _ := s.contestRepository.GetContestById(game.ContestID)
	if contest != nil && contest.HasDiscordIntegration() {
		newCount := memberCount + 1
		go s.publishMemberJoinedEvent(ctx, game, contest, member, newCount, maxMembers)
	}

	// Send SSE notification to inviter (the leader or whoever invited)
	teamName := ""
	if cachedTeam.TeamName != nil {
		teamName = *cachedTeam.TeamName
	}
	go s.sendTeamInviteAcceptedNotification(cachedTeam.LeaderUserID, invitee.Username, teamName, gameID, game.ContestID)

	return member, nil
}

// RejectInvite rejects a team invitation
func (s *TeamService) RejectInvite(ctx context.Context, gameID, inviteeUserID int64) error {
	// Get game and team info before rejecting
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return s.teamRedisRepo.RejectInvite(ctx, gameID, inviteeUserID)
	}

	cachedTeam, _ := s.teamRedisRepo.GetTeam(ctx, gameID)
	invitee, _ := s.userQueryRepo.FindById(inviteeUserID)

	// Reject the invite
	if err := s.teamRedisRepo.RejectInvite(ctx, gameID, inviteeUserID); err != nil {
		return err
	}

	// Send SSE notification to leader
	if cachedTeam != nil && invitee != nil {
		teamName := ""
		if cachedTeam.TeamName != nil {
			teamName = *cachedTeam.TeamName
		}
		go s.sendTeamInviteRejectedNotification(cachedTeam.LeaderUserID, invitee.Username, teamName, gameID, game.ContestID)
	}

	return nil
}

// KickMember removes a member from the team (Leader only)
func (s *TeamService) KickMember(ctx context.Context, gameID, kickerUserID, targetUserID int64) error {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get game
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return err
	}

	if !game.IsPending() {
		return exception.ErrGameNotPending
	}

	// Get kicker and verify they're the leader
	kicker, err := s.teamRedisRepo.GetMember(ctx, gameID, kickerUserID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if kicker.MemberType != port.TeamMemberTypeLeader {
		return exception.ErrNoPermissionToKick
	}

	// Get target
	target, err := s.teamRedisRepo.GetMember(ctx, gameID, targetUserID)
	if err != nil {
		return exception.ErrTeamMemberNotFound
	}

	// Cannot kick leader
	if target.MemberType == port.TeamMemberTypeLeader {
		return exception.ErrCannotKickLeader
	}

	if err := s.teamRedisRepo.RemoveMember(ctx, gameID, targetUserID); err != nil {
		return err
	}

	return nil
}

// LeaveTeam allows a member to leave the team voluntarily
func (s *TeamService) LeaveTeam(ctx context.Context, gameID, userID int64) error {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get game
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return err
	}

	if !game.IsPending() {
		return exception.ErrGameNotPending
	}

	// Get member
	member, err := s.teamRedisRepo.GetMember(ctx, gameID, userID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	// Leader cannot leave
	if member.MemberType == port.TeamMemberTypeLeader {
		return exception.ErrCannotLeaveAsLeader
	}

	return s.teamRedisRepo.RemoveMember(ctx, gameID, userID)
}

// TransferLeadership transfers leadership to another member (Leader only)
func (s *TeamService) TransferLeadership(ctx context.Context, gameID, currentLeaderUserID, newLeaderUserID int64) error {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get game
	game, err := s.gameRepository.GetByID(gameID)
	if err != nil {
		return err
	}

	if !game.IsPending() {
		return exception.ErrGameNotPending
	}

	// Verify current leader
	currentLeader, err := s.teamRedisRepo.GetMember(ctx, gameID, currentLeaderUserID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if currentLeader.MemberType != port.TeamMemberTypeLeader {
		return exception.ErrNoPermissionToKick
	}

	// Verify new leader is a member
	_, err = s.teamRedisRepo.GetMember(ctx, gameID, newLeaderUserID)
	if err != nil {
		return exception.ErrTeamMemberNotFound
	}

	if err := s.teamRedisRepo.TransferLeadership(ctx, gameID, currentLeaderUserID, newLeaderUserID); err != nil {
		return err
	}

	return nil
}

// FinalizeTeam moves team data from Redis to database when team is complete
func (s *TeamService) FinalizeTeam(ctx context.Context, gameID, userID int64) error {
	// Verify user is the leader
	leader, err := s.teamRedisRepo.GetLeader(ctx, gameID)
	if err != nil {
		return err
	}

	if leader.UserID != userID {
		return exception.ErrNoPermissionToDelete
	}

	// Check if already finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get team info
	cachedTeam, err := s.teamRedisRepo.GetTeam(ctx, gameID)
	if err != nil {
		return err
	}

	// Check if team has reached max members
	if cachedTeam.CurrentCount < cachedTeam.MaxMembers {
		return exception.ErrTeamNotReady
	}

	// Get all members
	members, err := s.teamRedisRepo.GetAllMembers(ctx, gameID)
	if err != nil {
		return err
	}

	// Persist to database
	// First, create and save the Team
	teamName := ""
	if cachedTeam.TeamName != nil {
		teamName = *cachedTeam.TeamName
	}
	dbTeam := domain.NewTeam(cachedTeam.ContestID, teamName)
	savedTeam, err := s.teamDBRepository.Save(dbTeam)
	if err != nil {
		return err
	}

	// Then save all TeamMembers
	for _, cachedMember := range members {
		memberType := domain.TeamMemberTypeMember
		if cachedMember.MemberType == port.TeamMemberTypeLeader {
			memberType = domain.TeamMemberTypeLeader
		}

		dbMember := domain.NewTeamMember(savedTeam.TeamID, cachedMember.UserID, memberType)
		if _, err := s.teamDBRepository.SaveMember(dbMember); err != nil {
			return err
		}
	}

	// Mark as finalized in Redis
	if err := s.teamRedisRepo.MarkAsFinalized(ctx, gameID); err != nil {
		return err
	}

	// Publish finalized event
	game, _ := s.gameRepository.GetByID(gameID)
	contest, _ := s.contestRepository.GetContestById(cachedTeam.ContestID)
	if game != nil && contest != nil && contest.HasDiscordIntegration() {
		memberUserIDs := make([]int64, len(members))
		for i, m := range members {
			memberUserIDs[i] = m.UserID
		}
		go s.publishTeamFinalizedEvent(ctx, game, contest, leader, len(members), memberUserIDs)
	}

	return nil
}

// GetMembers returns all members of a team
func (s *TeamService) GetMembers(ctx context.Context, gameID int64) ([]*port.CachedTeamMember, error) {
	// Check if finalized, fallback to DB
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		dbMembers, err := s.teamDBRepository.GetMembersByGameID(gameID)
		if err != nil {
			return nil, err
		}
		// Convert to cached format
		cachedMembers := make([]*port.CachedTeamMember, len(dbMembers))
		for i, m := range dbMembers {
			memberType := port.TeamMemberTypeMember
			if m.MemberType == domain.TeamMemberTypeLeader {
				memberType = port.TeamMemberTypeLeader
			}
			cachedMembers[i] = &port.CachedTeamMember{
				UserID:     m.UserID,
				GameID:     gameID,
				TeamID:     m.TeamID,
				MemberType: memberType,
			}
		}
		return cachedMembers, nil
	}

	return s.teamRedisRepo.GetAllMembers(ctx, gameID)
}

// GetMember returns a specific member
func (s *TeamService) GetMember(ctx context.Context, gameID, userID int64) (*port.CachedTeamMember, error) {
	// Check if finalized, fallback to DB
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		dbMember, err := s.teamDBRepository.GetByGameAndUser(gameID, userID)
		if err != nil {
			return nil, err
		}
		memberType := port.TeamMemberTypeMember
		if dbMember.MemberType == domain.TeamMemberTypeLeader {
			memberType = port.TeamMemberTypeLeader
		}
		return &port.CachedTeamMember{
			UserID:     dbMember.UserID,
			GameID:     gameID,
			TeamID:     dbMember.TeamID,
			MemberType: memberType,
		}, nil
	}

	return s.teamRedisRepo.GetMember(ctx, gameID, userID)
}

// GetMembersByTeamID returns all members of a team by game_id and team_id
func (s *TeamService) GetMembersByTeamID(ctx context.Context, gameID, teamID int64) ([]*port.CachedTeamMember, error) {
	// This function only works with finalized teams (in DB)
	dbMembers, err := s.teamDBRepository.GetByGameAndTeamID(gameID, teamID)
	if err != nil {
		return nil, err
	}

	cachedMembers := make([]*port.CachedTeamMember, len(dbMembers))
	for i, m := range dbMembers {
		memberType := port.TeamMemberTypeMember
		if m.MemberType == domain.TeamMemberTypeLeader {
			memberType = port.TeamMemberTypeLeader
		}
		cachedMembers[i] = &port.CachedTeamMember{
			UserID:     m.UserID,
			GameID:     gameID,
			TeamID:     m.TeamID,
			MemberType: memberType,
		}
	}

	return cachedMembers, nil
}

// DeleteTeam deletes the entire team (Leader only)
func (s *TeamService) DeleteTeam(ctx context.Context, gameID, userID int64) error {
	// Check if finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, gameID)
	if isFinalized {
		// Delete from DB
		member, err := s.teamDBRepository.GetByGameAndUser(gameID, userID)
		if err != nil {
			return exception.ErrNotTeamMember
		}
		if !member.CanDeleteTeam() {
			return exception.ErrNoPermissionToDelete
		}
		return s.teamDBRepository.DeleteAllByGameID(gameID)
	}

	// Delete from Redis
	leader, err := s.teamRedisRepo.GetLeader(ctx, gameID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if leader.UserID != userID {
		return exception.ErrNoPermissionToDelete
	}

	return s.teamRedisRepo.ClearTeam(ctx, gameID)
}

// Event publishing helper methods
func (s *TeamService) publishInviteEvent(
	ctx context.Context,
	game *domain.Game,
	contest interface{ HasDiscordIntegration() bool },
	inviterUserID int64, inviterDiscordID, inviterUsername string,
	inviteeUserID int64, inviteeDiscordID, inviteeUsername string,
) {
	// Type assert contest to get Discord fields
	type contestWithDiscord interface {
		HasDiscordIntegration() bool
	}

	// Get contest details for Discord info
	contestDetails, err := s.contestRepository.GetContestById(game.ContestID)
	if err != nil || contestDetails.DiscordGuildId == nil {
		return
	}

	event := &port.TeamInviteEvent{
		EventType:            port.TeamEventTypeInviteSent,
		Timestamp:            time.Now(),
		GameID:               game.GameID,
		ContestID:            game.ContestID,
		InviterUserID:        inviterUserID,
		InviterDiscordID:     inviterDiscordID,
		InviterUsername:      inviterUsername,
		InviteeUserID:        inviteeUserID,
		InviteeDiscordID:     inviteeDiscordID,
		InviteeUsername:      inviteeUsername,
		DiscordGuildID:       *contestDetails.DiscordGuildId,
		DiscordTextChannelID: *contestDetails.DiscordTextChannelId,
	}

	_ = s.eventPublisher.PublishTeamInviteEvent(ctx, event)
}

func (s *TeamService) publishMemberJoinedEvent(
	ctx context.Context,
	game *domain.Game,
	contest interface{ HasDiscordIntegration() bool },
	member *port.CachedTeamMember,
	currentCount, maxMembers int,
) {
	contestDetails, err := s.contestRepository.GetContestById(game.ContestID)
	if err != nil || contestDetails.DiscordGuildId == nil {
		return
	}

	event := &port.TeamMemberEvent{
		EventType:            port.TeamEventTypeMemberJoined,
		Timestamp:            time.Now(),
		GameID:               game.GameID,
		ContestID:            game.ContestID,
		UserID:               member.UserID,
		DiscordUserID:        member.DiscordID,
		Username:             member.Username,
		DiscordGuildID:       *contestDetails.DiscordGuildId,
		DiscordTextChannelID: *contestDetails.DiscordTextChannelId,
		CurrentMemberCount:   currentCount,
		MaxMembers:           maxMembers,
	}

	_ = s.eventPublisher.PublishTeamMemberEvent(ctx, event)
}

func (s *TeamService) publishTeamFinalizedEvent(
	ctx context.Context,
	game *domain.Game,
	contest interface{ HasDiscordIntegration() bool },
	leader *port.CachedTeamMember,
	memberCount int,
	memberUserIDs []int64,
) {
	contestDetails, err := s.contestRepository.GetContestById(game.ContestID)
	if err != nil || contestDetails.DiscordGuildId == nil {
		return
	}

	event := &port.TeamFinalizedEvent{
		EventType:            port.TeamEventTypeTeamFinalized,
		Timestamp:            time.Now(),
		GameID:               game.GameID,
		ContestID:            game.ContestID,
		LeaderUserID:         leader.UserID,
		LeaderDiscordID:      leader.DiscordID,
		DiscordGuildID:       *contestDetails.DiscordGuildId,
		DiscordTextChannelID: *contestDetails.DiscordTextChannelId,
		MemberCount:          memberCount,
		MemberUserIDs:        memberUserIDs,
	}

	_ = s.eventPublisher.PublishTeamFinalizedEvent(ctx, event)
}

// SSE notification helper methods

// sendTeamInviteReceivedNotification sends SSE notification when user receives team invite
func (s *TeamService) sendTeamInviteReceivedNotification(inviteeUserID int64, inviterUsername, teamName string, gameID, contestID int64) {
	if s.notificationHandler == nil {
		return
	}

	if err := s.notificationHandler.HandleTeamInviteReceived(inviteeUserID, inviterUsername, teamName, gameID, contestID); err != nil {
		log.Printf("Failed to send team invite received notification: %v", err)
	}
}

// sendTeamInviteAcceptedNotification sends SSE notification when invite is accepted
func (s *TeamService) sendTeamInviteAcceptedNotification(inviterUserID int64, inviteeUsername, teamName string, gameID, contestID int64) {
	if s.notificationHandler == nil {
		return
	}

	if err := s.notificationHandler.HandleTeamInviteAccepted(inviterUserID, inviteeUsername, teamName, gameID, contestID); err != nil {
		log.Printf("Failed to send team invite accepted notification: %v", err)
	}
}

// sendTeamInviteRejectedNotification sends SSE notification when invite is rejected
func (s *TeamService) sendTeamInviteRejectedNotification(inviterUserID int64, inviteeUsername, teamName string, gameID, contestID int64) {
	if s.notificationHandler == nil {
		return
	}

	if err := s.notificationHandler.HandleTeamInviteRejected(inviterUserID, inviteeUsername, teamName, gameID, contestID); err != nil {
		log.Printf("Failed to send team invite rejected notification: %v", err)
	}
}
