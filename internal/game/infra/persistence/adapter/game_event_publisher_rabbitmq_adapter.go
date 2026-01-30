package adapter

import (
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/global/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// GameEventPublisherRabbitMQAdapter publishes game events to RabbitMQ
type GameEventPublisherRabbitMQAdapter struct {
	connection *config.RabbitMQConnection
	exchange   string
}

func NewGameEventPublisherRabbitMQAdapter(
	connection *config.RabbitMQConnection,
	exchange string,
) *GameEventPublisherRabbitMQAdapter {
	return &GameEventPublisherRabbitMQAdapter{
		connection: connection,
		exchange:   exchange,
	}
}

func (a *GameEventPublisherRabbitMQAdapter) PublishGameEvent(ctx context.Context, event *port.GameEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	return a.publish(ctx, event.EventID, string(event.EventType), event.Timestamp, event)
}

func (a *GameEventPublisherRabbitMQAdapter) PublishMatchDetectedEvent(ctx context.Context, event *port.MatchDetectedEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	return a.publish(ctx, event.EventID, string(event.EventType), event.Timestamp, event)
}

func (a *GameEventPublisherRabbitMQAdapter) publish(
	ctx context.Context,
	messageID, routingKey string,
	timestamp time.Time,
	payload interface{},
) error {
	channel, err := a.connection.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		publishCtx,
		a.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    timestamp,
			MessageId:    messageID,
			Body:         body,
			Headers: amqp.Table{
				"event_type": routingKey,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish game event: %w", err)
	}

	return nil
}
