package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/domain"
)

type ValorantService struct {
	scoreTableRepository port.ValorantScoreTableDatabasePort
}

func NewValorantService(
	scoreTableRepository port.ValorantScoreTableDatabasePort,
) *ValorantService {
	return &ValorantService{
		scoreTableRepository: scoreTableRepository,
	}
}

// CreateScoreTable creates a new valorant score table
func (s *ValorantService) CreateScoreTable(req *dto.CreateValorantScoreTableDto) (*domain.ValorantScoreTable, error) {
	scoreTable := domain.NewValorantScoreTable(
		req.Radiant,
		req.Immortal3, req.Immortal2, req.Immortal1,
		req.Ascendant3, req.Ascendant2, req.Ascendant1,
		req.Diamond3, req.Diamond2, req.Diamond1,
		req.Platinum3, req.Platinum2, req.Platinum1,
		req.Gold3, req.Gold2, req.Gold1,
		req.Silver3, req.Silver2, req.Silver1,
		req.Bronze3, req.Bronze2, req.Bronze1,
		req.Iron3, req.Iron2, req.Iron1,
	)

	return s.scoreTableRepository.Save(scoreTable)
}

// GetScoreTable returns a score table by ID
func (s *ValorantService) GetScoreTable(scoreTableID int64) (*domain.ValorantScoreTable, error) {
	return s.scoreTableRepository.GetByID(scoreTableID)
}

// GetAllScoreTables returns all score tables
func (s *ValorantService) GetAllScoreTables() ([]*domain.ValorantScoreTable, error) {
	return s.scoreTableRepository.GetAll()
}

// DeleteScoreTable deletes a score table by ID
func (s *ValorantService) DeleteScoreTable(scoreTableID int64) error {
	return s.scoreTableRepository.Delete(scoreTableID)
}
