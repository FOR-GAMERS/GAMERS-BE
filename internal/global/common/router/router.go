package router

import (
	"GAMERS-BE/internal/auth/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine         *gin.Engine
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(authMiddleware *middleware.AuthMiddleware) *Router {
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

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
