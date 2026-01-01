package command

import (
	"GAMERS-BE/internal/auth/domain"
	"GAMERS-BE/internal/global/exception"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenPrefix = "refresh_token:"
	userTokensPrefix   = "user_tokens:"
)

type RefreshTokenRedisCommandAdapter struct {
	client *redis.Client
}

func NewRefreshTokenRedisCommandAdapter(client *redis.Client) *RefreshTokenRedisCommandAdapter {
	return &RefreshTokenRedisCommandAdapter{
		client: client,
	}
}

func (a *RefreshTokenRedisCommandAdapter) Save(ctx context.Context, token *domain.RefreshToken, ttl time.Duration) error {
	key := refreshTokenPrefix + token.Token

	data, err := json.Marshal(token)
	if err != nil {
		return exception.ErrRedisCannotSave
	}

	if err := a.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return exception.ErrRedisCannotSave
	}

	userTokensKey := fmt.Sprintf("%s%d", userTokensPrefix, token.UserID)
	if err := a.client.SAdd(ctx, userTokensKey, token.Token).Err(); err != nil {
		return exception.ErrRedisCannotSet
	}

	if err := a.client.Expire(ctx, userTokensKey, ttl).Err(); err != nil {
		return exception.ErrRedisCannotSetDuration
	}

	return nil
}

func (a *RefreshTokenRedisCommandAdapter) Delete(ctx context.Context, token string) error {
	key := refreshTokenPrefix + token

	refreshToken, err := a.findByToken(ctx, token)
	if err != nil {
		// 토큰이 없으면 그냥 성공 처리
		return nil
	}

	if err := a.client.Del(ctx, key).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	userTokensKey := fmt.Sprintf("%s%d", userTokensPrefix, refreshToken.UserID)
	if err := a.client.SRem(ctx, userTokensKey, token).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	return nil
}

func (a *RefreshTokenRedisCommandAdapter) DeleteByUserID(ctx context.Context, userID uint) error {
	userTokensKey := fmt.Sprintf("%s%d", userTokensPrefix, userID)

	tokens, err := a.client.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return exception.ErrRedisCannotGetToken
	}

	for _, token := range tokens {
		key := refreshTokenPrefix + token
		if err := a.client.Del(ctx, key).Err(); err != nil {
			return exception.ErrRedisCannotDelete
		}
	}

	if err := a.client.Del(ctx, userTokensKey).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	return nil
}

func (a *RefreshTokenRedisCommandAdapter) findByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
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
