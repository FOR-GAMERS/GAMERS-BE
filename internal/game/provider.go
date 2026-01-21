package game

import (
	contestPort "GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/game/application"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/infra/persistence/adapter"
	"GAMERS-BE/internal/game/presentation"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/database"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	GameController     *presentation.GameController
	TeamController     *presentation.TeamController
	GameTeamController *presentation.GameTeamController
	GameRepository     port.GameDatabasePort
	TeamRepository     port.TeamDatabasePort
}

func ProvideGameDependencies(
	db *gorm.DB,
	redisClient *redis.Client,
	rabbitmqConn *database.RabbitMQConnection,
	router *router.Router,
	contestRepository contestPort.ContestDatabasePort,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	// Database Adapters
	gameDatabaseAdapter := adapter.NewGameDatabaseAdapter(db)
	teamDatabaseAdapter := adapter.NewTeamDatabaseAdapter(db)
	gameTeamDatabaseAdapter := adapter.NewGameTeamDatabaseAdapter(db)

	// Redis Adapter for Team
	teamRedisAdapter := adapter.NewTeamRedisAdapter(redisClient)

	// RabbitMQ Event Publisher for Team
	teamEventPublisher := adapter.NewTeamEventPublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// Services
	gameService := application.NewGameService(
		gameDatabaseAdapter,
		teamDatabaseAdapter,
	)

	teamService := application.NewTeamService(
		gameDatabaseAdapter,
		teamDatabaseAdapter,
		teamRedisAdapter,
		contestRepository,
		oauth2Repository,
		userQueryRepo,
		teamEventPublisher,
	)

	gameTeamService := application.NewGameTeamService(
		gameTeamDatabaseAdapter,
		gameDatabaseAdapter,
		teamDatabaseAdapter,
		contestRepository,
	)

	// Controllers
	gameController := presentation.NewGameController(
		router,
		gameService,
		controllerHelper,
	)

	teamController := presentation.NewTeamController(
		router,
		teamService,
		controllerHelper,
	)

	gameTeamController := presentation.NewGameTeamController(
		router,
		gameTeamService,
		controllerHelper,
	)

	return &Dependencies{
		GameController:     gameController,
		TeamController:     teamController,
		GameTeamController: gameTeamController,
		GameRepository:     gameDatabaseAdapter,
		TeamRepository:     teamDatabaseAdapter,
	}
}
