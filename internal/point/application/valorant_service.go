package application

import (
	"GAMERS-BE/internal/point/application/dto"
	"GAMERS-BE/internal/point/application/port"
	"GAMERS-BE/internal/point/domain"
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
		req.Immortal,
		req.Ascendant,
		req.Diamond,
		req.Platinum,
		req.Gold,
		req.Silver,
		req.Bronze,
		req.Iron,
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
