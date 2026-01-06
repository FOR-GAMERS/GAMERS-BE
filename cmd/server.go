package main

import (
	"GAMERS-BE/internal/auth"
	authMiddleware "GAMERS-BE/internal/auth/middleware"
	"GAMERS-BE/internal/contest"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/database"
	"GAMERS-BE/internal/global/middleware"
	authProvider "GAMERS-BE/internal/global/security/jwt"
	"GAMERS-BE/internal/oauth2"
	"GAMERS-BE/internal/user"
	"context"
	"log"
	"os"

	_ "GAMERS-BE/docs"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

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

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Fatal("Failed to close Redis client:", err)
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
	userDeps := user.ProvideUserDependencies(db, appRouter)
	oauth2Deps := oauth2.ProvideOAuth2Dependencies(db, appRouter)
	contestDeps := contest.ProvideContestDependencies(db, appRouter)

	setupRouter(appRouter, authDeps, userDeps, oauth2Deps, contestDeps)

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
) *router.Router {

	appRouter.Engine().Use(middleware.GlobalErrorHandler())

	appRouter.RegisterHealthCheck()
	appRouter.RegisterSwagger(ginSwagger.WrapHandler(swaggerFiles.Handler))

	authDeps.Controller.RegisterRoutes()
	userDeps.Controller.RegisterRoutes()
	oauth2Deps.Controller.RegisterRoutes()
	contestDeps.Controller.RegisterRoute()

	return appRouter
}

func initDatabase() *gorm.DB {
	dbConfig := database.NewConfigFromEnv()
	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		log.Println("🔄 Running database migrations...")
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("Failed to get SQL DB:", err)
		}

		migrationsPath := os.Getenv("MIGRATIONS_PATH")
		if migrationsPath == "" {
			migrationsPath = "./db/migrations"
		}

		if err := database.RunMigrations(sqlDB, migrationsPath); err != nil {
			log.Fatal("Failed to run migrations:", err)
		}
	}

	return db
}

func initRedis() *redis.Client {
	redisConfig := database.NewRedisConfigFromEnv()
	redisClient, err := database.InitRedis(redisConfig)

	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	return redisClient
}
