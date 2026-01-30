package notification

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification/presentation"

	"gorm.io/gorm"
)

// Dependencies holds all notification domain dependencies
type Dependencies struct {
	Controller          *presentation.NotificationController
	Service             *application.NotificationService
	SSEManager          *application.SSEManager
	NotificationAdapter *adapter.NotificationDatabaseAdapter
}

// ProvideNotificationDependencies creates and wires all notification dependencies
func ProvideNotificationDependencies(db *gorm.DB, router *router.Router) *Dependencies {
	// Create adapters
	notificationAdapter := adapter.NewNotificationDatabaseAdapter(db)

	// Create SSE manager
	sseManager := application.NewSSEManager()

	// Create service
	notificationService := application.NewNotificationService(notificationAdapter, sseManager)

	// Create controller
	notificationController := presentation.NewNotificationController(router, notificationService)

	// Register routes
	notificationController.RegisterRoutes()

	return &Dependencies{
		Controller:          notificationController,
		Service:             notificationService,
		SSEManager:          sseManager,
		NotificationAdapter: notificationAdapter,
	}
}
