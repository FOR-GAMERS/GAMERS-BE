package database

import (
	"fmt"
	"log"
)

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
	Exchange string
	Queue    string
}

func NewRabbitMQConfigFromEnv() *RabbitMQConfig {
	return &RabbitMQConfig{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnv("RABBITMQ_PORT", "5672"),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
		Exchange: getEnv("RABBITMQ_EXCHANGE", "gamers.events"),
		Queue:    getEnv("RABBITMQ_QUEUE_NOTIFICATIONS", "notifications.contest.applications"),
	}
}

func (c *RabbitMQConfig) URI() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.VHost,
	)
}

func InitRabbitMQ(config *RabbitMQConfig) (*RabbitMQConnection, error) {
	conn := NewRabbitMQConnection(config)

	if err := conn.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	if err := conn.SetupTopology(); err != nil {
		err := conn.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to setup RabbitMQ topology: %w", err)
	}

	log.Println("RabbitMQ connected successfully")
	return conn, nil
}
