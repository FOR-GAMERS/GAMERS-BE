package query

import (
	"GAMERS-BE/internal/user/domain"
	"errors"

	"gorm.io/gorm"
)

type MYSQLProfileQueryAdapter struct {
	db *gorm.DB
}

func NewMysqlProfileQueryAdapter(db *gorm.DB) *MYSQLProfileQueryAdapter {
	return &MYSQLProfileQueryAdapter{db: db}
}

func (p *MYSQLProfileQueryAdapter) FindById(id int64) (*domain.Profile, error) {
	var profile domain.Profile
	result := p.db.First(&profile, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, result.Error
	}

	return &profile, nil
}

func (p *MYSQLProfileQueryAdapter) FindByUserId(userId int64) (*domain.Profile, error) {
	var profile domain.Profile
	result := p.db.Where("user_id = ?", userId).First(&profile)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, result.Error
	}

	return &profile, nil
}
