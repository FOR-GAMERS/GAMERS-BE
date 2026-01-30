package port

import "github.com/FOR-GAMERS/GAMERS-BE/internal/user/domain"

type UserDatabasePort interface {
	Save(user *domain.User) error
	Update(user *domain.User) error
	DeleteById(id int64) error
	FindById(id int64) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
}
