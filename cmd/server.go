package main

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/infra/persistence"
	"GAMERS-BE/internal/user/presentation"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	userRepository := persistence.NewInMemoryUserRepository()
	passwordHasher := password.NewBcryptPasswordHasher()
	userService := application.NewUserService(userRepository, passwordHasher)
	userController := presentation.NewUserController(userService)

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	userController.RegisterRoutes(router)

	log.Println("===========================================")
	log.Println("🎮 GAMERS API Server Starting")
	log.Println("===========================================")
	log.Println("Server:          http://localhost:8080")
	log.Println("Health Check:    http://localhost:8080/health")
	log.Println("===========================================")

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
