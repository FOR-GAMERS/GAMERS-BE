package adapter

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/domain"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type ValorantScoreTableDatabaseAdapter struct {
	db *gorm.DB
}

func NewValorantScoreTableDatabaseAdapter(db *gorm.DB) *ValorantScoreTableDatabaseAdapter {
	return &ValorantScoreTableDatabaseAdapter{
		db: db,
	}
}

func (a *ValorantScoreTableDatabaseAdapter) Save(scoreTable *domain.ValorantScoreTable) (*domain.ValorantScoreTable, error) {
	if err := a.db.Create(scoreTable).Error; err != nil {
		return nil, a.translateError(err)
	}
	return scoreTable, nil
}

func (a *ValorantScoreTableDatabaseAdapter) GetByID(scoreTableID int64) (*domain.ValorantScoreTable, error) {
	var scoreTable domain.ValorantScoreTable
	result := a.db.Where("score_table_id = ?", scoreTableID).First(&scoreTable)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &scoreTable, nil
}

func (a *ValorantScoreTableDatabaseAdapter) GetAll() ([]*domain.ValorantScoreTable, error) {
	var scoreTables []*domain.ValorantScoreTable
	result := a.db.Order("created_at DESC").Find(&scoreTables)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return scoreTables, nil
}

func (a *ValorantScoreTableDatabaseAdapter) Update(scoreTable *domain.ValorantScoreTable) error {
	result := a.db.Save(scoreTable)
	if result.Error != nil {
		return a.translateError(result.Error)
	}
	return nil
}

func (a *ValorantScoreTableDatabaseAdapter) Delete(scoreTableID int64) error {
	result := a.db.Where("score_table_id = ?", scoreTableID).Delete(&domain.ValorantScoreTable{})
	if result.Error != nil {
		return a.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrScoreTableNotFound
	}

	return nil
}

func (a *ValorantScoreTableDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrScoreTableNotFound
	}

	if a.isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (a *ValorantScoreTableDatabaseAdapter) isConnectionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "refused")
}
