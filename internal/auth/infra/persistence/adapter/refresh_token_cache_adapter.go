package adapter

import (
	"GAMERS-BE/internal/auth/domain"
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenPrefix = "refresh_token:"
	userTokensPrefix   = "user_tokens:"
)

type RefreshTokenCacheAdapter struct {
	ctx        *context.Context
	repository *redis.Client
}

func NewRefreshTokenCacheAdapter(ctx *context.Context, repository *redis.Client) *RefreshTokenCacheAdapter {
	return &RefreshTokenCacheAdapter{
		ctx:        ctx,
		repository: repository,
	}
}

func (r RefreshTokenCacheAdapter) Save(token *domain.RefreshToken, ttl *int64) error {
	key := refreshTokenPrefix + token.Token

	data, err := json.Marshal(token)
	if err != nil {
		return exception.ErrRedisCannotSave
	}

	durationTTL := utils.ConvertIntToDuration(*ttl)

	if err := r.repository.Set(*r.ctx, key, data, durationTTL).Err(); err != nil {
		return exception.ErrRedisCannotSave
	}

	userTokensKey := fmt.Sprintf("%s%d", userTokensPrefix, token.UserID)
	if err := r.repository.SAdd(*r.ctx, userTokensKey, token.Token).Err(); err != nil {
		return exception.ErrRedisCannotSet
	}

	if err := r.repository.Expire(*r.ctx, userTokensKey, durationTTL).Err(); err != nil {
		return exception.ErrRedisCannotSetDuration
	}

	return nil
}

func (r RefreshTokenCacheAdapter) FindByToken(token *string) (*domain.RefreshToken, error) {
	key := refreshTokenPrefix + *token

	data, err := r.repository.Get(*r.ctx, key).Result()
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

func (r RefreshTokenCacheAdapter) ExistsByToken(token *string) (bool, error) {
	key := refreshTokenPrefix + *token

	exists, err := r.repository.Exists(*r.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	return exists > 0, nil
}

func (r RefreshTokenCacheAdapter) Delete(token *string) error {
	key := refreshTokenPrefix + *token

	refreshToken, err := r.FindByToken(token)
	if err != nil {
		// 토큰이 없으면 그냥 성공 처리
		return nil
	}

	if err := r.repository.Del(*r.ctx, key).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	userTokensKey := fmt.Sprintf("%s%d", userTokensPrefix, refreshToken.UserID)
	if err := r.repository.SRem(*r.ctx, userTokensKey, *token).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	return nil
}

func (r RefreshTokenCacheAdapter) DeleteByUserID(userID *int64) error {
	userTokensKey := fmt.Sprintf("%s%d", userTokensPrefix, *userID)

	tokens, err := r.repository.SMembers(*r.ctx, userTokensKey).Result()
	if err != nil {
		return exception.ErrRedisCannotGetToken
	}

	for _, token := range tokens {
		key := refreshTokenPrefix + token
		if err := r.repository.Del(*r.ctx, key).Err(); err != nil {
			return exception.ErrRedisCannotDelete
		}
	}

	if err := r.repository.Del(*r.ctx, userTokensKey).Err(); err != nil {
		return exception.ErrRedisCannotDelete
	}

	return nil
}
