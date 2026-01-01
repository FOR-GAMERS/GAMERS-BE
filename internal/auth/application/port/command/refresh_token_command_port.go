package command

import (
	"context"
	"time"

	"GAMERS-BE/internal/auth/domain"
)

type RefreshTokenCommandPort interface {
	Save(ctx context.Context, token *domain.RefreshToken, ttl time.Duration) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID uint) error
}
