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
	teamDBRepository     port.TeamDatabasePort
	teamRedisRepo        port.TeamRedisPort
	contestRepository    contestPort.ContestDatabasePort
	oauth2Repository     oauth2Port.OAuth2DatabasePort
	userQueryRepo        userQueryPort.UserQueryPort
	eventPublisher       port.TeamEventPublisherPort
	persistencePublisher port.TeamPersistencePublisherPort
	notificationHandler  notificationPort.NotificationHandlerPort
}

func NewTeamService(
	teamDBRepository port.TeamDatabasePort,
	teamRedisRepo port.TeamRedisPort,
	contestRepository contestPort.ContestDatabasePort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
	eventPublisher port.TeamEventPublisherPort,
	persistencePublisher port.TeamPersistencePublisherPort,
) *TeamService {
	return &TeamService{
		teamDBRepository:     teamDBRepository,
		teamRedisRepo:        teamRedisRepo,
		contestRepository:    contestRepository,
		oauth2Repository:     oauth2Repository,
		userQueryRepo:        userQueryRepo,
		eventPublisher:       eventPublisher,
		persistencePublisher: persistencePublisher,
	}
}

// SetNotificationHandler sets the notification handler (to avoid circular dependency)
func (s *TeamService) SetNotificationHandler(handler notificationPort.NotificationHandlerPort) {
	s.notificationHandler = handler
}

// SetContestRepository sets the contest repository (to avoid circular dependency)
func (s *TeamService) SetContestRepository(repository contestPort.ContestDatabasePort) {
	s.contestRepository = repository
}

