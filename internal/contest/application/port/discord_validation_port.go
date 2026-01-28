package port

// DiscordChannel represents a Discord channel for contest service
type DiscordChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	GuildID  string `json:"guild_id"`
	Position int    `json:"position"`
}

// DiscordGuild represents a Discord guild for contest service
type DiscordGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// DiscordValidationPort defines the interface for Discord validation in contest service
type DiscordValidationPort interface {
	// ValidateGuildForContest validates that the bot and user are in the guild,
	// and that the channel exists and is a text channel
	ValidateGuildForContest(guildID, channelID, userDiscordID string) error

	// ValidateBotInGuild checks if the bot is in the specified guild
	ValidateBotInGuild(guildID string) error

	// ValidateUserInGuild checks if the user is in the specified guild
	ValidateUserInGuild(guildID, userDiscordID string) error

	// GetGuildTextChannels returns all text channels in a guild
	GetGuildTextChannels(guildID string) ([]DiscordChannel, error)

	// GetBotGuilds returns all guilds the bot is a member of
	GetBotGuilds() ([]DiscordGuild, error)
}
