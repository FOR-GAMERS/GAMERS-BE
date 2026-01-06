package contest

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/infra/persistence/adapter"
	"GAMERS-BE/internal/contest/presentation"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"

	"gorm.io/gorm"
)

type Dependencies struct {
	Controller *presentation.ContestController
}

func ProvideContestDependencies(db *gorm.DB, router *router.Router) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	contestDatabaseAdapter := adapter.NewContestDatabaseAdapter(db)
	contestService := application.NewContestService(contestDatabaseAdapter)
	contestController := presentation.NewContestController(router, contestService, controllerHelper)

	return &Dependencies{
		Controller: contestController,
	}
}
