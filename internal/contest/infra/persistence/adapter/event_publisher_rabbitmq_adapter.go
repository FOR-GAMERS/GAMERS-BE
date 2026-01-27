package adapter

import (
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/global/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventPublisherRabbitMQAdapter struct {
	connection *config.RabbitMQConnection
	exchange   string
}

func NewEventPublisherRabbitMQAdapter(connection *config.RabbitMQConnection, exchange string) *EventPublisherRabbitMQAdapter {
	return &EventPublisherRabbitMQAdapter{
		connection: connection,
		exchange:   exchange,
	}
}

func (a *EventPublisherRabbitMQAdapter) PublishContestApplicationEvent(
	ctx context.Context,
	event *port.ContestApplicationEvent,
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

	// Construct routing key
	routingKey := a.buildRoutingKey(event.EventType)

	// Publish with context timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		publishCtx,
		a.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Persistent messages
			Timestamp:    event.Timestamp,
			MessageId:    event.EventID,
			Body:         body,
			Headers: amqp.Table{
				"event_type":              string(event.EventType),
				"contest_id":              event.ContestID,
				"user_id":                 event.UserID,
				"discord_user_id":         event.DiscordUserID,
				"discord_guild_id":        event.DiscordGuildID,
				"discord_text_channel_id": event.DiscordTextChannelID,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (a *EventPublisherRabbitMQAdapter) PublishContestCreatedEvent(
	ctx context.Context,
	event *port.ContestCreatedEvent,
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

	// Construct routing key
	routingKey := a.buildRoutingKey(event.EventType)

	// Publish with context timeout
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		publishCtx,
		a.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Persistent messages
			Timestamp:    event.Timestamp,
			MessageId:    event.EventID,
			Body:         body,
			Headers: amqp.Table{
				"event_type":       string(event.EventType),
				"contest_id":       event.ContestID,
				"creator_user_id":  event.CreatorUserID,
				"discord_guild_id": event.DiscordGuildID,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (a *EventPublisherRabbitMQAdapter) buildRoutingKey(eventType port.EventType) string {
	return fmt.Sprintf("contest.%s", eventType)
}

func (a *EventPublisherRabbitMQAdapter) Close() error {
	return a.connection.Close()
}

func (a *EventPublisherRabbitMQAdapter) HealthCheck(ctx context.Context) error {
	return a.connection.HealthCheck(ctx)
}
