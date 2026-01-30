package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/game/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"errors"

	"gorm.io/gorm"
)

// MatchResultDatabaseAdapter implements MatchResultDatabasePort using GORM
type MatchResultDatabaseAdapter struct {
	db *gorm.DB
}

func NewMatchResultDatabaseAdapter(db *gorm.DB) *MatchResultDatabaseAdapter {
	return &MatchResultDatabaseAdapter{db: db}
}

func (a *MatchResultDatabaseAdapter) Save(result *domain.MatchResult) (*domain.MatchResult, error) {
	if err := a.db.Create(result).Error; err != nil {
		return nil, a.translateError(err)
	}
	return result, nil
}

func (a *MatchResultDatabaseAdapter) GetByGameID(gameID int64) (*domain.MatchResult, error) {
	var result domain.MatchResult
	if err := a.db.Where("game_id = ?", gameID).First(&result).Error; err != nil {
		return nil, a.translateError(err)
	}
	return &result, nil
}

func (a *MatchResultDatabaseAdapter) SavePlayerStats(stats []*domain.MatchPlayerStat) error {
	if len(stats) == 0 {
		return nil
	}
	if err := a.db.Create(&stats).Error; err != nil {
		return err
	}
	return nil
}

func (a *MatchResultDatabaseAdapter) GetPlayerStatsByMatchResult(matchResultID int64) ([]*domain.MatchPlayerStat, error) {
	var stats []*domain.MatchPlayerStat
	if err := a.db.Where("match_result_id = ?", matchResultID).Find(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (a *MatchResultDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrMatchResultNotFound
	}
	return err
}
