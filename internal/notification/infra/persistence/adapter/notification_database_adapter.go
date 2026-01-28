package adapter

import (
	"GAMERS-BE/internal/notification/domain"
	"time"

	"gorm.io/gorm"
)

// NotificationDatabaseAdapter implements the NotificationDatabasePort interface
type NotificationDatabaseAdapter struct {
	db *gorm.DB
}

// NewNotificationDatabaseAdapter creates a new notification config adapter
func NewNotificationDatabaseAdapter(db *gorm.DB) *NotificationDatabaseAdapter {
	return &NotificationDatabaseAdapter{db: db}
}

// Save saves a new notification to the config
func (a *NotificationDatabaseAdapter) Save(notification *domain.Notification) error {
	return a.db.Create(notification).Error
}

// FindByID finds a notification by ID
func (a *NotificationDatabaseAdapter) FindByID(id int64) (*domain.Notification, error) {
	var notification domain.Notification
	if err := a.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// FindByUserID finds all notifications for a user with pagination
func (a *NotificationDatabaseAdapter) FindByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := a.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// FindUnreadByUserID finds unread notifications for a user
func (a *NotificationDatabaseAdapter) FindUnreadByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	var notifications []*domain.Notification
	err := a.db.Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// CountByUserID counts total notifications for a user
func (a *NotificationDatabaseAdapter) CountByUserID(userID int64) (int64, error) {
	var count int64
	err := a.db.Model(&domain.Notification{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountUnreadByUserID counts unread notifications for a user
func (a *NotificationDatabaseAdapter) CountUnreadByUserID(userID int64) (int64, error) {
	var count int64
	err := a.db.Model(&domain.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read
func (a *NotificationDatabaseAdapter) MarkAsRead(id int64) error {
	return a.db.Model(&domain.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}

// MarkAllAsRead marks all notifications as read for a user
func (a *NotificationDatabaseAdapter) MarkAllAsRead(userID int64) error {
	return a.db.Model(&domain.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

// Delete deletes a notification
func (a *NotificationDatabaseAdapter) Delete(id int64) error {
	return a.db.Delete(&domain.Notification{}, id).Error
}

// DeleteOldNotifications deletes notifications older than a certain number of days
func (a *NotificationDatabaseAdapter) DeleteOldNotifications(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return a.db.Where("created_at < ?", cutoff).Delete(&domain.Notification{}).Error
}
