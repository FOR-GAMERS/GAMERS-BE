package port

// DiscordGuild represents a Discord guild (server)
type DiscordGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions,omitempty"`
}

// DiscordChannel represents a Discord channel
type DiscordChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	GuildID  string `json:"guild_id"`
	Position int    `json:"position"`
	ParentID string `json:"parent_id,omitempty"`
}

// Discord channel types
const (
	ChannelTypeGuildText          = 0
	ChannelTypeDM                 = 1
	ChannelTypeGuildVoice         = 2
	ChannelTypeGroupDM            = 3
	ChannelTypeGuildCategory      = 4
	ChannelTypeGuildAnnouncement  = 5
	ChannelTypeAnnouncementThread = 10
	ChannelTypePublicThread       = 11
	ChannelTypePrivateThread      = 12
	ChannelTypeGuildStageVoice    = 13
	ChannelTypeGuildDirectory     = 14
	ChannelTypeGuildForum         = 15
)

// DiscordGuildMember represents a Discord guild member
type DiscordGuildMember struct {
	UserID   string   `json:"user_id"`
	Nick     string   `json:"nick,omitempty"`
	Roles    []string `json:"roles"`
	JoinedAt string   `json:"joined_at"`
}

// DiscordBotPort defines the interface for Discord Bot API operations
type DiscordBotPort interface {
	// GetBotGuilds returns all guilds the bot is a member of
	GetBotGuilds() ([]DiscordGuild, error)

	// GetGuildChannels returns all channels in a guild
	GetGuildChannels(guildID string) ([]DiscordChannel, error)

	// GetGuildTextChannels returns only text channels in a guild
	GetGuildTextChannels(guildID string) ([]DiscordChannel, error)

	// IsBotInGuild checks if the bot is in a specific guild
	IsBotInGuild(guildID string) (bool, error)

	// ValidateGuildChannel validates if a channel exists in a guild and is a text channel
	ValidateGuildChannel(guildID, channelID string) (bool, error)

	// IsUserInGuild checks if a user (by Discord ID) is in a specific guild
	// This uses the bot's permissions to check guild membership
	IsUserInGuild(guildID, discordUserID string) (bool, error)

	// GetGuildMember returns information about a guild member
	GetGuildMember(guildID, discordUserID string) (*DiscordGuildMember, error)
}

// DiscordUserPort defines the interface for Discord User API operations (using user's OAuth token)
type DiscordUserPort interface {
	// GetUserGuilds returns all guilds the user is a member of
	GetUserGuilds(accessToken string) ([]DiscordGuild, error)

	// IsUserInGuild checks if the user is in a specific guild
	IsUserInGuild(accessToken, guildID string) (bool, error)

	// HasManageGuildPermission checks if the user has MANAGE_GUILD permission
	HasManageGuildPermission(accessToken, guildID string) (bool, error)
}
