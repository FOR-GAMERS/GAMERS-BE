package adapter

import (
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type GameTeamDatabaseAdapter struct {
	db *gorm.DB
}

func NewGameTeamDatabaseAdapter(db *gorm.DB) *GameTeamDatabaseAdapter {
	return &GameTeamDatabaseAdapter{db: db}
}

func (a *GameTeamDatabaseAdapter) Save(gameTeam *domain.GameTeam) (*domain.GameTeam, error) {
	if err := a.db.Create(gameTeam).Error; err != nil {
		return nil, a.translateError(err)
	}
	return gameTeam, nil
}

func (a *GameTeamDatabaseAdapter) GetByID(gameTeamID int64) (*domain.GameTeam, error) {
	var gameTeam domain.GameTeam
	result := a.db.Where("game_team_id = ?", gameTeamID).First(&gameTeam)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &gameTeam, nil
}

func (a *GameTeamDatabaseAdapter) GetByGameID(gameID int64) ([]*domain.GameTeam, error) {
	var gameTeams []*domain.GameTeam
	result := a.db.Where("game_id = ?", gameID).
		Order("grade ASC NULLS LAST").
		Find(&gameTeams)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return gameTeams, nil
}

func (a *GameTeamDatabaseAdapter) GetByGameAndTeam(gameID, teamID int64) (*domain.GameTeam, error) {
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ? AND team_id = ?", gameID, teamID).First(&gameTeam)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &gameTeam, nil
}

func (a *GameTeamDatabaseAdapter) GetByGrade(gameID int64, grade int) (*domain.GameTeam, error) {
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ? AND grade = ?", gameID, grade).First(&gameTeam)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &gameTeam, nil
}

func (a *GameTeamDatabaseAdapter) Delete(gameTeamID int64) error {
	result := a.db.Where("game_team_id = ?", gameTeamID).Delete(&domain.GameTeam{})
	if result.Error != nil {
		return a.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrGameTeamNotFound
	}

	return nil
}

func (a *GameTeamDatabaseAdapter) DeleteByGameID(gameID int64) error {
	result := a.db.Where("game_id = ?", gameID).Delete(&domain.GameTeam{})
	if result.Error != nil {
		return a.translateError(result.Error)
	}

	return nil
}

func (a *GameTeamDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrGameTeamNotFound
	}

	if a.isDuplicateKeyError(err) {
		return exception.ErrGameTeamAlreadyExists
	}

	if a.isForeignKeyError(err) {
		return exception.ErrGameNotFound
	}

	if a.isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (a *GameTeamDatabaseAdapter) isDuplicateKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "1062") ||
		strings.Contains(errMsg, "duplicate key value") ||
		strings.Contains(errMsg, "23505")
}

func (a *GameTeamDatabaseAdapter) isForeignKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "foreign key constraint") ||
		strings.Contains(errMsg, "1452") ||
		strings.Contains(errMsg, "23503")
}

func (a *GameTeamDatabaseAdapter) isConnectionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "refused")
}
