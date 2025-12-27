package command

import (
	"GAMERS-BE/internal/user/domain"
	"errors"

	"gorm.io/gorm"
)

type MYSQLProfileCommandAdapter struct {
	db *gorm.DB
}

func NewMysqlProfileCommandAdapter(db *gorm.DB) *MYSQLProfileCommandAdapter {
	return &MYSQLProfileCommandAdapter{db: db}
}

func (r *MYSQLProfileCommandAdapter) Save(profile *domain.Profile) error {
	result := r.db.Create(profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
	}
	return result.Error
}

func (r *MYSQLProfileCommandAdapter) Update(profile *domain.Profile) error {
	result := r.db.Model(&domain.Profile{}).
		Where("profile_id = ?", profile.Id).
		Updates(map[string]interface{}{
			"username":    profile.Username,
			"tag":         profile.Tag,
			"bio":         profile.Bio,
			"profile_url": profile.Avatar,
		})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}

func (r *MYSQLProfileCommandAdapter) DeleteById(id int64) error {
	result := r.db.Delete(&domain.Profile{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}