// CreateTeamInCache creates a new team in Redis cache with the creator as leader
func (s *TeamService) CreateTeamInCache(ctx context.Context, contestID, leaderUserID int64, teamName *string) (*port.CachedTeam, error) {
	// Get contest for max members and Discord channel info
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	// Check if contest is in a valid state for team creation
	if !contest.IsActive() && !contest.IsPending() {
		return nil, exception.ErrContestNotActive
	}

	// Check if user already has a team in this contest
	existingTeam, _ := s.teamRedisRepo.GetTeam(ctx, contestID)
	if existingTeam != nil {
		// Check if user is already in this team
		isMember, _ := s.teamRedisRepo.IsMember(ctx, contestID, leaderUserID)
		if isMember {
			return nil, exception.ErrTeamMemberAlreadyExists
		}
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

	// Get next team_id for this contest
	teamID, err := s.teamDBRepository.GetNextTeamID(contestID)
	if err != nil {
		return nil, err
	}

	maxMembers := contest.TotalTeamMember

	// Create cached team
	cachedTeam := &port.CachedTeam{
		ContestID:    contestID,
		TeamID:       teamID,
		TeamName:     teamName,
		MaxMembers:   maxMembers,
		CurrentCount: 1,
		LeaderUserID: leaderUserID,
		CreatedAt:    time.Now(),
		IsFinalized:  false,
	}

	// Create leader member
	leader := &port.CachedTeamMember{
		UserID:     leaderUserID,
		ContestID:  contestID,
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

	// If single player team, auto-finalize
	if maxMembers == 1 {
		if err := s.FinalizeTeam(ctx, contestID, leaderUserID); err != nil {
			return nil, err
		}
		cachedTeam.IsFinalized = true
	}

	// Publish member joined event if contest has Discord integration
	if contest.HasDiscordIntegration() {
		go s.publishMemberJoinedEventForContest(ctx, contest, leader, 1, maxMembers)
	}

	// Publish team created for persistence (Write-Behind)
	go s.publishTeamCreatedForPersistence(ctx, cachedTeam, leader)

	return cachedTeam, nil
}

// GetTeam returns team information from Redis cache
func (s *TeamService) GetTeam(ctx context.Context, contestID int64) (*dto.TeamResponse, error) {
	// Check if team is finalized (in DB)
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return s.getTeamFromDB(contestID)
	}

	// Get from cache
	cachedTeam, err := s.teamRedisRepo.GetTeam(ctx, contestID)
	if err != nil {
		// Fallback to DB
		return s.getTeamFromDB(contestID)
	}

	members, err := s.teamRedisRepo.GetAllMembers(ctx, contestID)
	if err != nil {
		return nil, err
	}

	return dto.ToCachedTeamResponseForContest(cachedTeam, members), nil
}

func (s *TeamService) getTeamFromDB(contestID int64) (*dto.TeamResponse, error) {
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	teams, err := s.teamDBRepository.GetTeamsByContestWithMembers(contestID)
	if err != nil {
		return nil, err
	}

	if len(teams) == 0 {
		return nil, exception.ErrTeamMemberNotFound
	}

	// Return the first team (one team per user per contest)
	teamWithMembers := teams[0]
	return dto.ToTeamResponseForContest(contest, teamWithMembers.Team, teamWithMembers.Members), nil
}

// InviteMember sends an invitation to a user to join the team
func (s *TeamService) InviteMember(ctx context.Context, contestID, inviterUserID, inviteeUserID int64) (*port.TeamInvite, error) {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return nil, exception.ErrTeamAlreadyFinalized
	}

	// Get contest info
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	if !contest.IsActive() && !contest.IsPending() {
		return nil, exception.ErrContestNotActive
	}

	// Check if inviter is a member
	isMember, err := s.teamRedisRepo.IsMember(ctx, contestID, inviterUserID)
	if err != nil || !isMember {
		return nil, exception.ErrNotTeamMember
	}

	// Check if invitee is already a member
	isInviteeMember, _ := s.teamRedisRepo.IsMember(ctx, contestID, inviteeUserID)
	if isInviteeMember {
		return nil, exception.ErrTeamMemberAlreadyExists
	}

	// Check if invitee already has a pending invite
	hasPending, _ := s.teamRedisRepo.HasPendingInvite(ctx, contestID, inviteeUserID)
	if hasPending {
		return nil, exception.ErrTeamMemberAlreadyExists
	}

	// Check team capacity
	memberCount, err := s.teamRedisRepo.GetMemberCount(ctx, contestID)
	if err != nil {
		return nil, err
	}

	maxMembers := contest.TotalTeamMember
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
		ContestID:   contestID,
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
	if contest.HasDiscordIntegration() {
		go s.publishInviteEventForContest(ctx, contest, inviterUserID, inviterDiscordID, inviter.Username, inviteeUserID, inviteeDiscordID, invitee.Username)
	}

	// Get team name for notification
	cachedTeam, _ := s.teamRedisRepo.GetTeam(ctx, contestID)
	teamName := ""
	if cachedTeam != nil && cachedTeam.TeamName != nil {
		teamName = *cachedTeam.TeamName
	}

	// Send SSE notification to invitee
	go s.sendTeamInviteReceivedNotification(inviteeUserID, inviter.Username, teamName, 0, contestID)

	return invite, nil
}

// AcceptInvite accepts a team invitation
func (s *TeamService) AcceptInvite(ctx context.Context, contestID, inviteeUserID int64) (*port.CachedTeamMember, error) {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return nil, exception.ErrTeamAlreadyFinalized
	}

	// Get contest info
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return nil, err
	}

	// Get team info to get teamID
	cachedTeam, err := s.teamRedisRepo.GetTeam(ctx, contestID)
	if err != nil {
		return nil, err
	}

	// Check team capacity before accepting
	memberCount, err := s.teamRedisRepo.GetMemberCount(ctx, contestID)
	if err != nil {
		return nil, err
	}

	maxMembers := contest.TotalTeamMember
	if memberCount >= maxMembers {
		return nil, exception.ErrTeamIsFull
	}

	// Accept the invite
	if err := s.teamRedisRepo.AcceptInvite(ctx, contestID, inviteeUserID); err != nil {
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
		ContestID:  contestID,
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
	_ = s.teamRedisRepo.CancelInvite(ctx, contestID, inviteeUserID)

	// Publish member joined event
	if contest.HasDiscordIntegration() {
		newCount := memberCount + 1
		go s.publishMemberJoinedEventForContest(ctx, contest, member, newCount, maxMembers)
	}

	// Send SSE notification to inviter (the leader or whoever invited)
	teamName := ""
	if cachedTeam.TeamName != nil {
		teamName = *cachedTeam.TeamName
	}
	go s.sendTeamInviteAcceptedNotification(cachedTeam.LeaderUserID, invitee.Username, teamName, 0, contestID)

	// Publish member added for persistence (Write-Behind)
	go s.publishMemberAddedForPersistence(ctx, cachedTeam, member)

	return member, nil
}

