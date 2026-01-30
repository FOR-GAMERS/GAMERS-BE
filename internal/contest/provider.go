package contest

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/application/port"
	contestAdapter "GAMERS-BE/internal/contest/infra/adapter"
	"GAMERS-BE/internal/contest/infra/persistence/adapter"
	"GAMERS-BE/internal/contest/presentation"
	discordApplication "GAMERS-BE/internal/discord/application"
	gameApplication "GAMERS-BE/internal/game/application"
	gamePort "GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/config"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Controller            *presentation.ContestController
	ApplicationController *presentation.ContestApplicationController
	ContestRepository     port.ContestDatabasePort
	ContestService        *application.ContestService
	ApplicationService    *application.ContestApplicationService
}

func ProvideContestDependencies(
	db *gorm.DB,
	redisClient *redis.Client,
	rabbitmqConn *config.RabbitMQConnection,
	router *router.Router,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	// Contest 관련
	contestDatabaseAdapter := adapter.NewContestDatabaseAdapter(db)
	contestMemberDatabaseAdapter := adapter.NewContestMemberDatabaseAdapter(db)
	contestApplicationRedisAdapter := adapter.NewContestApplicationRedisAdapter(redisClient)

	// RabbitMQ Event Publisher
	eventPublisher := adapter.NewEventPublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	contestService := application.NewContestService(
		contestDatabaseAdapter,
		contestMemberDatabaseAdapter,
		contestApplicationRedisAdapter,
		oauth2Repository,
		eventPublisher,
	)
	contestController := presentation.NewContestController(router, contestService, controllerHelper)

	// Contest Application 관련
	contestApplicationService := application.NewContestApplicationService(
		contestApplicationRedisAdapter,
		contestDatabaseAdapter,
		contestMemberDatabaseAdapter,
		eventPublisher,
		oauth2Repository,
		userQueryRepo,
	)
	contestApplicationController := presentation.NewContestApplicationController(
		router,
		contestApplicationService,
		controllerHelper,
	)

	return &Dependencies{
		Controller:            contestController,
		ApplicationController: contestApplicationController,
		ContestRepository:     contestDatabaseAdapter,
		ContestService:        contestService,
		ApplicationService:    contestApplicationService,
	}
}

// ProvideContestDependenciesWithDiscord creates contest dependencies with Discord validation
func ProvideContestDependenciesWithDiscord(
	db *gorm.DB,
	redisClient *redis.Client,
	rabbitmqConn *config.RabbitMQConnection,
	router *router.Router,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
	discordValidationService *discordApplication.DiscordValidationService,
) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	// Contest 관련
	contestDatabaseAdapter := adapter.NewContestDatabaseAdapter(db)
	contestMemberDatabaseAdapter := adapter.NewContestMemberDatabaseAdapter(db)
	contestApplicationRedisAdapter := adapter.NewContestApplicationRedisAdapter(redisClient)

	// RabbitMQ Event Publisher
	eventPublisher := adapter.NewEventPublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// Discord Validation Adapter
	discordValidationAdapter := contestAdapter.NewDiscordValidationAdapter(discordValidationService)

	contestService := application.NewContestServiceWithDiscord(
		contestDatabaseAdapter,
		contestMemberDatabaseAdapter,
		contestApplicationRedisAdapter,
		oauth2Repository,
		eventPublisher,
		discordValidationAdapter,
	)
	contestController := presentation.NewContestController(router, contestService, controllerHelper)

	// Contest Application 관련
	contestApplicationService := application.NewContestApplicationService(
		contestApplicationRedisAdapter,
		contestDatabaseAdapter,
		contestMemberDatabaseAdapter,
		eventPublisher,
		oauth2Repository,
		userQueryRepo,
	)
	contestApplicationController := presentation.NewContestApplicationController(
		router,
		contestApplicationService,
		controllerHelper,
	)

	return &Dependencies{
		Controller:            contestController,
		ApplicationController: contestApplicationController,
		ContestRepository:     contestDatabaseAdapter,
		ContestService:        contestService,
		ApplicationService:    contestApplicationService,
	}
}

// ProvideContestDependenciesFull creates contest dependencies with all features
func ProvideContestDependenciesFull(
	db *gorm.DB,
	redisClient *redis.Client,
	rabbitmqConn *config.RabbitMQConnection,
	router *router.Router,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
	userQueryRepo userQueryPort.UserQueryPort,
	discordValidationService *discordApplication.DiscordValidationService,
	gameRepository gamePort.GameDatabasePort,
	teamRepository gamePort.TeamDatabasePort,
	gameTeamRepository gamePort.GameTeamDatabasePort,
) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	// Contest 관련
	contestDatabaseAdapter := adapter.NewContestDatabaseAdapter(db)
	contestMemberDatabaseAdapter := adapter.NewContestMemberDatabaseAdapter(db)
	contestApplicationRedisAdapter := adapter.NewContestApplicationRedisAdapter(redisClient)

	// RabbitMQ Event Publisher
	eventPublisher := adapter.NewEventPublisherRabbitMQAdapter(
		rabbitmqConn,
		rabbitmqConn.Config().Exchange,
	)

	// Discord Validation Adapter
	discordValidationAdapter := contestAdapter.NewDiscordValidationAdapter(discordValidationService)

	// Tournament Service
	tournamentService := gameApplication.NewTournamentService(gameRepository, teamRepository)

	contestService := application.NewContestServiceFull(
		contestDatabaseAdapter,
		contestMemberDatabaseAdapter,
		contestApplicationRedisAdapter,
		oauth2Repository,
		eventPublisher,
		discordValidationAdapter,
		tournamentService,
		teamRepository,
		gameTeamRepository,
	)
	contestController := presentation.NewContestController(router, contestService, controllerHelper)

	// Contest Application 관련
	contestApplicationService := application.NewContestApplicationService(
		contestApplicationRedisAdapter,
		contestDatabaseAdapter,
		contestMemberDatabaseAdapter,
		eventPublisher,
		oauth2Repository,
		userQueryRepo,
	)
	contestApplicationController := presentation.NewContestApplicationController(
		router,
		contestApplicationService,
		controllerHelper,
	)

	return &Dependencies{
		Controller:            contestController,
		ApplicationController: contestApplicationController,
		ContestRepository:     contestDatabaseAdapter,
		ContestService:        contestService,
		ApplicationService:    contestApplicationService,
	}
}
