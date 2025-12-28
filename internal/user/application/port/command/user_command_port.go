package command

import "GAMERS-BE/internal/user/domain"

type UserCommandPort interface {
	Save(user *domain.User) error
	Update(user *domain.User) error
	DeleteById(id int64) error
}