// RejectInvite rejects a team invitation
func (s *TeamService) RejectInvite(ctx context.Context, contestID, inviteeUserID int64) error {
	cachedTeam, _ := s.teamRedisRepo.GetTeam(ctx, contestID)
	invitee, _ := s.userQueryRepo.FindById(inviteeUserID)

	// Reject the invite
	if err := s.teamRedisRepo.RejectInvite(ctx, contestID, inviteeUserID); err != nil {
		return err
	}

	// Send SSE notification to leader
	if cachedTeam != nil && invitee != nil {
		teamName := ""
		if cachedTeam.TeamName != nil {
			teamName = *cachedTeam.TeamName
		}
		go s.sendTeamInviteRejectedNotification(cachedTeam.LeaderUserID, invitee.Username, teamName, 0, contestID)
	}

	return nil
}

// KickMember removes a member from the team (Leader only)
func (s *TeamService) KickMember(ctx context.Context, contestID, kickerUserID, targetUserID int64) error {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get contest
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return err
	}

	if !contest.IsActive() && !contest.IsPending() {
		return exception.ErrContestNotActive
	}

	// Get kicker and verify they're the leader
	kicker, err := s.teamRedisRepo.GetMember(ctx, contestID, kickerUserID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if kicker.MemberType != port.TeamMemberTypeLeader {
		return exception.ErrNoPermissionToKick
	}

	// Get target
	target, err := s.teamRedisRepo.GetMember(ctx, contestID, targetUserID)
	if err != nil {
		return exception.ErrTeamMemberNotFound
	}

	// Cannot kick leader
	if target.MemberType == port.TeamMemberTypeLeader {
		return exception.ErrCannotKickLeader
	}

	if err := s.teamRedisRepo.RemoveMember(ctx, contestID, targetUserID); err != nil {
		return err
	}

	return nil
}

// LeaveTeam allows a member to leave the team voluntarily
func (s *TeamService) LeaveTeam(ctx context.Context, contestID, userID int64) error {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get contest
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return err
	}

	if !contest.IsActive() && !contest.IsPending() {
		return exception.ErrContestNotActive
	}

	// Get member
	member, err := s.teamRedisRepo.GetMember(ctx, contestID, userID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	// Leader cannot leave
	if member.MemberType == port.TeamMemberTypeLeader {
		return exception.ErrCannotLeaveAsLeader
	}

	return s.teamRedisRepo.RemoveMember(ctx, contestID, userID)
}

// TransferLeadership transfers leadership to another member (Leader only)
func (s *TeamService) TransferLeadership(ctx context.Context, contestID, currentLeaderUserID, newLeaderUserID int64) error {
	// Check if team is finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get contest
	contest, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		return err
	}

	if !contest.IsActive() && !contest.IsPending() {
		return exception.ErrContestNotActive
	}

	// Verify current leader
	currentLeader, err := s.teamRedisRepo.GetMember(ctx, contestID, currentLeaderUserID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if currentLeader.MemberType != port.TeamMemberTypeLeader {
		return exception.ErrNoPermissionToKick
	}

	// Verify new leader is a member
	_, err = s.teamRedisRepo.GetMember(ctx, contestID, newLeaderUserID)
	if err != nil {
		return exception.ErrTeamMemberNotFound
	}

	if err := s.teamRedisRepo.TransferLeadership(ctx, contestID, currentLeaderUserID, newLeaderUserID); err != nil {
		return err
	}

	return nil
}

