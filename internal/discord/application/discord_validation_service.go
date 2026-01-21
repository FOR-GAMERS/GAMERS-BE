package application

import (
	"GAMERS-BE/internal/discord/application/port"
	"GAMERS-BE/internal/global/exception"
)

// DiscordValidationService handles Discord validation logic for contests
type DiscordValidationService struct {
	botClient port.DiscordBotPort
}

// NewDiscordValidationService creates a new Discord validation service
func NewDiscordValidationService(botClient port.DiscordBotPort) *DiscordValidationService {
	return &DiscordValidationService{
		botClient: botClient,
	}
}

// ValidateGuildForContest validates that:
// 1. The bot is in the specified guild
// 2. The user (by Discord ID) is in the specified guild
// 3. The channel exists in the guild and is a text channel
func (s *DiscordValidationService) ValidateGuildForContest(guildID, channelID, userDiscordID string) error {
	// Check if bot is in the guild
	botInGuild, err := s.botClient.IsBotInGuild(guildID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !botInGuild {
		return exception.ErrBotNotInGuild
	}

	// Check if user is in the guild
	userInGuild, err := s.botClient.IsUserInGuild(guildID, userDiscordID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !userInGuild {
		return exception.ErrUserNotInGuild
	}

	// Check if channel exists and is a text channel
	validChannel, err := s.botClient.ValidateGuildChannel(guildID, channelID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !validChannel {
		return exception.ErrChannelNotInGuild
	}

	return nil
}

// ValidateBotInGuild checks if the bot is in the specified guild
func (s *DiscordValidationService) ValidateBotInGuild(guildID string) error {
	botInGuild, err := s.botClient.IsBotInGuild(guildID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !botInGuild {
		return exception.ErrBotNotInGuild
	}
	return nil
}

// ValidateUserInGuild checks if the user is in the specified guild
func (s *DiscordValidationService) ValidateUserInGuild(guildID, userDiscordID string) error {
	userInGuild, err := s.botClient.IsUserInGuild(guildID, userDiscordID)
	if err != nil {
		return exception.ErrDiscordAPIError
	}
	if !userInGuild {
		return exception.ErrUserNotInGuild
	}
	return nil
}

// GetGuildTextChannels returns all text channels in a guild
func (s *DiscordValidationService) GetGuildTextChannels(guildID string) ([]port.DiscordChannel, error) {
	// First verify bot is in the guild
	if err := s.ValidateBotInGuild(guildID); err != nil {
		return nil, err
	}

	channels, err := s.botClient.GetGuildTextChannels(guildID)
	if err != nil {
		return nil, exception.ErrDiscordAPIError
	}

	return channels, nil
}

// GetBotGuilds returns all guilds the bot is a member of
func (s *DiscordValidationService) GetBotGuilds() ([]port.DiscordGuild, error) {
	guilds, err := s.botClient.GetBotGuilds()
	if err != nil {
		return nil, exception.ErrDiscordAPIError
	}
	return guilds, nil
}
