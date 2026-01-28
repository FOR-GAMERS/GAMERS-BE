package port

import (
	"context"
)

// TeamPersistenceHandler is the function type for handling persistence events
type TeamPersistenceHandler func(ctx context.Context, event *TeamPersistenceEvent) error

// TeamPersistenceConsumerPort defines the interface for consuming team persistence events
type TeamPersistenceConsumerPort interface {
	// Start begins consuming messages from the persistence queue
	Start(ctx context.Context, handler TeamPersistenceHandler) error

	// Stop gracefully stops the consumer
	Stop() error

	// IsRunning returns whether the consumer is currently running
	IsRunning() bool
}
