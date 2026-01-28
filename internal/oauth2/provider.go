package oauth2

import (
	authAdapter "GAMERS-BE/internal/auth/infra/persistence/adapter"
	discordPort "GAMERS-BE/internal/discord/application/port"
	"GAMERS-BE/internal/global/common/router"
	jwtApplication "GAMERS-BE/internal/global/security/jwt"
	"GAMERS-BE/internal/global/utils"
	"GAMERS-BE/internal/oauth2/application"
	"GAMERS-BE/internal/oauth2/application/port"
	"GAMERS-BE/internal/oauth2/infra/discord"
	"GAMERS-BE/internal/oauth2/infra/persistence/adapter"
	"GAMERS-BE/internal/oauth2/infra/state"
	"GAMERS-BE/internal/oauth2/presentation"
	userAdapter "GAMERS-BE/internal/user/infra/persistence/adapter"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Controller       *presentation.DiscordController
	OAuth2Repository port.OAuth2DatabasePort
}

func ProvideOAuth2Dependencies(db *gorm.DB, redisClient *redis.Client, ctx *context.Context, router *router.Router, discordTokenPort discordPort.DiscordTokenPort) *Dependencies {
	discordConfig := discord.NewConfigFromEnv()

	discordClient := discord.NewDiscordClient()

	oauth2UserAdapter := userAdapter.NewOAuth2UserAdapter(db.WithContext(*ctx))
	oauth2DatabaseAdapter := adapter.NewOAuth2DatabaseAdapter(db.WithContext(*ctx))

	stateManager := state.NewStateManager(10 * time.Minute)
	tokenService := jwtApplication.ProvideJwtService()

	refreshTokenCacheAdapter := authAdapter.NewRefreshTokenCacheAdapter(ctx, redisClient)

	oauth2Service := application.NewOAuth2Service(
		*ctx,
		discordConfig,
		discordClient,
		stateManager,
		oauth2UserAdapter,
		oauth2DatabaseAdapter,
		discordTokenPort,
		refreshTokenCacheAdapter,
		*tokenService,
	)

	webURL := utils.GetEnv("WEB_URL", "http://localhost:3000")
	discordController := presentation.NewDiscordController(router, oauth2Service, webURL)

	return &Dependencies{
		Controller:       discordController,
		OAuth2Repository: oauth2DatabaseAdapter,
	}
}
