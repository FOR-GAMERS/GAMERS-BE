package contest

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/infra/persistence/adapter"
	"GAMERS-BE/internal/contest/presentation"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/database"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Controller            *presentation.ContestController
	ApplicationController *presentation.ContestApplicationController
}

func ProvideContestDependencies(
	db *gorm.DB,
	redisClient *redis.Client,
	rabbitmqConn *database.RabbitMQConnection,
	router *router.Router,
	oauth2Repository oauth2Port.OAuth2DatabasePort,
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
	)
	contestApplicationController := presentation.NewContestApplicationController(
		router,
		contestApplicationService,
		controllerHelper,
	)

	return &Dependencies{
		Controller:            contestController,
		ApplicationController: contestApplicationController,
	}
}
