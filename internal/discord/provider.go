package discord

import (
	"GAMERS-BE/internal/discord/application"
	"GAMERS-BE/internal/discord/application/port"
	"GAMERS-BE/internal/discord/infra"
	"GAMERS-BE/internal/discord/presentation"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
)

type Dependencies struct {
	BotClient         port.DiscordBotPort
	UserClient        port.DiscordUserPort
	ValidationService *application.DiscordValidationService
	Controller        *presentation.DiscordController
}

func ProvideDiscordDependencies(r *router.Router) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	botClient := infra.NewDiscordBotClient()
	userClient := infra.NewDiscordUserClient()
	validationService := application.NewDiscordValidationService(botClient)
	controller := presentation.NewDiscordController(r, validationService, controllerHelper)

	return &Dependencies{
		BotClient:         botClient,
		UserClient:        userClient,
		ValidationService: validationService,
		Controller:        controller,
	}
}
