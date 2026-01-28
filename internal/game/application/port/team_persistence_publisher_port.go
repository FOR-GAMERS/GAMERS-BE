package port

import (
	"context"
)

// TeamPersistencePublisherPort defines the interface for publishing team persistence events
type TeamPersistencePublisherPort interface {
	// PublishTeamCreated publishes an event when a team is created in cache
	PublishTeamCreated(ctx context.Context, event *TeamPersistenceEvent) error

	// PublishMemberAdded publishes an event when a member is added to a team
	PublishMemberAdded(ctx context.Context, event *TeamPersistenceEvent) error

	// PublishMemberRemoved publishes an event when a member is removed from a team
	PublishMemberRemoved(ctx context.Context, event *TeamPersistenceEvent) error

	// PublishTeamFinalized publishes an event when a team is finalized
	PublishTeamFinalized(ctx context.Context, event *TeamPersistenceEvent) error

	// PublishTeamDeleted publishes an event when a team is deleted
	PublishTeamDeleted(ctx context.Context, event *TeamPersistenceEvent) error

	// Close gracefully shuts down the publisher
	Close() error

	// HealthCheck verifies the connection is alive
	HealthCheck(ctx context.Context) error
}
