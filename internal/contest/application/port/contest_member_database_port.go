package port

import (
	"GAMERS-BE/internal/contest/domain"
	commonDto "GAMERS-BE/internal/global/common/dto"
)

// ContestMemberWithUser represents a contest member with user information
type ContestMemberWithUser struct {
	UserID             int64             `json:"user_id"`
	ContestID          int64             `json:"contest_id"`
	MemberType         domain.MemberType `json:"member_type"`
	LeaderType         domain.LeaderType `json:"leader_type"`
	Point              int               `json:"point"`
	Username           string            `json:"username"`
	Tag                string            `json:"tag"`
	Avatar             string            `json:"avatar"`
	DiscordId          *string           `json:"discord_id"`
	DiscordAvatar      *string           `json:"discord_avatar"`
	CurrentTier        *int              `json:"current_tier"`
	CurrentTierPatched *string           `json:"current_tier_patched"`
	PeakTier           *int              `json:"peak_tier"`
	PeakTierPatched    *string           `json:"peak_tier_patched"`
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
