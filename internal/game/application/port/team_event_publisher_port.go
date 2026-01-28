package port

import (
	"context"
	"time"
)

// TeamEventType represents the type of team event being published
type TeamEventType string

const (
	TeamEventTypeInviteSent         TeamEventType = "team.invite.sent"
	TeamEventTypeInviteAccepted     TeamEventType = "team.invite.accepted"
	TeamEventTypeInviteRejected     TeamEventType = "team.invite.rejected"
	TeamEventTypeMemberJoined       TeamEventType = "team.member.joined"
	TeamEventTypeMemberLeft         TeamEventType = "team.member.left"
	TeamEventTypeMemberKicked       TeamEventType = "team.member.kicked"
	TeamEventTypeLeadershipTransfer TeamEventType = "team.leadership.transferred"
	TeamEventTypeTeamFinalized      TeamEventType = "team.finalized"
	TeamEventTypeTeamDeleted        TeamEventType = "team.deleted"
)

// TeamInviteEvent represents an event when a team invite is sent
type TeamInviteEvent struct {
	EventID              string                 `json:"event_id"`
	EventType            TeamEventType          `json:"event_type"`
	Timestamp            time.Time              `json:"timestamp"`
	ContestID            int64                  `json:"contest_id"`
	InviterUserID        int64                  `json:"inviter_user_id"`
	InviterDiscordID     string                 `json:"inviter_discord_id"`
	InviterUsername      string                 `json:"inviter_username"`
	InviteeUserID        int64                  `json:"invitee_user_id"`
	InviteeDiscordID     string                 `json:"invitee_discord_id"`
	InviteeUsername      string                 `json:"invitee_username"`
	DiscordGuildID       string                 `json:"discord_guild_id"`
	DiscordTextChannelID string                 `json:"discord_text_channel_id"`
	TeamName             string                 `json:"team_name,omitempty"`
	Data                 map[string]interface{} `json:"data,omitempty"`
}

// TeamMemberEvent represents events related to team membership changes
type TeamMemberEvent struct {
	EventID              string                 `json:"event_id"`
	EventType            TeamEventType          `json:"event_type"`
	Timestamp            time.Time              `json:"timestamp"`
	ContestID            int64                  `json:"contest_id"`
	UserID               int64                  `json:"user_id"`
	DiscordUserID        string                 `json:"discord_user_id"`
	Username             string                 `json:"username"`
	DiscordGuildID       string                 `json:"discord_guild_id"`
	DiscordTextChannelID string                 `json:"discord_text_channel_id"`
	CurrentMemberCount   int                    `json:"current_member_count"`
	MaxMembers           int                    `json:"max_members"`
	Data                 map[string]interface{} `json:"data,omitempty"`
}

// TeamFinalizedEvent represents an event when a team is finalized
type TeamFinalizedEvent struct {
	EventID              string                 `json:"event_id"`
	EventType            TeamEventType          `json:"event_type"`
	Timestamp            time.Time              `json:"timestamp"`
	ContestID            int64                  `json:"contest_id"`
	LeaderUserID         int64                  `json:"leader_user_id"`
	LeaderDiscordID      string                 `json:"leader_discord_id"`
	DiscordGuildID       string                 `json:"discord_guild_id"`
	DiscordTextChannelID string                 `json:"discord_text_channel_id"`
	MemberCount          int                    `json:"member_count"`
	MemberUserIDs        []int64                `json:"member_user_ids"`
	Data                 map[string]interface{} `json:"data,omitempty"`
}

// TeamEventPublisherPort defines the interface for publishing team-related events
type TeamEventPublisherPort interface {
	// PublishTeamInviteEvent publishes a team invite event
	PublishTeamInviteEvent(ctx context.Context, event *TeamInviteEvent) error

	// PublishTeamMemberEvent publishes a team member change event
	PublishTeamMemberEvent(ctx context.Context, event *TeamMemberEvent) error

	// PublishTeamFinalizedEvent publishes a team finalized event
	PublishTeamFinalizedEvent(ctx context.Context, event *TeamFinalizedEvent) error

	// Close gracefully shuts down the publisher
	Close() error

	// HealthCheck verifies the connection is alive
	HealthCheck(ctx context.Context) error
}
