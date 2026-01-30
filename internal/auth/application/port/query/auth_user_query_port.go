package query

import "github.com/FOR-GAMERS/GAMERS-BE/internal/user/domain"

type AuthUserQueryPort interface {
	FindByEmail(email string) (*domain.User, error)
}
