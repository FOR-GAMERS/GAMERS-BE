package banner

import (
	"GAMERS-BE/internal/banner/application"
	"GAMERS-BE/internal/banner/infra/persistence/adapter"
	"GAMERS-BE/internal/banner/presentation"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"

	"gorm.io/gorm"
)

type Dependencies struct {
	Controller    *presentation.BannerController
	BannerService *application.BannerService
}

func ProvideBannerDependencies(db *gorm.DB, router *router.Router) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	bannerAdapter := adapter.NewMySQLBannerAdapter(db)
	bannerService := application.NewBannerService(bannerAdapter)
	bannerController := presentation.NewBannerController(router, bannerService, controllerHelper)

	return &Dependencies{
		Controller:    bannerController,
		BannerService: bannerService,
	}
}
