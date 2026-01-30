package domain

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
)

type MemberType string

const (
	MemberTypeStaff  MemberType = "STAFF"
	MemberTypeNormal MemberType = "NORMAL"
)

// IsValid checks if the MemberType is valid
func (mt MemberType) IsValid() bool {
	switch mt {
	case MemberTypeStaff, MemberTypeNormal:
		return true
	default:
		return false
	}
}

type LeaderType string

const (
	LeaderTypeLeader LeaderType = "LEADER"
	LeaderTypeMember LeaderType = "MEMBER"
)

type ContestMember struct {
	UserID     int64      `gorm:"column:user_id;primaryKey" json:"user_id"`
	ContestID  int64      `gorm:"column:contest_id;primaryKey" json:"contest_id"`
	MemberType MemberType `gorm:"column:member_type;type:varchar(16);not null" json:"member_type"`
	LeaderType LeaderType `gorm:"column:leader_type;type:varchar(8);not null" json:"leader_type"`
	Point      int        `gorm:"column:point;type:int;default:0" json:"point"`
}

func NewContestMemberAsLeader(userID, contestID int64) *ContestMember {
	return &ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: MemberTypeStaff,
		LeaderType: LeaderTypeLeader,
		Point:      0,
	}
}

func NewContestMember(userID, contestID int64, memberType MemberType, leaderType LeaderType) *ContestMember {
	return &ContestMember{
		UserID:     userID,
		ContestID:  contestID,
		MemberType: memberType,
		LeaderType: leaderType,
		Point:      0,
	}
}

func (cm *ContestMember) TableName() string {
	return "contests_members"
}

func (cm *ContestMember) IsLeader() bool {
	return cm.LeaderType == LeaderTypeLeader
}

func (cm *ContestMember) IsStaff() bool {
	return cm.MemberType == MemberTypeStaff
}

func (cm *ContestMember) IsValidMemberType() bool {
	switch cm.MemberType {
	case MemberTypeStaff, MemberTypeNormal:
		return true
	default:
		return false
	}
}

func (cm *ContestMember) IsValidLeaderType() bool {
	switch cm.LeaderType {
	case LeaderTypeLeader, LeaderTypeMember:
		return true
	default:
		return false
	}
}

func (cm *ContestMember) Validate() error {
	if cm.UserID == 0 {
		return exception.ErrInvalidUserID
	}

	if cm.ContestID == 0 {
		return exception.ErrInvalidContestID
	}

	if !cm.IsValidMemberType() {
		return exception.ErrInvalidMemberType
	}

	if !cm.IsValidLeaderType() {
		return exception.ErrInvalidLeaderType
	}

	if cm.Point < 0 {
		return exception.ErrInvalidPoint
	}

	return nil
}
