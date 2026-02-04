package infra

import (
	"context"
	"encoding/json"
	"time"

	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/utils"
	"github.com/redis/go-redis/v9"
)

const discordCacheTTL = 5 * time.Minute

// CachedDiscordBotClient is a decorator around DiscordBotPort that adds Redis caching
// for guild and channel data. User-specific lookups are delegated directly to the inner client.
type CachedDiscordBotClient struct {
	inner       port.DiscordBotPort
	redisClient *redis.Client
	ctx         context.Context
}

// NewCachedDiscordBotClient wraps an existing DiscordBotPort with Redis caching.
func NewCachedDiscordBotClient(inner port.DiscordBotPort, redisClient *redis.Client, ctx context.Context) *CachedDiscordBotClient {
	return &CachedDiscordBotClient{
		inner:       inner,
		redisClient: redisClient,
		ctx:         ctx,
	}
}

// GetBotGuilds returns all guilds the bot is a member of, with Redis caching.
func (c *CachedDiscordBotClient) GetBotGuilds() ([]dto.DiscordGuild, error) {
	key := utils.GetDiscordBotGuildsKey()

	cached, err := c.redisClient.Get(c.ctx, key).Result()
	if err == nil {
		var guilds []dto.DiscordGuild
		if jsonErr := json.Unmarshal([]byte(cached), &guilds); jsonErr == nil {
			return guilds, nil
		}
	}

	guilds, err := c.inner.GetBotGuilds()
	if err != nil {
		return nil, err
	}

	if data, jsonErr := json.Marshal(guilds); jsonErr == nil {
		c.redisClient.Set(c.ctx, key, data, discordCacheTTL)
	}

	return guilds, nil
}

// GetGuildChannels returns all channels in a guild, with Redis caching.
func (c *CachedDiscordBotClient) GetGuildChannels(guildID string) ([]dto.DiscordChannel, error) {
	key := utils.GetDiscordGuildChannelsKey(guildID)

	cached, err := c.redisClient.Get(c.ctx, key).Result()
	if err == nil {
		var channels []dto.DiscordChannel
		if jsonErr := json.Unmarshal([]byte(cached), &channels); jsonErr == nil {
			return channels, nil
		}
	}

	channels, err := c.inner.GetGuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	if data, jsonErr := json.Marshal(channels); jsonErr == nil {
		c.redisClient.Set(c.ctx, key, data, discordCacheTTL)
	}

	return channels, nil
}

// GetGuildTextChannels returns only text channels in a guild.
// Uses the cached GetGuildChannels internally.
func (c *CachedDiscordBotClient) GetGuildTextChannels(guildID string) ([]dto.DiscordChannel, error) {
	channels, err := c.GetGuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	textChannels := make([]dto.DiscordChannel, 0)
	for _, ch := range channels {
		if ch.Type == dto.ChannelTypeGuildText {
			textChannels = append(textChannels, ch)
		}
	}

	return textChannels, nil
}

// IsBotInGuild checks if the bot is in a specific guild.
// Uses the cached GetBotGuilds internally.
func (c *CachedDiscordBotClient) IsBotInGuild(guildID string) (bool, error) {
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

// ValidateGuildChannel validates if a channel exists in a guild and is a text channel.
// Uses the cached GetGuildChannels internally.
func (c *CachedDiscordBotClient) ValidateGuildChannel(guildID, channelID string) (bool, error) {
	channels, err := c.GetGuildChannels(guildID)
	if err != nil {
		return false, err
	}

	for _, ch := range channels {
		if ch.ID == channelID && ch.Type == dto.ChannelTypeGuildText {
			return true, nil
		}
	}

	return false, nil
}

// IsUserInGuild delegates directly to the inner client (user-specific, not cached).
func (c *CachedDiscordBotClient) IsUserInGuild(guildID, discordUserID string) (bool, error) {
	return c.inner.IsUserInGuild(guildID, discordUserID)
}

// GetGuildMember delegates directly to the inner client (user-specific, not cached).
func (c *CachedDiscordBotClient) GetGuildMember(guildID, discordUserID string) (*dto.DiscordGuildMember, error) {
	return c.inner.GetGuildMember(guildID, discordUserID)
}
