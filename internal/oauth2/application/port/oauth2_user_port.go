package port

import "github.com/FOR-GAMERS/GAMERS-BE/internal/user/domain"

type OAuth2UserPort interface {
	SaveRandomUser(user *domain.User) error
	FindById(id int64) (*domain.User, error)
}
