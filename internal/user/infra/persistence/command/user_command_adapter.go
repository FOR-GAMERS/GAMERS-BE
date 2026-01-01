package command

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
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
			return exception.ErrUserAlreadyExists
		}

		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return exception.ErrUserAlreadyExists
		}

		// SQLite UNIQUE constraint error
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return exception.ErrUserAlreadyExists
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
		return exception.ErrUserNotFound
	}

	return nil
}

func (r *MySQLUserRepository) DeleteById(id int64) error {
	result := r.db.Delete(&domain.User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}

	return nil
}
