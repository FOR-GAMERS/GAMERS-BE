package query

import (
	"context"

	"GAMERS-BE/internal/auth/domain"
)

type RefreshTokenQueryPort interface {
	FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	ExistsByToken(ctx context.Context, token string) (bool, error)
}
