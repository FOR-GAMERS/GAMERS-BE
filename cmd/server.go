package main

import (
	"GAMERS-BE/internal/auth"
	"GAMERS-BE/internal/global/database"
	"GAMERS-BE/internal/global/middleware"
	"GAMERS-BE/internal/oauth2"
	"GAMERS-BE/internal/user"
	"context"
	"log"
	"os"

	_ "GAMERS-BE/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	} else {
		log.Println("⏭️  Skipping migrations (RUN_MIGRATIONS not set)")
	}

	redisConfig := database.NewRedisConfigFromEnv()
	redisClient, err := database.InitRedis(redisConfig)
	if err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}

	ctx := context.Background()

	router := gin.Default()
	router.Use(middleware.GlobalErrorHandler())

	authDeps := auth.ProvideAuthDependencies(db, redisClient, ctx)
	userDeps := user.ProvideUserDependencies(db)
	_ = oauth2.ProvideOAuth2Dependencies(db, router)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authDeps.Controller.RegisterRoutes(router, authDeps.AuthMiddleware)
	userDeps.Controller.RegisterRoutes(router, authDeps.AuthMiddleware)

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

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
