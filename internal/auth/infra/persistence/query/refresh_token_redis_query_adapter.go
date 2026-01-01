package query

import (
	"GAMERS-BE/internal/auth/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenPrefix = "refresh_token:"
)

type RefreshTokenRedisQueryAdapter struct {
	client *redis.Client
}

func NewRefreshTokenRedisQueryAdapter(client *redis.Client) *RefreshTokenRedisQueryAdapter {
	return &RefreshTokenRedisQueryAdapter{
		client: client,
	}
}

func (a *RefreshTokenRedisQueryAdapter) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	key := refreshTokenPrefix + token

	data, err := a.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	var refreshToken domain.RefreshToken
	if err := json.Unmarshal([]byte(data), &refreshToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (a *RefreshTokenRedisQueryAdapter) ExistsByToken(ctx context.Context, token string) (bool, error) {
	key := refreshTokenPrefix + token

	exists, err := a.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	return exists > 0, nil
}
