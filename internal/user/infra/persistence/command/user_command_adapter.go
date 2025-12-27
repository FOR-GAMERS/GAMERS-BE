package command

import (
	"GAMERS-BE/internal/user/domain"
	"errors"

	"gorm.io/gorm"
)

type MySQLUserRepository struct {
	db *gorm.DB
}

func NewMySQLUserRepository(db *gorm.DB) *MySQLUserRepository {
	return &MySQLUserRepository{
		db: db,
	}
}

func (r *MySQLUserRepository) Save(user *domain.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrUserAlreadyExists
		}
		// SQLite unique constraint error
		if result.Error.Error() == "UNIQUE constraint failed: users.email" {
			return domain.ErrUserAlreadyExists
		}
		return result.Error
	}

	return nil
}

func (r *MySQLUserRepository) Update(user *domain.User) error {
	result := r.db.Model(&domain.User{}).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"password": user.Password,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *MySQLUserRepository) DeleteById(id int64) error {
	result := r.db.Delete(&domain.User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
