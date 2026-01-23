package port

import (
	"GAMERS-BE/internal/contest/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
)

// ContestMemberWithUser represents a contest member with user information
type ContestMemberWithUser struct {
	UserID     int64             `json:"user_id"`
	ContestID  int64             `json:"contest_id"`
	MemberType domain.MemberType `json:"member_type"`
	LeaderType domain.LeaderType `json:"leader_type"`
	Point      int               `json:"point"`
	Username   string            `json:"username"`
	Tag        string            `json:"tag"`
	Avatar     string            `json:"avatar"`
}

// ContestWithMembership represents a contest with the user's membership info
type ContestWithMembership struct {
	domain.Contest
	MemberType domain.MemberType `json:"member_type"`
	LeaderType domain.LeaderType `json:"leader_type"`
	Point      int               `json:"point"`
}

type ContestMemberDatabasePort interface {
	Save(member *domain.ContestMember) error
	DeleteById(contestId, userId int64) error
	GetByContestAndUser(contestId, userId int64) (*domain.ContestMember, error)
	GetMembersByContest(contestId int64) ([]*domain.ContestMember, error)
	SaveBatch(members []*domain.ContestMember) error
	GetMembersWithUserByContest(contestId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest) ([]*ContestMemberWithUser, int64, error)
	GetContestsByUserId(userId int64, pagination *commonDto.PaginationRequest, sort *commonDto.SortRequest, status *domain.ContestStatus) ([]*ContestWithMembership, int64, error)
	UpdateMemberType(contestId, userId int64, memberType domain.MemberType) error
}
