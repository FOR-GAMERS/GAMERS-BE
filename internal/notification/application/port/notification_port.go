package port

import (
	"GAMERS-BE/internal/notification/domain"
	"context"
)

// NotificationDatabasePort defines the interface for notification database operations
type NotificationDatabasePort interface {
	// Save saves a new notification to the database
	Save(notification *domain.Notification) error

	// FindByID finds a notification by ID
	FindByID(id int64) (*domain.Notification, error)

	// FindByUserID finds all notifications for a user with pagination
	FindByUserID(userID int64, limit, offset int) ([]*domain.Notification, error)

	// FindUnreadByUserID finds unread notifications for a user
	FindUnreadByUserID(userID int64, limit, offset int) ([]*domain.Notification, error)

	// CountByUserID counts total notifications for a user
	CountByUserID(userID int64) (int64, error)

	// CountUnreadByUserID counts unread notifications for a user
	CountUnreadByUserID(userID int64) (int64, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(id int64) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(userID int64) error

	// Delete deletes a notification
	Delete(id int64) error

	// DeleteOldNotifications deletes notifications older than a certain time
	DeleteOldNotifications(days int) error
}

// SSEClientPort defines the interface for SSE client operations
type SSEClientPort interface {
	// Send sends a message to the client
	Send(message *domain.SSEMessage) error

	// Close closes the client connection
	Close()

	// IsClosed checks if the client connection is closed
	IsClosed() bool

	// GetUserID returns the user ID of the client
	GetUserID() int64
}

// SSEManagerPort defines the interface for managing SSE connections
type SSEManagerPort interface {
	// RegisterClient registers a new SSE client
	RegisterClient(client SSEClientPort)

	// UnregisterClient removes an SSE client
	UnregisterClient(client SSEClientPort)

	// SendToUser sends a notification to a specific user
	SendToUser(userID int64, message *domain.SSEMessage) error

	// Broadcast sends a notification to all connected clients
	Broadcast(message *domain.SSEMessage)

	// GetConnectedUsers returns the list of connected user IDs
	GetConnectedUsers() []int64

	// IsUserConnected checks if a user is connected
	IsUserConnected(userID int64) bool
}

// NotificationConsumerPort defines the interface for consuming notification events
type NotificationConsumerPort interface {
	// Start starts consuming messages from the queue
	Start(ctx context.Context) error

	// Stop stops the consumer
	Stop() error

	// SetHandler sets the handler function for processing notifications
	SetHandler(handler func(notification *domain.Notification) error)
}