// FinalizeTeam moves team data from Redis to config when team is complete
// Uses Write-Behind pattern: marks as finalized in Redis and publishes event for async DB persistence
func (s *TeamService) FinalizeTeam(ctx context.Context, contestID, userID int64) error {
	// Verify user is the leader
	leader, err := s.teamRedisRepo.GetLeader(ctx, contestID)
	if err != nil {
		return err
	}

	if leader.UserID != userID {
		return exception.ErrNoPermissionToDelete
	}

	// Check if already finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		return exception.ErrTeamAlreadyFinalized
	}

	// Get team info
	cachedTeam, err := s.teamRedisRepo.GetTeam(ctx, contestID)
	if err != nil {
		return err
	}

	// Check if team has reached max members
	if cachedTeam.CurrentCount < cachedTeam.MaxMembers {
		return exception.ErrTeamNotReady
	}

	// Get all members
	members, err := s.teamRedisRepo.GetAllMembers(ctx, contestID)
	if err != nil {
		return err
	}

	// Mark as finalized in Redis first
	if err := s.teamRedisRepo.MarkAsFinalized(ctx, contestID); err != nil {
		return err
	}

	// Publish team finalized for persistence (Write-Behind)
	// DB persistence happens asynchronously via RabbitMQ consumer
	go s.publishTeamFinalizedForPersistence(ctx, cachedTeam, members)

	// Publish finalized event for Discord notification
	contest, _ := s.contestRepository.GetContestById(contestID)
	if contest != nil && contest.HasDiscordIntegration() {
		memberUserIDs := make([]int64, len(members))
		for i, m := range members {
			memberUserIDs[i] = m.UserID
		}
		go s.publishTeamFinalizedEventForContest(ctx, contest, leader, len(members), memberUserIDs)
	}

	// Increment finalized team count and check if all teams are ready
	finalizedCount, err := s.teamRedisRepo.IncrementFinalizedTeamCount(ctx, contestID)
	if err != nil {
		log.Printf("[TeamService] Failed to increment finalized team count for contest %d: %v", contestID, err)
	} else if contest != nil && int(finalizedCount) == contest.MaxTeamCount {
		go s.publishContestTeamsReadyEvent(ctx, contestID, int(finalizedCount))
	}

	return nil
}

// GetMembers returns all members of a team
func (s *TeamService) GetMembers(ctx context.Context, contestID int64) ([]*port.CachedTeamMember, error) {
	// Check if finalized, fallback to DB
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		teams, err := s.teamDBRepository.GetTeamsByContestWithMembers(contestID)
		if err != nil {
			return nil, err
		}
		if len(teams) == 0 {
			return nil, exception.ErrTeamMemberNotFound
		}

		dbMembers := teams[0].Members
		cachedMembers := make([]*port.CachedTeamMember, len(dbMembers))
		for i, m := range dbMembers {
			memberType := port.TeamMemberTypeMember
			if m.MemberType == domain.TeamMemberTypeLeader {
				memberType = port.TeamMemberTypeLeader
			}
			cachedMembers[i] = &port.CachedTeamMember{
				UserID:     m.UserID,
				ContestID:  contestID,
				TeamID:     m.TeamID,
				MemberType: memberType,
			}
		}
		return cachedMembers, nil
	}

	return s.teamRedisRepo.GetAllMembers(ctx, contestID)
}

// GetMember returns a specific member
func (s *TeamService) GetMember(ctx context.Context, contestID, userID int64) (*port.CachedTeamMember, error) {
	// Check if finalized, fallback to DB
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		team, err := s.teamDBRepository.GetUserTeamInContest(contestID, userID)
		if err != nil {
			return nil, err
		}

		dbMember, err := s.teamDBRepository.GetMemberByTeamAndUser(team.TeamID, userID)
		if err != nil {
			return nil, err
		}

		memberType := port.TeamMemberTypeMember
		if dbMember.MemberType == domain.TeamMemberTypeLeader {
			memberType = port.TeamMemberTypeLeader
		}
		return &port.CachedTeamMember{
			UserID:     dbMember.UserID,
			ContestID:  contestID,
			TeamID:     dbMember.TeamID,
			MemberType: memberType,
		}, nil
	}

	return s.teamRedisRepo.GetMember(ctx, contestID, userID)
}

