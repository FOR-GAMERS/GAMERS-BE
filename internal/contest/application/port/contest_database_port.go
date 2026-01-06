package port

import "GAMERS-BE/internal/contest/domain"

type ContestDatabasePort interface {
	Save(contest *domain.Contest) (*domain.Contest, error)

	GetContestById(contestId int64) (*domain.Contest, error)

	GetContests() ([]domain.Contest, error)

	DeleteContestById(contestId int64) error

	UpdateContest(contest *domain.Contest) error
}
