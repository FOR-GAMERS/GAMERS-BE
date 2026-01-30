package port

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/domain"
)

type ValorantScoreTableDatabasePort interface {
	Save(scoreTable *domain.ValorantScoreTable) (*domain.ValorantScoreTable, error)
	GetByID(scoreTableID int64) (*domain.ValorantScoreTable, error)
	GetAll() ([]*domain.ValorantScoreTable, error)
	Update(scoreTable *domain.ValorantScoreTable) error
	Delete(scoreTableID int64) error
}
