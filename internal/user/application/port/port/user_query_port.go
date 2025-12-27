package port

import "GAMERS-BE/internal/user/domain"

type UserQueryPort interface {
	FindById(id int64) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
}
