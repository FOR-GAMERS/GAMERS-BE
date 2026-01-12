package database

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	config      *RabbitMQConfig
	conn        *amqp.Connection
	channel     *amqp.Channel
	mu          sync.RWMutex
	closed      bool
	reconnectCh chan struct{}
}

func NewRabbitMQConnection(config *RabbitMQConfig) *RabbitMQConnection {
	return &RabbitMQConnection{
		config:      config,
		reconnectCh: make(chan struct{}, 1),
	}
}

func (r *RabbitMQConnection) Connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp.Dial(r.config.URI())
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.conn = conn
	r.channel = channel

	// Monitor connection
	go r.monitorConnection()

	return nil
}

func (r *RabbitMQConnection) monitorConnection() {
	closeChan := make(chan *amqp.Error)
	r.conn.NotifyClose(closeChan)

	err := <-closeChan
	if err != nil && !r.closed {
		// Connection lost, attempt reconnect
		select {
		case r.reconnectCh <- struct{}{}:
		default:
		}
	}
}

func (r *RabbitMQConnection) SetupTopology() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Declare exchange
	err := r.channel.ExchangeDeclare(
		r.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = r.channel.QueueDeclare(
		r.config.Queue, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with routing key pattern
	err = r.channel.QueueBind(
		r.config.Queue,          // queue name
		"contest.application.*", // routing key pattern
		r.config.Exchange,       // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

func (r *RabbitMQConnection) GetChannel() (*amqp.Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, fmt.Errorf("connection is closed")
	}

	if r.channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	return r.channel, nil
}

func (r *RabbitMQConnection) Config() *RabbitMQConfig {
	return r.config
}

func (r *RabbitMQConnection) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true

	if r.channel != nil {
		r.channel.Close()
	}

	if r.conn != nil {
		return r.conn.Close()
	}

	return nil
}

func (r *RabbitMQConnection) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return fmt.Errorf("connection is closed")
	}

	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("connection is not alive")
	}

	return nil
}
