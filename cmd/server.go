package main

import (
	"GAMERS-BE/internal/common/database"
	"GAMERS-BE/internal/user/domain"
	"log"
	"os"

	_ "GAMERS-BE/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

func main() {
	if err := godotenv.Load("env/.env"); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	dbConfig := database.NewConfigFromEnv()
	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	if err := db.AutoMigrate(&domain.User{}, &domain.Profile{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	userController := InitializeUserController(db)
	profileController := InitializeProfileController(db)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	userController.RegisterRoutes(router)
	profileController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
