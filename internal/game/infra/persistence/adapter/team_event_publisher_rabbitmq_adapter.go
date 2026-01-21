package adapter

import (
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/global/database"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TeamEventPublisherRabbitMQAdapter struct {
	connection *database.RabbitMQConnection
	exchange   string
}

func NewTeamEventPublisherRabbitMQAdapter(connection *database.RabbitMQConnection, exchange string) *TeamEventPublisherRabbitMQAdapter {
	return &TeamEventPublisherRabbitMQAdapter{
		connection: connection,
		exchange:   exchange,
	}
}

func (a *TeamEventPublisherRabbitMQAdapter) PublishTeamInviteEvent(
	ctx context.Context,
	event *port.TeamInviteEvent,
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
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.Timestamp,
			MessageId:    event.EventID,
			Body:         body,
			Headers: amqp.Table{
				"event_type":              string(event.EventType),
				"game_id":                 event.GameID,
				"contest_id":              event.ContestID,
				"inviter_user_id":         event.InviterUserID,
				"inviter_discord_id":      event.InviterDiscordID,
				"invitee_user_id":         event.InviteeUserID,
				"invitee_discord_id":      event.InviteeDiscordID,
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

func (a *TeamEventPublisherRabbitMQAdapter) PublishTeamMemberEvent(
	ctx context.Context,
	event *port.TeamMemberEvent,
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
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.Timestamp,
			MessageId:    event.EventID,
			Body:         body,
			Headers: amqp.Table{
				"event_type":              string(event.EventType),
				"game_id":                 event.GameID,
				"contest_id":              event.ContestID,
				"user_id":                 event.UserID,
				"discord_user_id":         event.DiscordUserID,
				"discord_guild_id":        event.DiscordGuildID,
				"discord_text_channel_id": event.DiscordTextChannelID,
				"current_member_count":    event.CurrentMemberCount,
				"max_members":             event.MaxMembers,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (a *TeamEventPublisherRabbitMQAdapter) PublishTeamFinalizedEvent(
	ctx context.Context,
	event *port.TeamFinalizedEvent,
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
			DeliveryMode: amqp.Persistent,
			Timestamp:    event.Timestamp,
			MessageId:    event.EventID,
			Body:         body,
			Headers: amqp.Table{
				"event_type":              string(event.EventType),
				"game_id":                 event.GameID,
				"contest_id":              event.ContestID,
				"leader_user_id":          event.LeaderUserID,
				"leader_discord_id":       event.LeaderDiscordID,
				"discord_guild_id":        event.DiscordGuildID,
				"discord_text_channel_id": event.DiscordTextChannelID,
				"member_count":            event.MemberCount,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (a *TeamEventPublisherRabbitMQAdapter) buildRoutingKey(eventType port.TeamEventType) string {
	return fmt.Sprintf("game.%s", eventType)
}

func (a *TeamEventPublisherRabbitMQAdapter) Close() error {
	return a.connection.Close()
}

func (a *TeamEventPublisherRabbitMQAdapter) HealthCheck(ctx context.Context) error {
	return a.connection.HealthCheck(ctx)
}
