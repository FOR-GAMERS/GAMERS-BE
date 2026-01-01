package presentation

import (
	"GAMERS-BE/internal/auth/application"
	"GAMERS-BE/internal/auth/application/dto"
	"GAMERS-BE/internal/auth/presentation/middleware"
	"GAMERS-BE/internal/global/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *application.AuthService
}

func NewAuthController(authService *application.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// RegisterRoutes registers all authentication routes
func (c *AuthController) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/login", c.Login)
		authGroup.POST("/logout", c.Logout)
		authGroup.POST("/refresh", c.Refresh)
	}
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	result, err := c.authService.Login(&req)
	if err != nil {
		response.JSON(ctx, response.Error(401, err.Error()))
		return
	}

	response.JSON(ctx, response.Success(result, "Login successful"))
}

// Logout godoc
// @Summary Logout
// @Description Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param logout body dto.LogoutRequest true "Logout request"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Router /api/auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	var req dto.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	if err := c.authService.Logout(req); err != nil {
		response.JSON(ctx, response.BadRequest(err.Error()))
		return
	}

	response.SendNoContent(ctx)
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	var req dto.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	token, err := c.authService.Refresh(req)
	if err != nil {
		response.JSON(ctx, response.Error(401, err.Error()))
		return
	}

	response.JSON(ctx, response.Success(token, "Refresh successful"))
}
