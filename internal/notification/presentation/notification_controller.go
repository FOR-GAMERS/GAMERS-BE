package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/middleware"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/infra/sse"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// swagger type alias
var _ = dto.NotificationResponse{}
var _ = dto.NotificationListResponse{}

// NotificationController handles notification-related HTTP requests
type NotificationController struct {
	router  *router.Router
	service *application.NotificationService
}

// NewNotificationController creates a new notification controller
func NewNotificationController(router *router.Router, service *application.NotificationService) *NotificationController {
	return &NotificationController{
		router:  router,
		service: service,
	}
}

// RegisterRoutes registers notification routes
func (c *NotificationController) RegisterRoutes() {
	group := c.router.ProtectedGroup("/api/notifications")
	group.GET("/stream", c.SSEStream)
	group.GET("", c.GetNotifications)
	group.PATCH("/:id/read", c.MarkAsRead)
	group.PATCH("/read-all", c.MarkAllAsRead)
}

// SSEStream godoc
// @Summary Subscribe to real-time notifications
// @Description Establishes an SSE connection for receiving real-time notifications
// @Tags notifications
// @Produce text/event-stream
// @Security BearerAuth
// @Success 200 {string} string "SSE stream"
// @Failure 401 {object} response.Response
// @Router /api/notifications/stream [get]
func (c *NotificationController) SSEStream(ctx *gin.Context) {
	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	// Set SSE headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("X-Accel-Buffering", "no")

	// Create SSE client
	client := sse.NewSSEClient(userID, ctx.Writer)
	manager := c.service.GetSSEManager()

	// Register client
	manager.RegisterClient(client)
	defer manager.UnregisterClient(client)

	// Send initial connection message
	ctx.SSEvent("connected", map[string]interface{}{
		"message":   "Connected to notification stream",
		"user_id":   userID,
		"timestamp": time.Now().Unix(),
	})
	ctx.Writer.Flush()

	log.Printf("SSE: User %d connected to notification stream", userID)

	// Keep connection alive with heartbeat
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Listen for client disconnect
	clientGone := ctx.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			log.Printf("SSE: User %d disconnected from notification stream", userID)
			return
		case <-heartbeatTicker.C:
			if err := client.SendHeartbeat(); err != nil {
				log.Printf("SSE: Failed to send heartbeat to user %d: %v", userID, err)
				return
			}
		}
	}
}

// GetNotifications godoc
// @Summary Get notifications
// @Description Retrieves notifications for the authenticated user with pagination
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of notifications to return (default: 20, max: 100)"
// @Param offset query int false "Offset for pagination"
// @Param unread query bool false "Filter only unread notifications"
// @Success 200 {object} response.Response{data=dto.NotificationListResponse}
// @Failure 401 {object} response.Response
// @Router /api/notifications [get]
func (c *NotificationController) GetNotifications(ctx *gin.Context) {
	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	var req dto.GetNotificationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.JSON(ctx, response.BadRequest("invalid query parameters"))
		return
	}

	result, err := c.service.GetNotifications(userID, &req)
	if err != nil {
		response.JSON(ctx, response.InternalServerError(err.Error()))
		return
	}

	response.JSON(ctx, response.Success(result, "notifications retrieved successfully"))
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Marks a specific notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/notifications/{id}/read [patch]
func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	idStr := ctx.Param("id")
	notificationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid notification id"))
		return
	}

	if err := c.service.MarkAsRead(userID, notificationID); err != nil {
		response.JSON(ctx, response.NotFound("notification not found"))
		return
	}

	response.JSON(ctx, response.Success[any](nil, "notification marked as read"))
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Marks all notifications as read for the authenticated user
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/notifications/read-all [patch]
func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	if err := c.service.MarkAllAsRead(userID); err != nil {
		response.JSON(ctx, response.InternalServerError(err.Error()))
		return
	}

	response.JSON(ctx, response.Success[any](nil, "all notifications marked as read"))
}
