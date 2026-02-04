package oauth2

import (
	authAdapter "github.com/FOR-GAMERS/GAMERS-BE/internal/auth/infra/persistence/adapter"
	discordPort "github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	jwtApplication "github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/utils"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/infra/discord"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/infra/state"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/presentation"
	userAdapter "github.com/FOR-GAMERS/GAMERS-BE/internal/user/infra/persistence/adapter"
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
	cookieDomain := utils.GetEnv("COOKIE_DOMAIN", "")
	discordController := presentation.NewDiscordController(router, oauth2Service, webURL, cookieDomain)

	return &Dependencies{
		Controller:       discordController,
		OAuth2Repository: oauth2DatabaseAdapter,
	}
}
