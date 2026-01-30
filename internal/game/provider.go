package game

import (
	contestPort "GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/game/application"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/infra/persistence/adapter"
	"GAMERS-BE/internal/game/presentation"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/config"
	"os"

	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	GameController            *presentation.GameController
	TeamController            *presentation.TeamController
	GameTeamController        *presentation.GameTeamController
	GameRepository            port.GameDatabasePort
	TeamRepository            port.TeamDatabasePort
	GameTeamRepository        port.GameTeamDatabasePort
	TeamService               *application.TeamService
	TeamPersistenceConsumer   port.TeamPersistenceConsumerPort
	TeamPersistenceHandler    *application.TeamPersistenceHandler
	GameSchedulerService      *application.GameSchedulerService
	MatchDetectionService     *application.MatchDetectionService
	TournamentResultService   *application.TournamentResultService
}

func ProvideGameDependencies(
	db *gorm.DB,
	redisClient *redis.Client,
	rabbitmqConn *config.RabbitMQConnection,
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
	matchResultDatabaseAdapter := adapter.NewMatchResultDatabaseAdapter(db)

	// Redis Adapter for Team
	teamRedisAdapter := adapter.NewTeamRedisAdapter(redisClient)

	// RabbitMQ Event Publisher for Team
	teamEventPublisher := adapter.NewTeamEventPublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// RabbitMQ Persistence Publisher for Write-Behind pattern
	teamPersistencePublisher := adapter.NewTeamPersistencePublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// RabbitMQ Persistence Consumer for Write-Behind pattern
	teamPersistenceConsumer := adapter.NewTeamPersistenceConsumerRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// Persistence Handler for DB operations
	teamPersistenceHandler := application.NewTeamPersistenceHandler(teamDatabaseAdapter)

	// Game Event Publisher (RabbitMQ)
	gameEventPublisher := adapter.NewGameEventPublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// Valorant API Adapter for match detection
	valorantAPIKey := os.Getenv("VALORANT_API_KEY")
	matchDetectionAdapter := adapter.NewMatchDetectionValorantAdapter(valorantAPIKey)

	// Services
	gameService := application.NewGameService(
		gameDatabaseAdapter,
		teamDatabaseAdapter,
	)

	teamService := application.NewTeamService(
		teamDatabaseAdapter,
		teamRedisAdapter,
		contestRepository,
		oauth2Repository,
		userQueryRepo,
		teamEventPublisher,
		teamPersistencePublisher,
	)

	gameTeamService := application.NewGameTeamService(
		gameTeamDatabaseAdapter,
		gameDatabaseAdapter,
		teamDatabaseAdapter,
		contestRepository,
	)

	// Match Detection Service
	matchDetectionService := application.NewMatchDetectionService(
		matchDetectionAdapter,
		gameDatabaseAdapter,
		gameTeamDatabaseAdapter,
		teamDatabaseAdapter,
		matchResultDatabaseAdapter,
		gameEventPublisher,
		userQueryRepo,
	)

	// Game Scheduler Service (with Redis distributed lock)
	gameSchedulerService := application.NewGameSchedulerService(
		gameDatabaseAdapter,
		matchDetectionService,
		gameEventPublisher,
		redisClient,
	)

	// Tournament Result Service
	tournamentResultService := application.NewTournamentResultService(
		gameDatabaseAdapter,
		gameTeamDatabaseAdapter,
		teamDatabaseAdapter,
		matchResultDatabaseAdapter,
		contestRepository,
	)

	// Controllers
	gameController := presentation.NewGameController(
		router,
		gameService,
		matchDetectionService,
		tournamentResultService,
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
		GameController:          gameController,
		TeamController:          teamController,
		GameTeamController:      gameTeamController,
		GameRepository:          gameDatabaseAdapter,
		TeamRepository:          teamDatabaseAdapter,
		GameTeamRepository:      gameTeamDatabaseAdapter,
		TeamService:             teamService,
		TeamPersistenceConsumer: teamPersistenceConsumer,
		TeamPersistenceHandler:  teamPersistenceHandler,
		GameSchedulerService:    gameSchedulerService,
		MatchDetectionService:   matchDetectionService,
		TournamentResultService: tournamentResultService,
	}
}