// DeleteTeam deletes the entire team (Leader only)
func (s *TeamService) DeleteTeam(ctx context.Context, contestID, userID int64) error {
	// Check if finalized
	isFinalized, _ := s.teamRedisRepo.IsFinalized(ctx, contestID)
	if isFinalized {
		// Delete from DB
		team, err := s.teamDBRepository.GetUserTeamInContest(contestID, userID)
		if err != nil {
			return exception.ErrNotTeamMember
		}

		member, err := s.teamDBRepository.GetMemberByTeamAndUser(team.TeamID, userID)
		if err != nil {
			return exception.ErrNotTeamMember
		}

		if !member.CanDeleteTeam() {
			return exception.ErrNoPermissionToDelete
		}

		// Delete all members first, then team
		if err := s.teamDBRepository.DeleteAllMembersByTeamID(team.TeamID); err != nil {
			return err
		}
		return s.teamDBRepository.Delete(team.TeamID)
	}

	// Get team info before deletion for persistence event
	cachedTeam, _ := s.teamRedisRepo.GetTeam(ctx, contestID)

	// Delete from Redis
	leader, err := s.teamRedisRepo.GetLeader(ctx, contestID)
	if err != nil {
		return exception.ErrNotTeamMember
	}

	if leader.UserID != userID {
		return exception.ErrNoPermissionToDelete
	}

	if err := s.teamRedisRepo.ClearTeam(ctx, contestID); err != nil {
		return err
	}

	// Publish team deleted for persistence (Write-Behind)
	if cachedTeam != nil {
		go s.publishTeamDeletedForPersistence(ctx, cachedTeam)
	}

	return nil
}

// Event publishing helper methods for contest-based events

func (s *TeamService) publishInviteEventForContest(
	ctx context.Context,
	contest interface {
		HasDiscordIntegration() bool
	},
	inviterUserID int64, inviterDiscordID, inviterUsername string,
	inviteeUserID int64, inviteeDiscordID, inviteeUsername string,
) {
	// Get contest details for Discord info
	contestDetails, err := s.contestRepository.GetContestById(0) // We need contestID here
	if err != nil || contestDetails.DiscordGuildId == nil {
		return
	}

	// Note: This function needs the contest object passed properly
	// For now, we'll type assert to get the contest ID
	type contestWithID interface {
		HasDiscordIntegration() bool
	}

	// We'll update this when we have the full contest object
}

