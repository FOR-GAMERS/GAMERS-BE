package port

import "github.com/FOR-GAMERS/GAMERS-BE/internal/user/domain"

type OAuth2UserPort interface {
	SaveRandomUser(user *domain.User) error
}
