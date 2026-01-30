package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application/port"
	discordApplication "github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application"
	discordDto "github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/dto"
)

// DiscordValidationAdapter adapts the Discord validation service to the contest port
type DiscordValidationAdapter struct {
	validationService *discordApplication.DiscordValidationService
}

// NewDiscordValidationAdapter creates a new Discord validation adapter
func NewDiscordValidationAdapter(validationService *discordApplication.DiscordValidationService) *DiscordValidationAdapter {
	return &DiscordValidationAdapter{
		validationService: validationService,
	}
}

// ValidateGuildForContest validates that the bot and user are in the guild,
// and that the channel exists and is a text channel
func (a *DiscordValidationAdapter) ValidateGuildForContest(guildID, channelID, userDiscordID string) error {
	return a.validationService.ValidateGuildForContest(guildID, channelID, userDiscordID)
}

// ValidateBotInGuild checks if the bot is in the specified guild
func (a *DiscordValidationAdapter) ValidateBotInGuild(guildID string) error {
	return a.validationService.ValidateBotInGuild(guildID)
}

// ValidateUserInGuild checks if the user is in the specified guild
func (a *DiscordValidationAdapter) ValidateUserInGuild(guildID, userDiscordID string) error {
	return a.validationService.ValidateUserInGuild(guildID, userDiscordID)
}

// GetGuildTextChannels returns all text channels in a guild
func (a *DiscordValidationAdapter) GetGuildTextChannels(guildID string) ([]port.DiscordChannel, error) {
	channels, err := a.validationService.GetGuildTextChannels(guildID)
	if err != nil {
		return nil, err
	}

	// Convert discord dto channels to contest port channels
	result := make([]port.DiscordChannel, len(channels))
	for i, ch := range channels {
		result[i] = convertChannel(ch)
	}
	return result, nil
}

// GetBotGuilds returns all guilds the bot is a member of
func (a *DiscordValidationAdapter) GetBotGuilds() ([]port.DiscordGuild, error) {
	guilds, err := a.validationService.GetBotGuilds()
	if err != nil {
		return nil, err
	}

	// Convert discord dto guilds to contest port guilds
	result := make([]port.DiscordGuild, len(guilds))
	for i, g := range guilds {
		result[i] = convertGuild(g)
	}
	return result, nil
}

func convertChannel(ch discordDto.DiscordChannel) port.DiscordChannel {
	return port.DiscordChannel{
		ID:       ch.ID,
		Name:     ch.Name,
		Type:     ch.Type,
		GuildID:  ch.GuildID,
		Position: ch.Position,
	}
}

func convertGuild(g discordDto.DiscordGuild) port.DiscordGuild {
	return port.DiscordGuild{
		ID:   g.ID,
		Name: g.Name,
		Icon: g.Icon,
	}
}
