package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/discord/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/utils"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// DiscordTokenRedisAdapter implements DiscordTokenPort using Redis
type DiscordTokenRedisAdapter struct {
	ctx    *context.Context
	client *redis.Client
}

// NewDiscordTokenRedisAdapter creates a new Discord token Redis adapter
func NewDiscordTokenRedisAdapter(ctx *context.Context, client *redis.Client) *DiscordTokenRedisAdapter {
	return &DiscordTokenRedisAdapter{
		ctx:    ctx,
		client: client,
	}
}

// SaveToken saves a Discord OAuth2 token to Redis with TTL based on expiration
func (a *DiscordTokenRedisAdapter) SaveToken(token *dto.DiscordToken) error {
	key := utils.GetDiscordTokenKey(token.UserID)

	data, err := json.Marshal(token)
	if err != nil {
		return exception.ErrRedisCannotSave
	}

	// Calculate TTL based on token expiration
	// Add some buffer time for refresh token usage
	ttl := time.Until(time.Unix(token.ExpiresAt, 0))
	if ttl <= 0 {
		// If token is already expired, use a default TTL for refresh token scenario
		ttl = 7 * 24 * time.Hour // 7 days for refresh token
	}

	if err := a.client.Set(*a.ctx, key, data, ttl).Err(); err != nil {
		return exception.ErrRedisCannotSave
	}

	return nil
}

// GetToken retrieves a Discord OAuth2 token from Redis by user ID
func (a *DiscordTokenRedisAdapter) GetToken(userID int64) (*dto.DiscordToken, error) {
	key := utils.GetDiscordTokenKey(userID)

	data, err := a.client.Get(*a.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, exception.ErrDiscordTokenNotFound
	}
	if err != nil {
		return nil, exception.ErrRedisGetError
	}

	var token dto.DiscordToken
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, exception.ErrRedisGetError
	}

	return &token, nil
}

// DeleteToken removes a Discord OAuth2 token from Redis
func (a *DiscordTokenRedisAdapter) DeleteToken(userID int64) error {
	key := utils.GetDiscordTokenKey(userID)

	if err := a.client.Del(*a.ctx, key).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	return nil
}

// ExistsToken checks if a token exists for the given user ID
func (a *DiscordTokenRedisAdapter) ExistsToken(userID int64) (bool, error) {
	key := utils.GetDiscordTokenKey(userID)

	exists, err := a.client.Exists(*a.ctx, key).Result()
	if err != nil {
		return false, exception.ErrRedisGetError
	}

	return exists > 0, nil
}