func (s *TeamService) publishMemberJoinedEventForContest(
	ctx context.Context,
	contest interface {
		HasDiscordIntegration() bool
	},
	member *port.CachedTeamMember,
	currentCount, maxMembers int,
) {
	contestDetails, err := s.contestRepository.GetContestById(member.ContestID)
	if err != nil || contestDetails.DiscordGuildId == nil {
		return
	}

	event := &port.TeamMemberEvent{
		EventType:            port.TeamEventTypeMemberJoined,
		Timestamp:            time.Now(),
		ContestID:            member.ContestID,
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

func (s *TeamService) publishTeamFinalizedEventForContest(
	ctx context.Context,
	contest interface {
		HasDiscordIntegration() bool
	},
	leader *port.CachedTeamMember,
	memberCount int,
	memberUserIDs []int64,
) {
	contestDetails, err := s.contestRepository.GetContestById(leader.ContestID)
	if err != nil || contestDetails.DiscordGuildId == nil {
		return
	}

	event := &port.TeamFinalizedEvent{
		EventType:            port.TeamEventTypeTeamFinalized,
		Timestamp:            time.Now(),
		ContestID:            leader.ContestID,
		LeaderUserID:         leader.UserID,
		LeaderDiscordID:      leader.DiscordID,
		DiscordGuildID:       *contestDetails.DiscordGuildId,
		DiscordTextChannelID: *contestDetails.DiscordTextChannelId,
		MemberCount:          memberCount,
		MemberUserIDs:        memberUserIDs,
	}

	_ = s.eventPublisher.PublishTeamFinalizedEvent(ctx, event)
}

// publishContestTeamsReadyEvent publishes an event when all teams in a contest are finalized
func (s *TeamService) publishContestTeamsReadyEvent(ctx context.Context, contestID int64, finalizedCount int) {
	contestDetails, err := s.contestRepository.GetContestById(contestID)
	if err != nil {
		log.Printf("[TeamService] Failed to get contest %d for teams ready event: %v", contestID, err)
		return
	}

	event := &port.ContestTeamsReadyEvent{
		EventType:          port.TeamEventTypeContestTeamsReady,
		Timestamp:          time.Now(),
		ContestID:          contestID,
		FinalizedTeamCount: finalizedCount,
		MaxTeamCount:       contestDetails.MaxTeamCount,
	}

	if contestDetails.DiscordGuildId != nil {
		event.DiscordGuildID = *contestDetails.DiscordGuildId
	}
	if contestDetails.DiscordTextChannelId != nil {
		event.DiscordTextChannelID = *contestDetails.DiscordTextChannelId
	}

	if err := s.eventPublisher.PublishContestTeamsReadyEvent(ctx, event); err != nil {
		log.Printf("[TeamService] Failed to publish contest teams ready event: %v", err)
	}
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

// Write-Behind Pattern: Persistence event publishing helper methods

// publishTeamCreatedForPersistence publishes event for async DB persistence
func (s *TeamService) publishTeamCreatedForPersistence(ctx context.Context, cachedTeam *port.CachedTeam, leader *port.CachedTeamMember) {
	if s.persistencePublisher == nil {
		return
	}

	leaderMember := &port.TeamMemberPersistence{
		UserID:     leader.UserID,
		MemberType: leader.MemberType,
		JoinedAt:   leader.JoinedAt,
	}

	event := &port.TeamPersistenceEvent{
		TeamID:    cachedTeam.TeamID,
		ContestID: cachedTeam.ContestID,
		TeamName:  cachedTeam.TeamName,
		Members:   []*port.TeamMemberPersistence{leaderMember},
	}

	if err := s.persistencePublisher.PublishTeamCreated(ctx, event); err != nil {
		log.Printf("Failed to publish team created persistence event: %v", err)
	}
}

// publishMemberAddedForPersistence publishes event for async DB persistence
func (s *TeamService) publishMemberAddedForPersistence(ctx context.Context, cachedTeam *port.CachedTeam, member *port.CachedTeamMember) {
	if s.persistencePublisher == nil {
		return
	}

	event := &port.TeamPersistenceEvent{
		TeamID:         cachedTeam.TeamID,
		ContestID:      cachedTeam.ContestID,
		TeamName:       cachedTeam.TeamName,
		MemberUserID:   &member.UserID,
		MemberType:     &member.MemberType,
		MemberJoinedAt: &member.JoinedAt,
	}

	if err := s.persistencePublisher.PublishMemberAdded(ctx, event); err != nil {
		log.Printf("Failed to publish member added persistence event: %v", err)
	}
}

// publishTeamFinalizedForPersistence publishes event for async DB persistence
func (s *TeamService) publishTeamFinalizedForPersistence(ctx context.Context, cachedTeam *port.CachedTeam, members []*port.CachedTeamMember) {
	if s.persistencePublisher == nil {
		return
	}

	persistenceMembers := make([]*port.TeamMemberPersistence, len(members))
	for i, m := range members {
		persistenceMembers[i] = &port.TeamMemberPersistence{
			UserID:     m.UserID,
			MemberType: m.MemberType,
			JoinedAt:   m.JoinedAt,
		}
	}

	event := &port.TeamPersistenceEvent{
		TeamID:    cachedTeam.TeamID,
		ContestID: cachedTeam.ContestID,
		TeamName:  cachedTeam.TeamName,
		Members:   persistenceMembers,
	}

	if err := s.persistencePublisher.PublishTeamFinalized(ctx, event); err != nil {
		log.Printf("Failed to publish team finalized persistence event: %v", err)
	}
}

// publishTeamDeletedForPersistence publishes event for async DB persistence
func (s *TeamService) publishTeamDeletedForPersistence(ctx context.Context, cachedTeam *port.CachedTeam) {
	if s.persistencePublisher == nil {
		return
	}

	event := &port.TeamPersistenceEvent{
		TeamID:    cachedTeam.TeamID,
		ContestID: cachedTeam.ContestID,
		TeamName:  cachedTeam.TeamName,
	}

	if err := s.persistencePublisher.PublishTeamDeleted(ctx, event); err != nil {
		log.Printf("Failed to publish team deleted persistence event: %v", err)
	}
}
