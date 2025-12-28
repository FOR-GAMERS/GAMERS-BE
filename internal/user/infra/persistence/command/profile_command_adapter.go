package command

import (
	"GAMERS-BE/internal/user/domain"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
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
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrProfileAlreadyExists
		}

		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.ErrProfileAlreadyExists
		}

		var sqliteErr sqlite3.Error
		if errors.As(result.Error, &sqliteErr) &&
			(errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) || sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique) {
			return domain.ErrProfileAlreadyExists
		}

		return result.Error
	}
	return nil
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
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrProfileAlreadyExists
		}

		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.ErrProfileAlreadyExists
		}

		var sqliteErr sqlite3.Error
		if errors.As(result.Error, &sqliteErr) &&
			(errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) || sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique) {
			return domain.ErrProfileAlreadyExists
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
