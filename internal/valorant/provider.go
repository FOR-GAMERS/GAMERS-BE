package valorant

import (
	"GAMERS-BE/internal/global/common/router"
	pointPort "GAMERS-BE/internal/point/application/port"
	userCommandPort "GAMERS-BE/internal/user/application/port/command"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/valorant/application"
	"GAMERS-BE/internal/valorant/infra"
	"GAMERS-BE/internal/valorant/presentation"
)

type Dependencies struct {
	Controller *presentation.ValorantUserController
}

func ProvideValorantDependencies(
	router *router.Router,
	userQueryRepo userQueryPort.UserQueryPort,
	userCommandRepo userCommandPort.UserCommandPort,
	scoreTableRepo pointPort.ValorantScoreTableDatabasePort,
) *Dependencies {
	// Infrastructure
	valorantApiClient := infra.NewValorantApiClient()

	// Services
	valorantUserService := application.NewValorantUserService(
		valorantApiClient,
		userQueryRepo,
		userCommandRepo,
		scoreTableRepo,
	)

	// Controllers
	valorantUserController := presentation.NewValorantUserController(
		router,
		valorantUserService,
	)

	return &Dependencies{
		Controller: valorantUserController,
	}
}
