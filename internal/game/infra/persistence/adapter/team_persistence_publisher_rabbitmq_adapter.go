package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/game/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TeamPersistencePublisherRabbitMQAdapter implements TeamPersistencePublisherPort
type TeamPersistencePublisherRabbitMQAdapter struct {
	connection *config.RabbitMQConnection
	exchange   string
}

// NewTeamPersistencePublisherRabbitMQAdapter creates a new publisher adapter
func NewTeamPersistencePublisherRabbitMQAdapter(
	connection *config.RabbitMQConnection,
	exchange string,
) *TeamPersistencePublisherRabbitMQAdapter {
	return &TeamPersistencePublisherRabbitMQAdapter{
		connection: connection,
		exchange:   exchange,
	}
}

// PublishTeamCreated publishes an event when a team is created in cache
func (a *TeamPersistencePublisherRabbitMQAdapter) PublishTeamCreated(
	ctx context.Context,
	event *port.TeamPersistenceEvent,
) error {
	event.EventType = port.TeamPersistenceCreated
	return a.publish(ctx, event)
}

// PublishMemberAdded publishes an event when a member is added to a team
func (a *TeamPersistencePublisherRabbitMQAdapter) PublishMemberAdded(
	ctx context.Context,
	event *port.TeamPersistenceEvent,
) error {
	event.EventType = port.TeamPersistenceMemberAdded
	return a.publish(ctx, event)
}

// PublishMemberRemoved publishes an event when a member is removed from a team
func (a *TeamPersistencePublisherRabbitMQAdapter) PublishMemberRemoved(
	ctx context.Context,
	event *port.TeamPersistenceEvent,
) error {
	event.EventType = port.TeamPersistenceMemberRemoved
	return a.publish(ctx, event)
}

// PublishTeamFinalized publishes an event when a team is finalized
func (a *TeamPersistencePublisherRabbitMQAdapter) PublishTeamFinalized(
	ctx context.Context,
	event *port.TeamPersistenceEvent,
) error {
	event.EventType = port.TeamPersistenceFinalized
	return a.publish(ctx, event)
}

// PublishTeamDeleted publishes an event when a team is deleted
func (a *TeamPersistencePublisherRabbitMQAdapter) PublishTeamDeleted(
	ctx context.Context,
	event *port.TeamPersistenceEvent,
) error {
	event.EventType = port.TeamPersistenceDeleted
	return a.publish(ctx, event)
}

// publish sends the event to RabbitMQ
func (a *TeamPersistencePublisherRabbitMQAdapter) publish(
	ctx context.Context,
	event *port.TeamPersistenceEvent,
) error {
	channel, err := a.connection.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Ensure event has ID and timestamp
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Build routing key from event type
	routingKey := string(event.EventType)

	// Publish with context timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		publishCtx,
		a.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.Timestamp,
			MessageId:    event.EventID,
			Body:         body,
			Headers: amqp.Table{
				"event_type":  string(event.EventType),
				"contest_id":  event.ContestID,
				"team_id":     event.TeamID,
				"retry_count": event.RetryCount,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Close gracefully shuts down the publisher
func (a *TeamPersistencePublisherRabbitMQAdapter) Close() error {
	return a.connection.Close()
}

// HealthCheck verifies the connection is alive
func (a *TeamPersistencePublisherRabbitMQAdapter) HealthCheck(ctx context.Context) error {
	return a.connection.HealthCheck(ctx)
}
