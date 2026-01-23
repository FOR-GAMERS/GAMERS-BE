package exception

var (
	ErrNotificationNotFound = &BusinessError{
		Status:  404,
		Code:    "NOTIFICATION_NOT_FOUND",
		Message: "Notification not found",
	}

	ErrNotificationSendFailed = &BusinessError{
		Status:  500,
		Code:    "NOTIFICATION_SEND_FAILED",
		Message: "Failed to send notification",
	}

	ErrSSEConnectionFailed = &BusinessError{
		Status:  500,
		Code:    "SSE_CONNECTION_FAILED",
		Message: "Failed to establish SSE connection",
	}
)
