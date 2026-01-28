package adapter

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type UserDatabaseAdapter struct {
	db *gorm.DB
}

func NewUserDatabaseAdapter(db *gorm.DB) *UserDatabaseAdapter {
	return &UserDatabaseAdapter{
		db: db,
	}
}

func (r *UserDatabaseAdapter) Save(user *domain.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return exception.ErrUserAlreadyExists
		}

		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return exception.ErrUserAlreadyExists
		}

		return result.Error
	}

	return nil
}

func (r *UserDatabaseAdapter) Update(user *domain.User) error {
	result := r.db.Model(&domain.User{}).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"password": user.Password,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}

	return nil
}

func (r *UserDatabaseAdapter) DeleteById(id int64) error {
	result := r.db.Delete(&domain.User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}

	return nil
}

func (r *UserDatabaseAdapter) FindById(id int64) (*domain.User, error) {
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

func (r *UserDatabaseAdapter) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}
