package port

import (
	"context"
	"time"
)

// EventType represents the type of event being published
type EventType string

const (
	EventTypeApplicationRequested EventType = "application.requested"
	EventTypeApplicationAccepted  EventType = "application.accepted"
	EventTypeApplicationRejected  EventType = "application.rejected"
	EventTypeMemberWithdrawn      EventType = "member.withdrawn"
	EventTypeContestCreated       EventType = "contest.created"
)

// ContestApplicationEvent represents an event in the contest application lifecycle
type ContestApplicationEvent struct {
	EventID              string                 `json:"event_id"`
	EventType            EventType              `json:"event_type"`
	Timestamp            time.Time              `json:"timestamp"`
	ContestID            int64                  `json:"contest_id"`
	UserID               int64                  `json:"user_id"`
	DiscordUserID        string                 `json:"discord_user_id"`
	DiscordGuildID       string                 `json:"discord_guild_id"`
	DiscordTextChannelID string                 `json:"discord_text_channel_id"`
	Data                 map[string]interface{} `json:"data"`
}

// ContestCreatedEvent represents an event when a new contest is created
type ContestCreatedEvent struct {
	EventID              string                 `json:"event_id"`
	EventType            EventType              `json:"event_type"`
	Timestamp            time.Time              `json:"timestamp"`
	ContestID            int64                  `json:"contest_id"`
	CreatorUserID        int64                  `json:"creator_user_id"`
	CreatorDiscordID     string                 `json:"creator_discord_id"`
	DiscordGuildID       string                 `json:"discord_guild_id"`
	DiscordTextChannelID string                 `json:"discord_text_channel_id"`
	ContestTitle         string                 `json:"contest_title"`
	Data                 map[string]interface{} `json:"data"`
}

// EventPublisherPort defines the interface for publishing domain events
type EventPublisherPort interface {
	// PublishContestApplicationEvent publishes a contest application event
	PublishContestApplicationEvent(ctx context.Context, event *ContestApplicationEvent) error

	// PublishContestCreatedEvent publishes a contest created event
	PublishContestCreatedEvent(ctx context.Context, event *ContestCreatedEvent) error

	// Close gracefully shuts down the publisher
	Close() error

	// HealthCheck verifies the connection is alive
	HealthCheck(ctx context.Context) error
}
