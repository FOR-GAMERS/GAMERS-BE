package adapter

import (
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type GameDatabaseAdapter struct {
	db *gorm.DB
}

func NewGameDatabaseAdapter(db *gorm.DB) *GameDatabaseAdapter {
	return &GameDatabaseAdapter{db: db}
}

func (a *GameDatabaseAdapter) Save(game *domain.Game) (*domain.Game, error) {
	if err := a.db.Create(game).Error; err != nil {
		return nil, a.translateError(err)
	}
	return game, nil
}

func (a *GameDatabaseAdapter) SaveBatch(games []*domain.Game) error {
	if len(games) == 0 {
		return nil
	}

	if err := a.db.Create(&games).Error; err != nil {
		return a.translateError(err)
	}
	return nil
}

func (a *GameDatabaseAdapter) GetByID(gameID int64) (*domain.Game, error) {
	var game domain.Game
	result := a.db.Where("game_id = ?", gameID).First(&game)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &game, nil
}

func (a *GameDatabaseAdapter) GetByContestID(contestID int64) ([]*domain.Game, error) {
	var games []*domain.Game
	result := a.db.Where("contest_id = ?", contestID).
		Order("round ASC, match_number ASC, created_at DESC").
		Find(&games)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return games, nil
}

func (a *GameDatabaseAdapter) GetByContestAndRound(contestID int64, round int) ([]*domain.Game, error) {
	var games []*domain.Game
	result := a.db.Where("contest_id = ? AND round = ?", contestID, round).
		Order("match_number ASC").
		Find(&games)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return games, nil
}

func (a *GameDatabaseAdapter) Update(game *domain.Game) error {
	result := a.db.Save(game)
	if result.Error != nil {
		return a.translateError(result.Error)
	}
	return nil
}

func (a *GameDatabaseAdapter) Delete(gameID int64) error {
	result := a.db.Where("game_id = ?", gameID).Delete(&domain.Game{})
	if result.Error != nil {
		return a.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrGameNotFound
	}

	return nil
}

func (a *GameDatabaseAdapter) DeleteByContestID(contestID int64) error {
	result := a.db.Where("contest_id = ?", contestID).Delete(&domain.Game{})
	if result.Error != nil {
		return a.translateError(result.Error)
	}
	return nil
}

func (a *GameDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrGameNotFound
	}

	if a.isDuplicateKeyError(err) {
		return exception.ErrGameNotFound // or create specific error
	}

	if a.isForeignKeyError(err) {
		return exception.ErrContestNotFound
	}

	if a.isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (a *GameDatabaseAdapter) isDuplicateKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "1062") ||
		strings.Contains(errMsg, "duplicate key value") ||
		strings.Contains(errMsg, "23505")
}

func (a *GameDatabaseAdapter) isForeignKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "foreign key constraint") ||
		strings.Contains(errMsg, "1452") ||
		strings.Contains(errMsg, "23503")
}

func (a *GameDatabaseAdapter) isConnectionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "refused")
}
