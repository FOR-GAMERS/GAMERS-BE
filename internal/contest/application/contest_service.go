package application

import (
	"GAMERS-BE/internal/contest/application/dto"
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/contest/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/exception"
)

type ContestService struct {
	repository port.ContestDatabasePort
}

func NewContestService(repository port.ContestDatabasePort) *ContestService {
	return &ContestService{
		repository: repository,
	}
}

func (c *ContestService) SaveContest(req *dto.CreateContestRequest) (*domain.Contest, error) {
	contest := *domain.NewContestInstance(req.Title, req.Description, req.MaxTeamCount, req.TotalPoint, req.ContestType, req.StartedAt, req.EndedAt, req.AutoStart)
	savedContest, err := c.repository.Save(&contest)
	if err != nil {
		return nil, err
	}

	return savedContest, nil
}

func (c *ContestService) GetContestById(id int64) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(id)

	if err != nil {
		return nil, err
	}

	return contest, nil
}

func (c *ContestService) GetAllContests(offset, limit int, sortReq *commonDto.SortRequest) ([]domain.Contest, int64, error) {
	contests, totalCount, err := c.repository.GetContests(offset, limit, sortReq)

	if err != nil {
		return nil, 0, err
	}

	return contests, totalCount, nil
}

func (c *ContestService) UpdateContest(id int64, req *dto.UpdateContestRequest) (*domain.Contest, error) {
	contest, err := c.repository.GetContestById(id)

	if err != nil {
		return nil, err
	}

	if !req.HasChanges() {
		return nil, exception.ErrContestNoChanges
	}

	if err = req.Validate(); err != nil {
		return nil, err
	}

	req.ApplyTo(contest)

	err = c.repository.UpdateContest(contest)

	if err != nil {
		return nil, err
	}

	return contest, nil
}

func (c *ContestService) DeleteContestById(id int64) error {
	return c.repository.DeleteContestById(id)
}
