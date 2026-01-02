package presentation

import (
	"GAMERS-BE/internal/global/response"
	"GAMERS-BE/internal/oauth2/application"
	"GAMERS-BE/internal/oauth2/application/dto"

	"github.com/gin-gonic/gin"
)

type DiscordController struct {
	oauth2Service *application.OAuth2Service
}

func NewDiscordController(oauth2Service *application.OAuth2Service) *DiscordController {
	return &DiscordController{
		oauth2Service: oauth2Service,
	}
}

func (c *DiscordController) RegisterRoutes(router *gin.Engine) {
	oauth2Group := router.Group("/api/oauth2")
	{
		oauth2Group.GET("/discord/login", c.DiscordLogin)
		oauth2Group.GET("/discord/callback", c.DiscordCallback)
	}
}

// DiscordLogin godoc
// @Summary Discord OAuth2 Login
// @Description Redirect to Discord OAuth2 login page
// @Tags oauth2
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to Discord login page"
// @Router /api/oauth2/discord/login [get]
func (c *DiscordController) DiscordLogin(ctx *gin.Context) {
	loginURL := c.oauth2Service.GetDiscordLoginURL()

	// Redirect to Discord login page
	ctx.Redirect(302, loginURL)
}

// DiscordCallback godoc
// @Summary Discord OAuth2 Callback
// @Description Handle Discord OAuth2 callback
// @Tags oauth2
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string false "State"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/oauth2/discord/callback [get]
func (c *DiscordController) DiscordCallback(ctx *gin.Context) {
	var req dto.DiscordCallbackRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request parameters"))
		return
	}

	result, err := c.oauth2Service.HandleDiscordCallback(&req)
	if err != nil {
		response.JSON(ctx, response.InternalServerError(err.Error()))
		return
	}

	response.JSON(ctx, response.Success(result, "Discord login successful"))
}
