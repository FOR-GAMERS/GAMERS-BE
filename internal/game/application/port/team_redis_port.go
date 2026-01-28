package port

import (
	"context"
	"time"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "PENDING"
	InviteStatusAccepted InviteStatus = "ACCEPTED"
	InviteStatusRejected InviteStatus = "REJECTED"
	InviteStatusExpired  InviteStatus = "EXPIRED"
)

type TeamMemberType string

const (
	TeamMemberTypeLeader TeamMemberType = "LEADER"
	TeamMemberTypeMember TeamMemberType = "MEMBER"
)

// CachedTeamMember represents a team member stored in Redis cache
type CachedTeamMember struct {
	UserID     int64          `json:"user_id"`
	ContestID  int64          `json:"contest_id"`
	TeamID     int64          `json:"team_id"`
	MemberType TeamMemberType `json:"member_type"`
	JoinedAt   time.Time      `json:"joined_at"`
	DiscordID  string         `json:"discord_id,omitempty"`
	Username   string         `json:"username,omitempty"`
	Tag        string         `json:"tag,omitempty"`
}

// CachedTeam represents team metadata stored in Redis cache
type CachedTeam struct {
	ContestID    int64      `json:"contest_id"`
	TeamID       int64      `json:"team_id"`
	TeamName     *string    `json:"team_name,omitempty"`
	MaxMembers   int        `json:"max_members"`
	CurrentCount int        `json:"current_count"`
	LeaderUserID int64      `json:"leader_user_id"`
	CreatedAt    time.Time  `json:"created_at"`
	IsFinalized  bool       `json:"is_finalized"`
	FinalizedAt  *time.Time `json:"finalized_at,omitempty"`
}

// TeamInvite represents a pending team invitation
type TeamInvite struct {
	ContestID    int64        `json:"contest_id"`
	InviterID    int64        `json:"inviter_id"`
	InviteeID    int64        `json:"invitee_id"`
	Status       InviteStatus `json:"status"`
	InvitedAt    time.Time    `json:"invited_at"`
	RespondedAt  *time.Time   `json:"responded_at,omitempty"`
	InviterName  string       `json:"inviter_name,omitempty"`
	InviteeName  string       `json:"invitee_name,omitempty"`
	DiscordID    string       `json:"discord_id,omitempty"`
}

// TeamRedisPort defines the interface for Team caching operations in Redis
type TeamRedisPort interface {
	// Team Management
	CreateTeam(ctx context.Context, team *CachedTeam, leader *CachedTeamMember, ttl time.Duration) error
	GetTeam(ctx context.Context, contestID int64) (*CachedTeam, error)
	UpdateTeamCount(ctx context.Context, contestID int64, count int) error
	DeleteTeam(ctx context.Context, contestID int64) error

	// Member Management
	AddMember(ctx context.Context, member *CachedTeamMember, ttl time.Duration) error
	GetMember(ctx context.Context, contestID, userID int64) (*CachedTeamMember, error)
	GetAllMembers(ctx context.Context, contestID int64) ([]*CachedTeamMember, error)
	RemoveMember(ctx context.Context, contestID, userID int64) error
	GetMemberCount(ctx context.Context, contestID int64) (int, error)
	IsMember(ctx context.Context, contestID, userID int64) (bool, error)

	// Invite Management
	CreateInvite(ctx context.Context, invite *TeamInvite, ttl time.Duration) error
	GetInvite(ctx context.Context, contestID, inviteeID int64) (*TeamInvite, error)
	GetPendingInvites(ctx context.Context, contestID int64) ([]*TeamInvite, error)
	AcceptInvite(ctx context.Context, contestID, inviteeID int64) error
	RejectInvite(ctx context.Context, contestID, inviteeID int64) error
	CancelInvite(ctx context.Context, contestID, inviteeID int64) error
	HasPendingInvite(ctx context.Context, contestID, inviteeID int64) (bool, error)

	// Leadership
	TransferLeadership(ctx context.Context, contestID, currentLeaderID, newLeaderID int64) error
	GetLeader(ctx context.Context, contestID int64) (*CachedTeamMember, error)

	// Finalization (move to DB)
	MarkAsFinalized(ctx context.Context, contestID int64) error
	IsFinalized(ctx context.Context, contestID int64) (bool, error)

	// User's teams tracking (by contestID)
	AddUserTeam(ctx context.Context, userID, contestID int64, ttl time.Duration) error
	RemoveUserTeam(ctx context.Context, userID, contestID int64) error
	GetUserTeams(ctx context.Context, userID int64) ([]int64, error)

	// Cleanup
	ClearTeam(ctx context.Context, contestID int64) error
	ExtendTTL(ctx context.Context, contestID int64, newTTL time.Duration) error
}
