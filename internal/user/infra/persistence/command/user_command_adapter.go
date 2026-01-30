package command

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user/domain"
	"errors"

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

func (r *MySQLUserRepository) UpdateUserInfo(user *domain.User) error {
	result := r.db.Model(&domain.User{}).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"username": user.Username,
			"tag":      user.Tag,
			"bio":      user.Bio,
			"avatar":   user.Avatar,
		})

	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			return exception.ErrUserAlreadyExists
		}
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

func (r *MySQLUserRepository) UpdateValorantInfo(user *domain.User) error {
	result := r.db.Model(&domain.User{}).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"riot_name":            user.RiotName,
			"riot_tag":             user.RiotTag,
			"region":               user.Region,
			"current_tier":         user.CurrentTier,
			"current_tier_patched": user.CurrentTierPatched,
			"elo":                  user.Elo,
			"ranking_in_tier":      user.RankingInTier,
			"peak_tier":            user.PeakTier,
			"peak_tier_patched":    user.PeakTierPatched,
			"valorant_updated_at":  user.ValorantUpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}

	return nil
}

func (r *MySQLUserRepository) ClearValorantInfo(userId int64) error {
	result := r.db.Model(&domain.User{}).
		Where("id = ?", userId).
		Updates(map[string]interface{}{
			"riot_name":            nil,
			"riot_tag":             nil,
			"region":               nil,
			"current_tier":         nil,
			"current_tier_patched": nil,
			"elo":                  nil,
			"ranking_in_tier":      nil,
			"peak_tier":            nil,
			"peak_tier_patched":    nil,
			"valorant_updated_at":  nil,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return exception.ErrUserNotFound
	}

	return nil
}
