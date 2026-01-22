package presentation

import (
	"GAMERS-BE/internal/auth/middleware"
	"GAMERS-BE/internal/discord/application"
	"GAMERS-BE/internal/discord/application/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Swagger type aliases for documentation
var (
	_ dto.DiscordGuild
	_ dto.DiscordChannel
	_ response.Response
)

// DiscordController handles Discord-related HTTP requests
type DiscordController struct {
	validationService *application.DiscordValidationService
	helper            *handler.ControllerHelper
}

// NewDiscordController creates a new Discord controller
func NewDiscordController(
	r *router.Router,
	validationService *application.DiscordValidationService,
	helper *handler.ControllerHelper,
) *DiscordController {
	controller := &DiscordController{
		validationService: validationService,
		helper:            helper,
	}

	controller.RegisterRoutes(r)
	return controller
}

// RegisterRoutes registers the Discord routes
func (c *DiscordController) RegisterRoutes(r *router.Router) {
	discordGroup := r.ProtectedGroup("/api/discord")
	{
		discordGroup.GET("/guilds", c.GetAvailableGuilds)
		discordGroup.GET("/guilds/:guildId/channels", c.GetAvailableGuildTextChannels)
	}
}

// GetAvailableGuilds godoc
// @Summary Get available guilds for contest creation
// @Description Returns all guilds where both the GAMERS bot and the authenticated user are members
// @Tags Discord
// @Accept json
// @Produce json
// @Success 200 {array} dto.DiscordGuild
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /discord/guilds [get]
func (c *DiscordController) GetAvailableGuilds(ctx *gin.Context) {
	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	guilds, err := c.validationService.GetAvailableGuilds(userID)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"guilds": guilds,
	})
}

// GetAvailableGuildTextChannels godoc
// @Summary Get guild's text channels
// @Description Returns all text channels in a guild where both the bot and the authenticated user are members
// @Tags Discord
// @Accept json
// @Produce json
// @Param guildId path string true "Discord Guild ID"
// @Success 200 {array} dto.DiscordChannel
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /discord/guilds/{guildId}/channels [get]
func (c *DiscordController) GetAvailableGuildTextChannels(ctx *gin.Context) {
	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	guildID := ctx.Param("guildId")
	if guildID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "guild_id is required",
		})
		return
	}

	channels, err := c.validationService.GetAvailableGuildTextChannels(guildID, userID)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"channels": channels,
	})
}
