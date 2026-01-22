package port

import (
	"GAMERS-BE/internal/discord/application/dto"
)

// DiscordBotPort defines the interface for Discord Bot API operations
type DiscordBotPort interface {
	// GetBotGuilds returns all guilds the bot is a member of
	GetBotGuilds() ([]dto.DiscordGuild, error)

	// GetGuildChannels returns all channels in a guild
	GetGuildChannels(guildID string) ([]dto.DiscordChannel, error)

	// GetGuildTextChannels returns only text channels in a guild
	GetGuildTextChannels(guildID string) ([]dto.DiscordChannel, error)

	// IsBotInGuild checks if the bot is in a specific guild
	IsBotInGuild(guildID string) (bool, error)

	// ValidateGuildChannel validates if a channel exists in a guild and is a text channel
	ValidateGuildChannel(guildID, channelID string) (bool, error)

	// IsUserInGuild checks if a user (by Discord ID) is in a specific guild
	// This uses the bot's permissions to check guild membership
	IsUserInGuild(guildID, discordUserID string) (bool, error)

	// GetGuildMember returns information about a guild member
	GetGuildMember(guildID, discordUserID string) (*dto.DiscordGuildMember, error)
}

// DiscordUserPort defines the interface for Discord User API operations (using user's OAuth token)
type DiscordUserPort interface {
	// GetUserGuilds returns all guilds the user is a member of
	GetUserGuilds(accessToken string) ([]dto.DiscordGuild, error)

	// IsUserInGuild checks if the user is in a specific guild
	IsUserInGuild(accessToken, guildID string) (bool, error)

	// HasManageGuildPermission checks if the user has MANAGE_GUILD permission
	HasManageGuildPermission(accessToken, guildID string) (bool, error)
}

// DiscordTokenPort defines the interface for Discord token storage operations (Redis)
type DiscordTokenPort interface {
	// SaveToken saves a Discord OAuth2 token to Redis
	SaveToken(token *dto.DiscordToken) error

	// GetToken retrieves a Discord OAuth2 token from Redis by user ID
	GetToken(userID int64) (*dto.DiscordToken, error)

	// DeleteToken removes a Discord OAuth2 token from Redis
	DeleteToken(userID int64) error

	// ExistsToken checks if a token exists for the given user ID
	ExistsToken(userID int64) (bool, error)
}
