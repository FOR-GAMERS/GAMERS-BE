package dto

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

// DiscordToken represents a Discord OAuth2 token stored in Redis
type DiscordToken struct {
	UserID       int64  `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
}
