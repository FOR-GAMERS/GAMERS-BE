package infra

import (
	"GAMERS-BE/internal/discord/application/port"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DiscordAPIBaseURL = "https://discord.com/api/v10"
)

// DiscordBotClient implements DiscordBotPort using Discord Bot token
type DiscordBotClient struct {
	httpClient *http.Client
	botToken   string
}

// NewDiscordBotClient creates a new Discord Bot client
func NewDiscordBotClient() *DiscordBotClient {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	return &DiscordBotClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		botToken: botToken,
	}
}

// NewDiscordBotClientWithToken creates a new Discord Bot client with a specific token
func NewDiscordBotClientWithToken(botToken string) *DiscordBotClient {
	return &DiscordBotClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		botToken: botToken,
	}
}

// doRequest performs an HTTP request with Bot authentication
func (c *DiscordBotClient) doRequest(method, endpoint string) ([]byte, error) {
	req, err := http.NewRequest(method, DiscordAPIBaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+c.botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetBotGuilds returns all guilds the bot is a member of
func (c *DiscordBotClient) GetBotGuilds() ([]port.DiscordGuild, error) {
	body, err := c.doRequest("GET", "/users/@me/guilds")
	if err != nil {
		return nil, err
	}

	var guilds []port.DiscordGuild
	if err := json.Unmarshal(body, &guilds); err != nil {
		return nil, err
	}

	return guilds, nil
}

// GetGuildChannels returns all channels in a guild
func (c *DiscordBotClient) GetGuildChannels(guildID string) ([]port.DiscordChannel, error) {
	body, err := c.doRequest("GET", "/guilds/"+guildID+"/channels")
	if err != nil {
		return nil, err
	}

	var channels []port.DiscordChannel
	if err := json.Unmarshal(body, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

// GetGuildTextChannels returns only text channels in a guild
func (c *DiscordBotClient) GetGuildTextChannels(guildID string) ([]port.DiscordChannel, error) {
	channels, err := c.GetGuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	textChannels := make([]port.DiscordChannel, 0)
	for _, ch := range channels {
		if ch.Type == port.ChannelTypeGuildText {
			textChannels = append(textChannels, ch)
		}
	}

	return textChannels, nil
}

// IsBotInGuild checks if the bot is in a specific guild
func (c *DiscordBotClient) IsBotInGuild(guildID string) (bool, error) {
	guilds, err := c.GetBotGuilds()
	if err != nil {
		return false, err
	}

	for _, guild := range guilds {
		if guild.ID == guildID {
			return true, nil
		}
	}

	return false, nil
}

// ValidateGuildChannel validates if a channel exists in a guild and is a text channel
func (c *DiscordBotClient) ValidateGuildChannel(guildID, channelID string) (bool, error) {
	channels, err := c.GetGuildChannels(guildID)
	if err != nil {
		return false, err
	}

	for _, ch := range channels {
		if ch.ID == channelID && ch.Type == port.ChannelTypeGuildText {
			return true, nil
		}
	}

	return false, nil
}

// guildMemberResponse represents the Discord API response for a guild member
type guildMemberResponse struct {
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	Nick     string   `json:"nick"`
	Roles    []string `json:"roles"`
	JoinedAt string   `json:"joined_at"`
}

// GetGuildMember returns information about a guild member
func (c *DiscordBotClient) GetGuildMember(guildID, discordUserID string) (*port.DiscordGuildMember, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/guilds/%s/members/%s", guildID, discordUserID))
	if err != nil {
		return nil, err
	}

	var member guildMemberResponse
	if err := json.Unmarshal(body, &member); err != nil {
		return nil, err
	}

	return &port.DiscordGuildMember{
		UserID:   member.User.ID,
		Nick:     member.Nick,
		Roles:    member.Roles,
		JoinedAt: member.JoinedAt,
	}, nil
}

// IsUserInGuild checks if a user (by Discord ID) is in a specific guild
func (c *DiscordBotClient) IsUserInGuild(guildID, discordUserID string) (bool, error) {
	_, err := c.GetGuildMember(guildID, discordUserID)
	if err != nil {
		// If we get a 404 or 10007 (Unknown Member), the user is not in the guild
		errStr := err.Error()
		if strings.Contains(errStr, "404") || strings.Contains(errStr, "10007") || strings.Contains(errStr, "Unknown Member") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// DiscordUserClient implements DiscordUserPort using user's OAuth access token
type DiscordUserClient struct {
	httpClient *http.Client
}

// NewDiscordUserClient creates a new Discord User client
func NewDiscordUserClient() *DiscordUserClient {
	return &DiscordUserClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with Bearer token authentication
func (c *DiscordUserClient) doRequest(method, endpoint, accessToken string) ([]byte, error) {
	req, err := http.NewRequest(method, DiscordAPIBaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetUserGuilds returns all guilds the user is a member of
func (c *DiscordUserClient) GetUserGuilds(accessToken string) ([]port.DiscordGuild, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	body, err := c.doRequest("GET", "/users/@me/guilds", accessToken)
	if err != nil {
		return nil, err
	}

	var guilds []port.DiscordGuild
	if err := json.Unmarshal(body, &guilds); err != nil {
		return nil, err
	}

	return guilds, nil
}

// IsUserInGuild checks if the user is in a specific guild
func (c *DiscordUserClient) IsUserInGuild(accessToken, guildID string) (bool, error) {
	guilds, err := c.GetUserGuilds(accessToken)
	if err != nil {
		return false, err
	}

	for _, guild := range guilds {
		if guild.ID == guildID {
			return true, nil
		}
	}

	return false, nil
}

// HasManageGuildPermission checks if the user has MANAGE_GUILD permission
// MANAGE_GUILD permission bit is 0x00000020 (32)
func (c *DiscordUserClient) HasManageGuildPermission(accessToken, guildID string) (bool, error) {
	guilds, err := c.GetUserGuilds(accessToken)
	if err != nil {
		return false, err
	}

	for _, guild := range guilds {
		if guild.ID == guildID {
			// Check if user is owner or has MANAGE_GUILD permission
			if guild.Owner {
				return true, nil
			}
			// Permission check would require parsing the permissions string
			// For now, we consider owner as having full permissions
			return guild.Owner, nil
		}
	}

	return false, nil
}
