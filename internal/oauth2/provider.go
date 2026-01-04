package oauth2

import (
	"GAMERS-BE/internal/auth/infra/jwt"
	"GAMERS-BE/internal/oauth2/application"
	"GAMERS-BE/internal/oauth2/infra/discord"
	"GAMERS-BE/internal/oauth2/infra/persistence/adapter"
	"GAMERS-BE/internal/oauth2/infra/state"
	"GAMERS-BE/internal/oauth2/presentation"
	userAdapter "GAMERS-BE/internal/user/infra/persistence/adapter"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Dependencies struct {
	Controller *presentation.DiscordController
}

func ProvideOAuth2Dependencies(db *gorm.DB, router *gin.Engine) *Dependencies {
	ctx := context.Background()

	discordConfig := discord.NewConfigFromEnv()

	discordClient := discord.NewDiscordClient()

	oauth2UserAdapter := userAdapter.NewOAuth2UserAdapter(db.WithContext(ctx))
	oauth2DatabaseAdapter := adapter.NewOAuth2DatabaseAdapter(db.WithContext(ctx))

	jwtConfig := jwt.NewConfigFromEnv()
	tokenManager := jwt.NewTokenManager(jwtConfig)
	tokenProvider := jwt.NewTokenProvider(tokenManager)
	stateManager := state.NewStateManager(10 * time.Minute)

	oauth2Service := application.NewOAuth2Service(
		ctx,
		discordConfig,
		discordClient,
		stateManager,
		oauth2UserAdapter,
		oauth2DatabaseAdapter,
		tokenProvider,
	)

	discordController := presentation.NewDiscordController(oauth2Service)

	discordController.RegisterRoutes(router)

	return &Dependencies{
		Controller: discordController,
	}
}
