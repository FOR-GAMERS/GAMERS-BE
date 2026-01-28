package dto

import (
	"GAMERS-BE/internal/notification/domain"
	"time"
)

// NotificationResponse represents a single notification response
type NotificationResponse struct {
	ID        int64                    `json:"id"`
	Type      domain.NotificationType  `json:"type"`
	Title     string                   `json:"title"`
	Message   string                   `json:"message"`
	Data      map[string]interface{}   `json:"data,omitempty"`
	IsRead    bool                     `json:"is_read"`
	CreatedAt time.Time                `json:"created_at"`
}

// NotificationListResponse represents a list of notifications
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
	Total         int64                  `json:"total"`
}

// FromNotification converts a domain notification to a response DTO
func FromNotification(n *domain.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Data:      n.NotificationData(),
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
	}
}

// FromNotifications converts a list of domain notifications to response DTOs
func FromNotifications(notifications []*domain.Notification) []NotificationResponse {
	result := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = *FromNotification(n)
	}
	return result
}

// SSEEvent represents an event sent through SSE
type SSEEvent struct {
	ID        string                   `json:"id"`
	Type      domain.NotificationType  `json:"type"`
	Title     string                   `json:"title"`
	Message   string                   `json:"message"`
	Data      map[string]interface{}   `json:"data,omitempty"`
	Timestamp time.Time                `json:"timestamp"`
}

// GetNotificationsRequest represents the query parameters for listing notifications
type GetNotificationsRequest struct {
	Limit  int  `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int  `form:"offset" binding:"omitempty,min=0"`
	Unread bool `form:"unread"`
}
