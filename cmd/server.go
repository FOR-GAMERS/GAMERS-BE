package main

import (
	"context"
	"log"
	"os"

	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth"
	authMiddleware "github.com/FOR-GAMERS/GAMERS-BE/internal/auth/middleware"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/banner"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/game"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/config"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/middleware"
	authProvider "github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/notification"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/valorant"

	_ "github.com/FOR-GAMERS/GAMERS-BE/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func init() {
	if os.Getenv("GIN_MODE") == "release" {
		log.Println("Production mode detected. Using system environment variables.")
		return
	}

	// Load .env file from env directory
	if err := godotenv.Load("env/.env"); err != nil {
		log.Println("No env/.env file found, using system environment variables")
	}
}

// @title GAMERS API
// @version 1.0
// @description API Server for GAMERS platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@gamers.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	db := initDatabase()
	redisClient := initRedis()
	rabbitmqConn := initRabbitMQ()

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Fatal("Failed to close Redis client:", err)
		}

		if err := rabbitmqConn.Close(); err != nil {
			log.Fatal("Failed to close RabbitMQ connection:", err)
		}

		sqlDB, _ := db.DB()
		if sqlDB != nil {
			err := sqlDB.Close()
			if err != nil {
				log.Fatal("Database is not ended normally")
				return
			}
		}
	}()

	ctx := context.Background()

	tokenService := authProvider.ProvideJwtService()
	authInterceptor := authMiddleware.NewAuthMiddleware(tokenService)

	appRouter := router.NewRouter(authInterceptor)

	authDeps := auth.ProvideAuthDependencies(db, redisClient, &ctx, appRouter)

	// Discord module - provides Discord API integration (must be initialized before OAuth2)
	discordDeps := discord.ProvideDiscordDependencies(appRouter, db, redisClient, &ctx)

	// OAuth2 module - uses Discord token port for storing Discord tokens
	oauth2Deps := oauth2.ProvideOAuth2Dependencies(db, redisClient, &ctx, appRouter, discordDeps.DiscordTokenPort)

	// User module - uses OAuth2 repository for Discord avatar URL generation
	userDeps := user.ProvideUserDependencies(db, appRouter, oauth2Deps.OAuth2Repository)

	// Set user query port for admin middleware after user dependencies are initialized
	authInterceptor.SetUserQueryPort(userDeps.UserQueryRepo)

	// Game module - provides Game, Team, and GameTeam management
	gameDeps := game.ProvideGameDependencies(
		db,
		redisClient,
		rabbitmqConn,
		appRouter,
		nil, // Contest repository will be set after contest initialization
		oauth2Deps.OAuth2Repository,
		userDeps.UserQueryRepo,
	)

	// Contest module with full features (Discord validation + Tournament generation)
	contestDeps := contest.ProvideContestDependenciesFull(
		db,
		redisClient,
		rabbitmqConn,
		appRouter,
		oauth2Deps.OAuth2Repository,
		userDeps.UserQueryRepo,
		discordDeps.ValidationService,
		gameDeps.GameRepository,
		gameDeps.TeamRepository,
		gameDeps.GameTeamRepository,
	)

	// Set contest repository for team service and tournament result service (to resolve circular dependency)
	gameDeps.TeamService.SetContestRepository(contestDeps.ContestRepository)
	gameDeps.TournamentResultService.SetContestDBPort(contestDeps.ContestRepository)

	commentDeps := comment.ProvideCommentDependencies(db, appRouter, contestDeps.ContestRepository)

	// Point module - provides Valorant score table management
	pointDeps := point.ProvidePointDependencies(db, appRouter)

	// Valorant module - provides Valorant MMR/Rank integration
	valorantDeps := valorant.ProvideValorantDependencies(
		appRouter,
		userDeps.UserQueryRepo,
		userDeps.UserCommandRepo,
		pointDeps.ScoreTableRepository,
	)

	// Storage module - provides R2 storage integration for images
	storageDeps := storage.ProvideStorageDependencies(appRouter)

	// Banner module - provides main banner management for homepage
	bannerDeps := banner.ProvideBannerDependencies(db, appRouter)

	// Notification module - provides SSE real-time notifications
	notificationDeps := notification.ProvideNotificationDependencies(db, appRouter)

	// Wire notification handler to contest and game services
	contestDeps.ApplicationService.SetNotificationHandler(notificationDeps.Service)
	gameDeps.TeamService.SetNotificationHandler(notificationDeps.Service)

	// Start Team Persistence Consumer for Write-Behind pattern
	startTeamPersistenceConsumer(ctx, gameDeps)

	setupRouter(appRouter, authDeps, userDeps, oauth2Deps, contestDeps, commentDeps, discordDeps, gameDeps, pointDeps, valorantDeps, storageDeps, bannerDeps, notificationDeps)

	startServer(appRouter.Engine())
}

