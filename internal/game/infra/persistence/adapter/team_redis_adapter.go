package adapter

import (
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/utils"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TeamRedisAdapter struct {
	client *redis.Client
}

func NewTeamRedisAdapter(client *redis.Client) *TeamRedisAdapter {
	return &TeamRedisAdapter{
		client: client,
	}
}

// CreateTeam creates a new team with its leader in Redis
func (a *TeamRedisAdapter) CreateTeam(ctx context.Context, team *port.CachedTeam, leader *port.CachedTeamMember, ttl time.Duration) error {
	pipe := a.client.Pipeline()

	// Store team metadata
	teamKey := utils.GetTeamKey(team.ContestID)
	teamData, err := json.Marshal(team)
	if err != nil {
		return err
	}
	pipe.Set(ctx, teamKey, teamData, ttl)

	// Store leader member
	memberKey := utils.GetTeamMemberKey(team.ContestID, leader.UserID)
	memberData, err := json.Marshal(leader)
	if err != nil {
		return err
	}
	pipe.Set(ctx, memberKey, memberData, ttl)

	// Add leader to members set
	membersKey := utils.GetTeamMembersKey(team.ContestID)
	pipe.SAdd(ctx, membersKey, leader.UserID)
	pipe.Expire(ctx, membersKey, ttl)

	// Track user's team
	userTeamsKey := utils.GetUserTeamsKey(leader.UserID)
	pipe.SAdd(ctx, userTeamsKey, team.ContestID)
	pipe.Expire(ctx, userTeamsKey, 30*24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

// GetTeam retrieves team metadata from Redis
func (a *TeamRedisAdapter) GetTeam(ctx context.Context, contestID int64) (*port.CachedTeam, error) {
	teamKey := utils.GetTeamKey(contestID)
	data, err := a.client.Get(ctx, teamKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, exception.ErrTeamMemberNotFound
	}
	if err != nil {
		return nil, err
	}

	var team port.CachedTeam
	if err := json.Unmarshal([]byte(data), &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// UpdateTeamCount updates the current member count
func (a *TeamRedisAdapter) UpdateTeamCount(ctx context.Context, contestID int64, count int) error {
	team, err := a.GetTeam(ctx, contestID)
	if err != nil {
		return err
	}

	team.CurrentCount = count
	teamKey := utils.GetTeamKey(contestID)
	teamData, err := json.Marshal(team)
	if err != nil {
		return err
	}

	ttl := a.client.TTL(ctx, teamKey).Val()
	return a.client.Set(ctx, teamKey, teamData, ttl).Err()
}

// DeleteTeam removes team from Redis
func (a *TeamRedisAdapter) DeleteTeam(ctx context.Context, contestID int64) error {
	return a.ClearTeam(ctx, contestID)
}

// AddMember adds a member to the team in Redis
func (a *TeamRedisAdapter) AddMember(ctx context.Context, member *port.CachedTeamMember, ttl time.Duration) error {
	pipe := a.client.Pipeline()

	// Store member data
	memberKey := utils.GetTeamMemberKey(member.ContestID, member.UserID)
	memberData, err := json.Marshal(member)
	if err != nil {
		return err
	}
	pipe.Set(ctx, memberKey, memberData, ttl)

	// Add to members set
	membersKey := utils.GetTeamMembersKey(member.ContestID)
	pipe.SAdd(ctx, membersKey, member.UserID)

	// Track user's team
	userTeamsKey := utils.GetUserTeamsKey(member.UserID)
	pipe.SAdd(ctx, userTeamsKey, member.ContestID)
	pipe.Expire(ctx, userTeamsKey, 30*24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// Update team count
	count, err := a.GetMemberCount(ctx, member.ContestID)
	if err != nil {
		return err
	}
	return a.UpdateTeamCount(ctx, member.ContestID, count)
}

// GetMember retrieves a member from Redis
func (a *TeamRedisAdapter) GetMember(ctx context.Context, contestID, userID int64) (*port.CachedTeamMember, error) {
	memberKey := utils.GetTeamMemberKey(contestID, userID)
	data, err := a.client.Get(ctx, memberKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, exception.ErrTeamMemberNotFound
	}
	if err != nil {
		return nil, err
	}

	var member port.CachedTeamMember
	if err := json.Unmarshal([]byte(data), &member); err != nil {
		return nil, err
	}

	return &member, nil
}

// GetAllMembers retrieves all team members from Redis
func (a *TeamRedisAdapter) GetAllMembers(ctx context.Context, contestID int64) ([]*port.CachedTeamMember, error) {
	membersKey := utils.GetTeamMembersKey(contestID)
	userIDs, err := a.client.SMembers(ctx, membersKey).Result()
	if err != nil {
		return nil, err
	}

	members := make([]*port.CachedTeamMember, 0, len(userIDs))
	for _, userIDStr := range userIDs {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			continue
		}

		member, err := a.GetMember(ctx, contestID, userID)
		if err != nil {
			continue
		}
		members = append(members, member)
	}

	return members, nil
}

// RemoveMember removes a member from the team
func (a *TeamRedisAdapter) RemoveMember(ctx context.Context, contestID, userID int64) error {
	pipe := a.client.Pipeline()

	// Remove member data
	memberKey := utils.GetTeamMemberKey(contestID, userID)
	pipe.Del(ctx, memberKey)

	// Remove from members set
	membersKey := utils.GetTeamMembersKey(contestID)
	pipe.SRem(ctx, membersKey, userID)

	// Remove from user's teams
	userTeamsKey := utils.GetUserTeamsKey(userID)
	pipe.SRem(ctx, userTeamsKey, contestID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// Update team count
	count, err := a.GetMemberCount(ctx, contestID)
	if err != nil {
		return err
	}
	return a.UpdateTeamCount(ctx, contestID, count)
}

// GetMemberCount returns the number of members in the team
func (a *TeamRedisAdapter) GetMemberCount(ctx context.Context, contestID int64) (int, error) {
	membersKey := utils.GetTeamMembersKey(contestID)
	count, err := a.client.SCard(ctx, membersKey).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// IsMember checks if a user is a member of the team
func (a *TeamRedisAdapter) IsMember(ctx context.Context, contestID, userID int64) (bool, error) {
	membersKey := utils.GetTeamMembersKey(contestID)
	return a.client.SIsMember(ctx, membersKey, userID).Result()
}

// CreateInvite creates a new team invitation
func (a *TeamRedisAdapter) CreateInvite(ctx context.Context, invite *port.TeamInvite, ttl time.Duration) error {
	pipe := a.client.Pipeline()

	// Store invite data
	inviteKey := utils.GetTeamInviteKey(invite.ContestID, invite.InviteeID)
	inviteData, err := json.Marshal(invite)
	if err != nil {
		return err
	}
	pipe.Set(ctx, inviteKey, inviteData, ttl)

	// Add to pending invites set
	invitesKey := utils.GetTeamInvitesKey(invite.ContestID)
	pipe.SAdd(ctx, invitesKey, invite.InviteeID)
	pipe.Expire(ctx, invitesKey, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// GetInvite retrieves an invite from Redis
func (a *TeamRedisAdapter) GetInvite(ctx context.Context, contestID, inviteeID int64) (*port.TeamInvite, error) {
	inviteKey := utils.GetTeamInviteKey(contestID, inviteeID)
	data, err := a.client.Get(ctx, inviteKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, exception.ErrTeamInviteNotFound
	}
	if err != nil {
		return nil, err
	}

	var invite port.TeamInvite
	if err := json.Unmarshal([]byte(data), &invite); err != nil {
		return nil, err
	}

	return &invite, nil
}

// GetPendingInvites retrieves all pending invites for a team
func (a *TeamRedisAdapter) GetPendingInvites(ctx context.Context, contestID int64) ([]*port.TeamInvite, error) {
	invitesKey := utils.GetTeamInvitesKey(contestID)
	inviteeIDs, err := a.client.SMembers(ctx, invitesKey).Result()
	if err != nil {
		return nil, err
	}

	invites := make([]*port.TeamInvite, 0, len(inviteeIDs))
	for _, inviteeIDStr := range inviteeIDs {
		inviteeID, err := strconv.ParseInt(inviteeIDStr, 10, 64)
		if err != nil {
			continue
		}

		invite, err := a.GetInvite(ctx, contestID, inviteeID)
		if err != nil {
			continue
		}

		if invite.Status == port.InviteStatusPending {
			invites = append(invites, invite)
		}
	}

	return invites, nil
}

// AcceptInvite marks an invite as accepted
func (a *TeamRedisAdapter) AcceptInvite(ctx context.Context, contestID, inviteeID int64) error {
	invite, err := a.GetInvite(ctx, contestID, inviteeID)
	if err != nil {
		return err
	}

	if invite.Status != port.InviteStatusPending {
		return exception.ErrTeamInviteNotPending
	}

	now := time.Now()
	invite.Status = port.InviteStatusAccepted
	invite.RespondedAt = &now

	inviteKey := utils.GetTeamInviteKey(contestID, inviteeID)
	inviteData, err := json.Marshal(invite)
	if err != nil {
		return err
	}

	ttl := a.client.TTL(ctx, inviteKey).Val()
	return a.client.Set(ctx, inviteKey, inviteData, ttl).Err()
}

// RejectInvite marks an invite as rejected
func (a *TeamRedisAdapter) RejectInvite(ctx context.Context, contestID, inviteeID int64) error {
	invite, err := a.GetInvite(ctx, contestID, inviteeID)
	if err != nil {
		return err
	}

	if invite.Status != port.InviteStatusPending {
		return exception.ErrTeamInviteNotPending
	}

	pipe := a.client.Pipeline()

	now := time.Now()
	invite.Status = port.InviteStatusRejected
	invite.RespondedAt = &now

	inviteKey := utils.GetTeamInviteKey(contestID, inviteeID)
	inviteData, err := json.Marshal(invite)
	if err != nil {
		return err
	}

	ttl := a.client.TTL(ctx, inviteKey).Val()
	pipe.Set(ctx, inviteKey, inviteData, ttl)

	// Remove from pending invites set
	invitesKey := utils.GetTeamInvitesKey(contestID)
	pipe.SRem(ctx, invitesKey, inviteeID)

	_, err = pipe.Exec(ctx)
	return err
}

// CancelInvite removes an invite
func (a *TeamRedisAdapter) CancelInvite(ctx context.Context, contestID, inviteeID int64) error {
	pipe := a.client.Pipeline()

	// Remove invite data
	inviteKey := utils.GetTeamInviteKey(contestID, inviteeID)
	pipe.Del(ctx, inviteKey)

	// Remove from pending invites set
	invitesKey := utils.GetTeamInvitesKey(contestID)
	pipe.SRem(ctx, invitesKey, inviteeID)

	_, err := pipe.Exec(ctx)
	return err
}

// HasPendingInvite checks if a user has a pending invite
func (a *TeamRedisAdapter) HasPendingInvite(ctx context.Context, contestID, inviteeID int64) (bool, error) {
	invite, err := a.GetInvite(ctx, contestID, inviteeID)
	if err != nil {
		if errors.Is(err, exception.ErrTeamInviteNotFound) {
			return false, nil
		}
		return false, err
	}

	return invite.Status == port.InviteStatusPending, nil
}

// TransferLeadership transfers leadership to another member
func (a *TeamRedisAdapter) TransferLeadership(ctx context.Context, contestID, currentLeaderID, newLeaderID int64) error {
	// Get current leader
	currentLeader, err := a.GetMember(ctx, contestID, currentLeaderID)
	if err != nil {
		return err
	}

	// Get new leader
	newLeader, err := a.GetMember(ctx, contestID, newLeaderID)
	if err != nil {
		return err
	}

	pipe := a.client.Pipeline()

	// Update current leader to member
	currentLeader.MemberType = port.TeamMemberTypeMember
	currentLeaderData, _ := json.Marshal(currentLeader)
	currentLeaderKey := utils.GetTeamMemberKey(contestID, currentLeaderID)
	ttl := a.client.TTL(ctx, currentLeaderKey).Val()
	pipe.Set(ctx, currentLeaderKey, currentLeaderData, ttl)

	// Update new leader
	newLeader.MemberType = port.TeamMemberTypeLeader
	newLeaderData, _ := json.Marshal(newLeader)
	newLeaderKey := utils.GetTeamMemberKey(contestID, newLeaderID)
	pipe.Set(ctx, newLeaderKey, newLeaderData, ttl)

	// Update team metadata
	team, err := a.GetTeam(ctx, contestID)
	if err != nil {
		return err
	}
	team.LeaderUserID = newLeaderID
	teamData, _ := json.Marshal(team)
	teamKey := utils.GetTeamKey(contestID)
	pipe.Set(ctx, teamKey, teamData, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// GetLeader retrieves the team leader
func (a *TeamRedisAdapter) GetLeader(ctx context.Context, contestID int64) (*port.CachedTeamMember, error) {
	team, err := a.GetTeam(ctx, contestID)
	if err != nil {
		return nil, err
	}

	return a.GetMember(ctx, contestID, team.LeaderUserID)
}

// MarkAsFinalized marks the team as finalized (ready for DB persistence)
func (a *TeamRedisAdapter) MarkAsFinalized(ctx context.Context, contestID int64) error {
	team, err := a.GetTeam(ctx, contestID)
	if err != nil {
		return err
	}

	now := time.Now()
	team.IsFinalized = true
	team.FinalizedAt = &now

	teamKey := utils.GetTeamKey(contestID)
	teamData, err := json.Marshal(team)
	if err != nil {
		return err
	}

	ttl := a.client.TTL(ctx, teamKey).Val()
	return a.client.Set(ctx, teamKey, teamData, ttl).Err()
}

// IsFinalized checks if the team is finalized
func (a *TeamRedisAdapter) IsFinalized(ctx context.Context, contestID int64) (bool, error) {
	team, err := a.GetTeam(ctx, contestID)
	if err != nil {
		return false, err
	}

	return team.IsFinalized, nil
}

// AddUserTeam tracks a user's team membership
func (a *TeamRedisAdapter) AddUserTeam(ctx context.Context, userID, contestID int64, ttl time.Duration) error {
	userTeamsKey := utils.GetUserTeamsKey(userID)
	pipe := a.client.Pipeline()
	pipe.SAdd(ctx, userTeamsKey, contestID)
	pipe.Expire(ctx, userTeamsKey, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// RemoveUserTeam removes a user's team membership tracking
func (a *TeamRedisAdapter) RemoveUserTeam(ctx context.Context, userID, contestID int64) error {
	userTeamsKey := utils.GetUserTeamsKey(userID)
	return a.client.SRem(ctx, userTeamsKey, contestID).Err()
}

// GetUserTeams retrieves all contest IDs a user belongs to
func (a *TeamRedisAdapter) GetUserTeams(ctx context.Context, userID int64) ([]int64, error) {
	userTeamsKey := utils.GetUserTeamsKey(userID)
	members, err := a.client.SMembers(ctx, userTeamsKey).Result()
	if err != nil {
		return nil, err
	}

	contestIDs := make([]int64, 0, len(members))
	for _, member := range members {
		contestID, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			continue
		}
		contestIDs = append(contestIDs, contestID)
	}

	return contestIDs, nil
}

// ClearTeam removes all team-related data from Redis
func (a *TeamRedisAdapter) ClearTeam(ctx context.Context, contestID int64) error {
	// Get all members to clean up user tracking
	members, _ := a.GetAllMembers(ctx, contestID)

	pipe := a.client.Pipeline()

	// Remove user team tracking for each member
	for _, member := range members {
		userTeamsKey := utils.GetUserTeamsKey(member.UserID)
		pipe.SRem(ctx, userTeamsKey, contestID)
	}

	// Scan and delete all keys matching the pattern
	pattern := utils.GetContestTeamPatternKey(contestID)
	var cursor uint64
	for {
		keys, nextCursor, err := a.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			pipe.Del(ctx, key)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// ExtendTTL extends the TTL for all team-related keys
func (a *TeamRedisAdapter) ExtendTTL(ctx context.Context, contestID int64, newTTL time.Duration) error {
	pattern := utils.GetContestTeamPatternKey(contestID)

	var cursor uint64
	for {
		keys, nextCursor, err := a.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		pipe := a.client.Pipeline()
		for _, key := range keys {
			pipe.Expire(ctx, key, newTTL)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			return err
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
