package storage

import (
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/storage/application"
	"GAMERS-BE/internal/storage/infra"
	"GAMERS-BE/internal/storage/presentation"
	"log"
)

type Dependencies struct {
	Controller     *presentation.StorageController
	StorageService *application.StorageService
}

func ProvideStorageDependencies(router *router.Router) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	r2Config := infra.NewR2ConfigFromEnv()
	storageAdapter, err := infra.NewR2StorageAdapter(r2Config)
	if err != nil {
		log.Printf("Warning: Failed to initialize R2 storage adapter: %v", err)
		log.Println("Storage endpoints will not be available")
		return nil
	}

	storageService := application.NewStorageService(storageAdapter)
	storageController := presentation.NewStorageController(router, storageService, controllerHelper)

	return &Dependencies{
		Controller:     storageController,
		StorageService: storageService,
	}
}