func startServer(engine interface{}) {
	log.Println("===========================================")
	log.Println("🎮 GAMERS API Server Starting")
	log.Println("===========================================")
	log.Println("Server:          http://localhost:8080")
	log.Println("Health Check:    http://localhost:8080/health")
	log.Println("Swagger UI:      http://localhost:8080/swagger/index.html")
	log.Println("===========================================")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if ginEngine, ok := engine.(*gin.Engine); ok {
		if err := ginEngine.Run(":" + port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}
}

func setupRouter(
	appRouter *router.Router,
	authDeps *auth.Dependencies,
	userDeps *user.Dependencies,
	oauth2Deps *oauth2.Dependencies,
	contestDeps *contest.Dependencies,
	commentDeps *comment.Dependencies,
	discordDeps *discord.Dependencies,
	gameDeps *game.Dependencies,
	pointDeps *point.Dependencies,
	valorantDeps *valorant.Dependencies,
	storageDeps *storage.Dependencies,
	bannerDeps *banner.Dependencies,
	notificationDeps *notification.Dependencies,
) *router.Router {

	appRouter.Engine().Use(middleware.GlobalErrorHandler())

	appRouter.RegisterHealthCheck()
	appRouter.RegisterSwagger(ginSwagger.WrapHandler(swaggerFiles.Handler))

	authDeps.Controller.RegisterRoutes()
	userDeps.Controller.RegisterRoutes()
	oauth2Deps.Controller.RegisterRoutes()
	contestDeps.Controller.RegisterRoute()
	contestDeps.ApplicationController.RegisterRoute()
	commentDeps.Controller.RegisterRoutes()
	// discordDeps.Controller routes are registered in the constructor
	gameDeps.GameController.RegisterRoutes()
	gameDeps.TeamController.RegisterRoutes()
	gameDeps.GameTeamController.RegisterRoutes()
	pointDeps.ValorantController.RegisterRoutes()
	valorantDeps.Controller.RegisterRoutes()
	if storageDeps != nil {
		storageDeps.Controller.RegisterRoutes()
	}
	if bannerDeps != nil {
		bannerDeps.Controller.RegisterRoutes()
	}
	// Notification routes are registered in provider
	_ = notificationDeps

	return appRouter
}

func initDatabase() *gorm.DB {
	dbConfig := config.NewConfigFromEnv()
	db, err := config.InitDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize config:", err)
	}

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		log.Println("🔄 Running config migrations...")
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("Failed to get SQL DB:", err)
		}

		migrationsPath := os.Getenv("MIGRATIONS_PATH")
		if migrationsPath == "" {
			migrationsPath = "./db/migrations"
		}

		if err := config.RunMigrations(sqlDB, migrationsPath); err != nil {
			log.Fatal("Failed to run migrations:", err)
		}
	}

	return db
}

func initRedis() *redis.Client {
	redisConfig := config.NewRedisConfigFromEnv()
	redisClient, err := config.InitRedis(redisConfig)

	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	return redisClient
}

func initRabbitMQ() *config.RabbitMQConnection {
	rabbitmqConfig := config.NewRabbitMQConfigFromEnv()
	rabbitmqConn, err := config.InitRabbitMQ(rabbitmqConfig)

	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}
	return rabbitmqConn
}

// startTeamPersistenceConsumer starts the consumer for Write-Behind pattern
func startTeamPersistenceConsumer(ctx context.Context, gameDeps *game.Dependencies) {
	if gameDeps.TeamPersistenceConsumer == nil || gameDeps.TeamPersistenceHandler == nil {
		log.Println("Team persistence consumer not initialized, skipping...")
		return
	}

	go func() {
		log.Println("🔄 Starting Team Persistence Consumer (Write-Behind pattern)...")
		if err := gameDeps.TeamPersistenceConsumer.Start(ctx, gameDeps.TeamPersistenceHandler.HandleTeamPersistence); err != nil {
			log.Printf("Failed to start team persistence consumer: %v", err)
		}
	}()
}
