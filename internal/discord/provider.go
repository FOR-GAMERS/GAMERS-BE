package discord

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/infra"
	discordAdapter "github.com/FOR-GAMERS/GAMERS-BE/internal/discord/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/presentation"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	oauth2Port "github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application/port"
	oauth2Adapter "github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/infra/persistence/adapter"
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	BotClient         port.DiscordBotPort
	UserClient        port.DiscordUserPort
	DiscordTokenPort  port.DiscordTokenPort
	ValidationService *application.DiscordValidationService
	Controller        *presentation.DiscordController
}

func ProvideDiscordDependencies(r *router.Router, db *gorm.DB, redisClient *redis.Client, ctx *context.Context) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	botClient := infra.NewDiscordBotClient()
	userClient := infra.NewDiscordUserClient()

	// Create Discord token Redis adapter for storing Discord OAuth2 tokens
	discordTokenAdapter := discordAdapter.NewDiscordTokenRedisAdapter(ctx, redisClient)

	// Create OAuth2 config adapter to look up user's Discord account
	var oauth2DBPort oauth2Port.OAuth2DatabasePort = oauth2Adapter.NewOAuth2DatabaseAdapter(db)

	validationService := application.NewDiscordValidationService(botClient, userClient, discordTokenAdapter, oauth2DBPort)
	controller := presentation.NewDiscordController(r, validationService, controllerHelper)

	return &Dependencies{
		BotClient:         botClient,
		UserClient:        userClient,
		DiscordTokenPort:  discordTokenAdapter,
		ValidationService: validationService,
		Controller:        controller,
	}
}
