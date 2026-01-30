package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user/domain"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type OAuth2UserAdapter struct {
	db *gorm.DB
}

func NewOAuth2UserAdapter(db *gorm.DB) *OAuth2UserAdapter {
	return &OAuth2UserAdapter{
		db: db,
	}
}

func (a *OAuth2UserAdapter) SaveRandomUser(user *domain.User) error {
	exists, err := a.isExistSameUsername(user)
	if err != nil {
		return err
	}

	if exists {
		return exception.ErrUserAlreadyExists
	}

	err = a.save(user)
	if err != nil {
		return err
	}

	return nil
}

func (a *OAuth2UserAdapter) save(user *domain.User) error {
	result := a.db.Create(user)
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

func (a *OAuth2UserAdapter) isExistSameUsername(user *domain.User) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
            SELECT 1 FROM users 
            WHERE username = ? AND tag = ? 
            LIMIT 1
        )
	`

	err := a.db.Raw(query, user.Username, user.Tag).Scan(&exists).Error

	if err != nil {
		return false, err
	}

	return exists, nil

}
