package query

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"
	"errors"

	"gorm.io/gorm"
)

type MYSQLUserRepository struct {
	db *gorm.DB
}

func NewMysqlUserRepository(db *gorm.DB) *MYSQLUserRepository {
	return &MYSQLUserRepository{db: db}
}

func (r *MYSQLUserRepository) FindById(id int64) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

func (r *MYSQLUserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
	}

	return &user, nil
}
