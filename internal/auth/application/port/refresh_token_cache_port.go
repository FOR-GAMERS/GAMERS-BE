package port

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/domain"
)

type RefreshTokenCachePort interface {
	Save(token *domain.RefreshToken, ttl *int64) error
	FindByToken(token *string) (*domain.RefreshToken, error)
	ExistsByToken(token *string) (bool, error)
	Delete(token *string) error
	DeleteByUserID(userID *int64) error
}
