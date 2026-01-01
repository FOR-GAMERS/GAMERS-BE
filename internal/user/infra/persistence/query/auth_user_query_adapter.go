package query

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/user/domain"

	"gorm.io/gorm"
)

type AuthUserQueryAdapter struct {
	db *gorm.DB
}

func NewAuthUserQueryAdapter(db *gorm.DB) *AuthUserQueryAdapter {
	return &AuthUserQueryAdapter{db: db}
}

func (adapter *AuthUserQueryAdapter) FindByEmail(email string) (*domain.User, error) {
	var user domain.User

	result := adapter.db.Where("email = ?", email).First(&user)

	if result.RowsAffected == 0 {
		return nil, exception.ErrUserNotFound
	}

	return &user, nil
}
