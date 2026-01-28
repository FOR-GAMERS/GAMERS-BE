package valorant

import (
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/utils"
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
	apiKey := utils.GetEnv("HENRIK_RIOT_API_KEY", "")
	valorantApiClient := infra.NewValorantApiClient(apiKey)

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
