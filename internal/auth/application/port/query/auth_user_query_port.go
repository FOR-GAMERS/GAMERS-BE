package query

import "GAMERS-BE/internal/user/domain"

type AuthUserQueryPort interface {
	FindByEmail(email string) (*domain.User, error)
}
