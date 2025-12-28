package port

import "GAMERS-BE/internal/user/domain"

type ProfileQueryPort interface {
	FindById(id int64) (*domain.Profile, error)
	FindByUserId(userId int64) (*domain.Profile, error)
}
