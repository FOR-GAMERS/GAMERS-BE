package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/game/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TeamPersistenceConsumerRabbitMQAdapter implements TeamPersistenceConsumerPort
type TeamPersistenceConsumerRabbitMQAdapter struct {
	connection *config.RabbitMQConnection
	exchange   string
	running    bool
	stopCh     chan struct{}
	mu         sync.RWMutex
}

// NewTeamPersistenceConsumerRabbitMQAdapter creates a new consumer adapter
func NewTeamPersistenceConsumerRabbitMQAdapter(
	connection *config.RabbitMQConnection,
	exchange string,
) *TeamPersistenceConsumerRabbitMQAdapter {
	return &TeamPersistenceConsumerRabbitMQAdapter{
		connection: connection,
		exchange:   exchange,
		stopCh:     make(chan struct{}),
	}
}

// Start begins consuming messages from the persistence queue
func (a *TeamPersistenceConsumerRabbitMQAdapter) Start(
	ctx context.Context,
	handler port.TeamPersistenceHandler,
) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("consumer is already running")
	}
	a.running = true
	a.stopCh = make(chan struct{})
	a.mu.Unlock()

	channel, err := a.connection.GetChannel()
	if err != nil {
		a.setRunning(false)
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Set prefetch to 1 for ordered processing
	if err := channel.Qos(1, 0, false); err != nil {
		a.setRunning(false)
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming
	deliveries, err := channel.Consume(
		config.TeamPersistenceQueue, // queue
		"",                          // consumer tag (auto-generated)
		false,                       // auto-ack (manual ack for reliability)
		false,                       // exclusive
		false,                       // no-local
		false,                       // no-wait
		nil,                         // args
	)
	if err != nil {
		a.setRunning(false)
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Printf("Team persistence consumer started, listening on queue: %s", config.TeamPersistenceQueue)

	// Start processing in a goroutine
	go a.processMessages(ctx, deliveries, handler)

	return nil
}

// processMessages handles incoming messages
func (a *TeamPersistenceConsumerRabbitMQAdapter) processMessages(
	ctx context.Context,
	deliveries <-chan amqp.Delivery,
	handler port.TeamPersistenceHandler,
) {
	for {
		select {
		case <-a.stopCh:
			log.Println("Team persistence consumer stopped")
			return
		case <-ctx.Done():
			log.Println("Team persistence consumer context cancelled")
			a.setRunning(false)
			return
		case delivery, ok := <-deliveries:
			if !ok {
				log.Println("Team persistence consumer channel closed")
				a.setRunning(false)
				return
			}
			a.handleDelivery(ctx, delivery, handler)
		}
	}
}

// handleDelivery processes a single message
func (a *TeamPersistenceConsumerRabbitMQAdapter) handleDelivery(
	ctx context.Context,
	delivery amqp.Delivery,
	handler port.TeamPersistenceHandler,
) {
	var event port.TeamPersistenceEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		log.Printf("Failed to unmarshal team persistence event: %v", err)
		// Reject without requeue for malformed messages
		_ = delivery.Nack(false, false)
		return
	}

	log.Printf("Processing team persistence event: type=%s, contestID=%d, teamID=%d, retry=%d",
		event.EventType, event.ContestID, event.TeamID, event.RetryCount)

	// Process the event with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := handler(processCtx, &event); err != nil {
		log.Printf("Failed to handle team persistence event: %v", err)
		a.handleFailure(delivery, &event, err)
		return
	}

	// Successfully processed - acknowledge
	if err := delivery.Ack(false); err != nil {
		log.Printf("Failed to ack message: %v", err)
	} else {
		log.Printf("Successfully processed team persistence event: type=%s, contestID=%d",
			event.EventType, event.ContestID)
	}
}

// handleFailure handles a failed message processing
func (a *TeamPersistenceConsumerRabbitMQAdapter) handleFailure(
	delivery amqp.Delivery,
	event *port.TeamPersistenceEvent,
	err error,
) {
	event.RetryCount++

	if event.RetryCount >= port.MaxRetryCount {
		// Max retries exceeded - send to DLQ
		log.Printf("Max retries exceeded for event %s (contestID=%d), sending to DLQ: %v",
			event.EventID, event.ContestID, err)
		// Reject without requeue - will go to DLQ due to queue configuration
		_ = delivery.Nack(false, false)
		return
	}

	// Requeue for retry with updated retry count
	// We need to republish with updated retry count since we can't modify the original message
	log.Printf("Requeueing event %s for retry (attempt %d/%d): %v",
		event.EventID, event.RetryCount, port.MaxRetryCount, err)

	// Republish with incremented retry count
	if pubErr := a.republishWithRetry(event); pubErr != nil {
		log.Printf("Failed to republish event for retry, sending to DLQ: %v", pubErr)
		_ = delivery.Nack(false, false)
		return
	}

	// Acknowledge the original message since we've republished it
	_ = delivery.Ack(false)
}

// republishWithRetry republishes the event with incremented retry count
func (a *TeamPersistenceConsumerRabbitMQAdapter) republishWithRetry(event *port.TeamPersistenceEvent) error {
	channel, err := a.connection.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	routingKey := string(event.EventType)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return channel.PublishWithContext(
		ctx,
		a.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
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
}

// Stop gracefully stops the consumer
func (a *TeamPersistenceConsumerRabbitMQAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	close(a.stopCh)
	a.running = false
	return nil
}

// IsRunning returns whether the consumer is currently running
func (a *TeamPersistenceConsumerRabbitMQAdapter) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// setRunning safely sets the running state
func (a *TeamPersistenceConsumerRabbitMQAdapter) setRunning(running bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = running
}
