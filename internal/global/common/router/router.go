package router

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine         *gin.Engine
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(authMiddleware *middleware.AuthMiddleware, webURL string) *Router {
	engine := gin.New()
	engine.SetTrustedProxies(nil)
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	allowedOrigins := []string{"http://localhost:3000", "http://localhost:5173"}
	if webURL != "" && webURL != "http://localhost:3000" {
		allowedOrigins = append(allowedOrigins, webURL)
	}

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	return &Router{
		engine:         engine,
		authMiddleware: authMiddleware,
	}
}

func (r *Router) PublicGroup(path string) *gin.RouterGroup {
	return r.engine.Group(path)
}

func (r *Router) ProtectedGroup(path string) *gin.RouterGroup {
	group := r.engine.Group(path)
	group.Use(r.authMiddleware.RequireAuth())
	return group
}

func (r *Router) AdminGroup(path string) *gin.RouterGroup {
	group := r.engine.Group(path)
	group.Use(r.authMiddleware.RequireAuth())
	group.Use(r.authMiddleware.RequireAdmin())
	return group
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) RegisterHealthCheck() {
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

func (r *Router) RegisterSwagger(handler gin.HandlerFunc) {
	r.engine.GET("/swagger/*any", handler)
}
