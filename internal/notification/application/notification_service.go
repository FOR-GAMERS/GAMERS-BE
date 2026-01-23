package application

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/notification/application/dto"
	"GAMERS-BE/internal/notification/application/port"
	"GAMERS-BE/internal/notification/domain"
	"fmt"
	"log"
	"strconv"
	"time"
)

// NotificationService handles notification business logic
type NotificationService struct {
	databasePort port.NotificationDatabasePort
	sseManager   *SSEManager
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	databasePort port.NotificationDatabasePort,
	sseManager *SSEManager,
) *NotificationService {
	return &NotificationService{
		databasePort: databasePort,
		sseManager:   sseManager,
	}
}

// CreateAndSendNotification creates a notification and sends it via SSE
func (s *NotificationService) CreateAndSendNotification(
	userID int64,
	notifType domain.NotificationType,
	title, message string,
	data map[string]interface{},
) error {
	// Create notification
	notification, err := domain.NewNotification(userID, notifType, title, message, data)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Save to database
	if err := s.databasePort.Save(notification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	// Send via SSE if user is connected
	sseMessage := &domain.SSEMessage{
		ID:        strconv.FormatInt(notification.ID, 10),
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Data:      notification.NotificationData(),
		Timestamp: notification.CreatedAt,
	}

	if err := s.sseManager.SendToUser(userID, sseMessage); err != nil {
		log.Printf("Failed to send SSE notification to user %d: %v", userID, err)
		// Don't return error - notification is saved, SSE is best effort
	}

	return nil
}

// GetNotifications returns notifications for a user with pagination
func (s *NotificationService) GetNotifications(userID int64, req *dto.GetNotificationsRequest) (*dto.NotificationListResponse, error) {
	// Set defaults
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	var notifications []*domain.Notification
	var err error

	if req.Unread {
		notifications, err = s.databasePort.FindUnreadByUserID(userID, limit, req.Offset)
	} else {
		notifications, err = s.databasePort.FindByUserID(userID, limit, req.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Get counts
	total, err := s.databasePort.CountByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	unreadCount, err := s.databasePort.CountUnreadByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return &dto.NotificationListResponse{
		Notifications: dto.FromNotifications(notifications),
		UnreadCount:   unreadCount,
		Total:         total,
	}, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(userID, notificationID int64) error {
	// Verify notification belongs to user
	notification, err := s.databasePort.FindByID(notificationID)
	if err != nil {
		return exception.ErrNotificationNotFound
	}

	if notification.UserID != userID {
		return exception.ErrNotificationNotFound
	}

	return s.databasePort.MarkAsRead(notificationID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID int64) error {
	return s.databasePort.MarkAllAsRead(userID)
}

// GetSSEManager returns the SSE manager
func (s *NotificationService) GetSSEManager() *SSEManager {
	return s.sseManager
}

// HandleTeamInviteReceived handles team invite received event
func (s *NotificationService) HandleTeamInviteReceived(inviteeUserID int64, inviterUsername, teamName string, gameID, contestID int64) error {
	data := map[string]interface{}{
		"inviter_username": inviterUsername,
		"team_name":        teamName,
		"game_id":          gameID,
		"contest_id":       contestID,
	}

	title := "팀 초대"
	message := fmt.Sprintf("%s님이 %s 팀에 초대했습니다.", inviterUsername, teamName)

	return s.CreateAndSendNotification(inviteeUserID, domain.NotificationTypeTeamInviteReceived, title, message, data)
}

// HandleTeamInviteAccepted handles team invite accepted event
func (s *NotificationService) HandleTeamInviteAccepted(inviterUserID int64, inviteeUsername, teamName string, gameID, contestID int64) error {
	data := map[string]interface{}{
		"invitee_username": inviteeUsername,
		"team_name":        teamName,
		"game_id":          gameID,
		"contest_id":       contestID,
	}

	title := "초대 수락됨"
	message := fmt.Sprintf("%s님이 팀 초대를 수락했습니다.", inviteeUsername)

	return s.CreateAndSendNotification(inviterUserID, domain.NotificationTypeTeamInviteAccepted, title, message, data)
}

// HandleTeamInviteRejected handles team invite rejected event
func (s *NotificationService) HandleTeamInviteRejected(inviterUserID int64, inviteeUsername, teamName string, gameID, contestID int64) error {
	data := map[string]interface{}{
		"invitee_username": inviteeUsername,
		"team_name":        teamName,
		"game_id":          gameID,
		"contest_id":       contestID,
	}

	title := "초대 거절됨"
	message := fmt.Sprintf("%s님이 팀 초대를 거절했습니다.", inviteeUsername)

	return s.CreateAndSendNotification(inviterUserID, domain.NotificationTypeTeamInviteRejected, title, message, data)
}

// HandleApplicationAccepted handles contest application accepted event
func (s *NotificationService) HandleApplicationAccepted(userID, contestID int64, contestTitle string) error {
	data := map[string]interface{}{
		"contest_id":    contestID,
		"contest_title": contestTitle,
	}

	title := "참가 신청 승인"
	message := fmt.Sprintf("%s 대회 참가 신청이 승인되었습니다.", contestTitle)

	return s.CreateAndSendNotification(userID, domain.NotificationTypeApplicationAccepted, title, message, data)
}

// HandleApplicationRejected handles contest application rejected event
func (s *NotificationService) HandleApplicationRejected(userID, contestID int64, contestTitle, reason string) error {
	data := map[string]interface{}{
		"contest_id":    contestID,
		"contest_title": contestTitle,
		"reason":        reason,
	}

	title := "참가 신청 거절"
	message := fmt.Sprintf("%s 대회 참가 신청이 거절되었습니다.", contestTitle)
	if reason != "" {
		message += fmt.Sprintf(" 사유: %s", reason)
	}

	return s.CreateAndSendNotification(userID, domain.NotificationTypeApplicationRejected, title, message, data)
}

// CleanupOldNotifications removes old notifications
func (s *NotificationService) CleanupOldNotifications(days int) error {
	return s.databasePort.DeleteOldNotifications(days)
}

// SendTestNotification sends a test notification (for debugging)
func (s *NotificationService) SendTestNotification(userID int64) error {
	return s.CreateAndSendNotification(
		userID,
		domain.NotificationTypeTeamInviteReceived,
		"테스트 알림",
		"SSE 알림 테스트 메시지입니다.",
		map[string]interface{}{
			"test":      true,
			"timestamp": time.Now().Unix(),
		},
	)
}
