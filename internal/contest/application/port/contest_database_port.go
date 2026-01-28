package port

import (
	"GAMERS-BE/internal/contest/domain"
	"GAMERS-BE/internal/global/common/dto"
)

type ContestDatabasePort interface {
	Save(contest *domain.Contest) (*domain.Contest, error)

	GetContestById(contestId int64) (*domain.Contest, error)

	GetContests(offset, limit int, sortReq *dto.SortRequest, title *string) ([]domain.Contest, int64, error)

	DeleteContestById(contestId int64) error

	UpdateContest(contest *domain.Contest) error
}
