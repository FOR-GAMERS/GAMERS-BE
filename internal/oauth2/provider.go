package oauth2

import (
	"GAMERS-BE/internal/auth/infra/jwt"
	"GAMERS-BE/internal/oauth2/application"
	"GAMERS-BE/internal/oauth2/infra/discord"
	"GAMERS-BE/internal/oauth2/infra/persistence/adapter"
	"GAMERS-BE/internal/oauth2/presentation"
	userCommand "GAMERS-BE/internal/user/infra/persistence/command"
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OAuth2Dependencies struct {
	Controller *presentation.DiscordController
}

func NewOAuth2Provider(db *gorm.DB, router *gin.Engine) *OAuth2Dependencies {
	ctx := context.Background()

	discordConfig := discord.NewConfigFromEnv()

	discordClient := discord.NewDiscordClient()

	oauth2DatabaseAdapter := adapter.NewOAuth2DatabaseAdapter(db)

	userCommandAdapter := userCommand.NewMySQLUserRepository(db)

	jwtConfig := jwt.NewConfigFromEnv()
	tokenManager := jwt.NewTokenManager(jwtConfig)
	tokenProvider := jwt.NewTokenProvider(tokenManager)

	oauth2Service := application.NewOAuth2Service(
		ctx,
		discordConfig,
		discordClient,
		oauth2DatabaseAdapter,
		userCommandAdapter,
		tokenProvider,
	)

	discordController := presentation.NewDiscordController(oauth2Service)

	discordController.RegisterRoutes(router)

	return &OAuth2Dependencies{
		Controller: discordController,
	}
}
