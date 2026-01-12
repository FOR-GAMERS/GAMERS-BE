package port

import "GAMERS-BE/internal/contest/domain"

type ContestMemberDatabasePort interface {
	Save(member *domain.ContestMember) error
	DeleteById(contestId, userId int64) error
	GetByContestAndUser(contestId, userId int64) (*domain.ContestMember, error)
	GetMembersByContest(contestId int64) ([]*domain.ContestMember, error)
	SaveBatch(members []*domain.ContestMember) error
}
