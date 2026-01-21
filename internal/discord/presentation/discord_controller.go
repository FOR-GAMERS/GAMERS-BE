package presentation

import (
	"GAMERS-BE/internal/discord/application"
	"GAMERS-BE/internal/discord/application/port"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Swagger type aliases for documentation
var (
	_ port.DiscordGuild
	_ port.DiscordChannel
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
		discordGroup.GET("/guilds", c.GetBotGuilds)
		discordGroup.GET("/guilds/:guildId/channels", c.GetGuildTextChannels)
	}
}

// GetBotGuilds godoc
// @Summary Get bot's guilds
// @Description Returns all guilds the GAMERS bot is a member of
// @Tags Discord
// @Accept json
// @Produce json
// @Success 200 {array} port.DiscordGuild
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /discord/guilds [get]
func (c *DiscordController) GetBotGuilds(ctx *gin.Context) {
	guilds, err := c.validationService.GetBotGuilds()
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"guilds": guilds,
	})
}

// GetGuildTextChannels godoc
// @Summary Get guild's text channels
// @Description Returns all text channels in a guild that the bot has access to
// @Tags Discord
// @Accept json
// @Produce json
// @Param guildId path string true "Discord Guild ID"
// @Success 200 {array} port.DiscordChannel
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /discord/guilds/{guildId}/channels [get]
func (c *DiscordController) GetGuildTextChannels(ctx *gin.Context) {
	guildID := ctx.Param("guildId")
	if guildID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "guild_id is required",
		})
		return
	}

	channels, err := c.validationService.GetGuildTextChannels(guildID)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"channels": channels,
	})
}
