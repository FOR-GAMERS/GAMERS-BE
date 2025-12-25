package main

import (
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/infra/persistence"
	"GAMERS-BE/internal/user/presentation"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	userRepository := persistence.NewInMemoryUserRepository()
	userService := application.NewUserService(userRepository)
	userController := presentation.NewUserController(userService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	userController.RegisterRoutes(router)

	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
