package domain

import (
	"encoding/json"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// Team invitation notifications
	NotificationTypeTeamInviteReceived NotificationType = "TEAM_INVITE_RECEIVED"
	NotificationTypeTeamInviteAccepted NotificationType = "TEAM_INVITE_ACCEPTED"
	NotificationTypeTeamInviteRejected NotificationType = "TEAM_INVITE_REJECTED"

	// Contest application notifications
	NotificationTypeApplicationAccepted NotificationType = "APPLICATION_ACCEPTED"
	NotificationTypeApplicationRejected NotificationType = "APPLICATION_REJECTED"
)

// Notification represents a user notification entity
type Notification struct {
	ID        int64            `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	UserID    int64            `gorm:"column:user_id;not null;index" json:"user_id"`
	Type      NotificationType `gorm:"column:type;type:varchar(50);not null" json:"type"`
	Title     string           `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Message   string           `gorm:"column:message;type:text" json:"message"`
	Data      string           `gorm:"column:data;type:json" json:"-"`
	IsRead    bool             `gorm:"column:is_read;default:false" json:"is_read"`
	CreatedAt time.Time        `gorm:"column:created_at;type:timestamp;autoCreateTime" json:"created_at"`
}

func (n *Notification) TableName() string {
	return "notifications"
}

// NotificationData returns the parsed JSON data
func (n *Notification) NotificationData() map[string]interface{} {
	if n.Data == "" {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(n.Data), &data); err != nil {
		return nil
	}
	return data
}

// SetNotificationData sets the JSON data from a map
func (n *Notification) SetNotificationData(data map[string]interface{}) error {
	if data == nil {
		n.Data = ""
		return nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	n.Data = string(bytes)
	return nil
}

// NewNotification creates a new notification
func NewNotification(userID int64, notifType NotificationType, title, message string, data map[string]interface{}) (*Notification, error) {
	notification := &Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		IsRead:  false,
	}

	if data != nil {
		if err := notification.SetNotificationData(data); err != nil {
			return nil, err
		}
	}

	return notification, nil
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}

// SSEMessage represents a message sent through SSE
type SSEMessage struct {
	ID        string           `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Data      interface{}      `json:"data,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

// ToSSEMessage converts a notification to an SSE message
func (n *Notification) ToSSEMessage() *SSEMessage {
	return &SSEMessage{
		ID:        string(rune(n.ID)),
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Data:      n.NotificationData(),
		Timestamp: n.CreatedAt,
	}
}
