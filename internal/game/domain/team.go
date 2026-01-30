package domain

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"time"
)

// Team represents a team registered for a contest
type Team struct {
	TeamID     int64     `gorm:"column:team_id;primaryKey;autoIncrement" json:"team_id"`
	ContestID  int64     `gorm:"column:contest_id;type:bigint;not null" json:"contest_id"`
	TeamName   string    `gorm:"column:team_name;type:varchar(50);not null" json:"team_name"`
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ModifiedAt time.Time `gorm:"column:modified_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"modified_at"`
}

func NewTeam(contestID int64, teamName string) *Team {
	return &Team{
		ContestID: contestID,
		TeamName:  teamName,
	}
}

func (t *Team) TableName() string {
	return "teams"
}

func (t *Team) Validate() error {
	if t.ContestID <= 0 {
		return exception.ErrInvalidContestID
	}

	if t.TeamName == "" {
		return exception.ErrInvalidTeamName
	}

	if len(t.TeamName) > 50 {
		return exception.ErrTeamNameTooLong
	}

	return nil
}

// TeamMemberType represents the type of team membership
type TeamMemberType string

const (
	TeamMemberTypeMember TeamMemberType = "MEMBER"
	TeamMemberTypeLeader TeamMemberType = "LEADER"
)

func (t TeamMemberType) IsValid() bool {
	switch t {
	case TeamMemberTypeMember, TeamMemberTypeLeader:
		return true
	default:
		return false
	}
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID         int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TeamID     int64          `gorm:"column:team_id;type:bigint;not null" json:"team_id"`
	UserID     int64          `gorm:"column:user_id;type:bigint;not null" json:"user_id"`
	MemberType TeamMemberType `gorm:"column:member_type;type:varchar(16);not null" json:"member_type"`
	JoinedAt   time.Time      `gorm:"column:joined_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"joined_at"`
}

func NewTeamMember(teamID, userID int64, memberType TeamMemberType) *TeamMember {
	return &TeamMember{
		TeamID:     teamID,
		UserID:     userID,
		MemberType: memberType,
	}
}

func NewTeamMemberAsLeader(teamID, userID int64) *TeamMember {
	return &TeamMember{
		TeamID:     teamID,
		UserID:     userID,
		MemberType: TeamMemberTypeLeader,
	}
}

func NewTeamMemberAsMember(teamID, userID int64) *TeamMember {
	return &TeamMember{
		TeamID:     teamID,
		UserID:     userID,
		MemberType: TeamMemberTypeMember,
	}
}

func (tm *TeamMember) TableName() string {
	return "team_members"
}

func (tm *TeamMember) IsLeader() bool {
	return tm.MemberType == TeamMemberTypeLeader
}

func (tm *TeamMember) IsMember() bool {
	return tm.MemberType == TeamMemberTypeMember
}

// CanInvite checks if this member can invite others
// Both Leader and Member can invite
func (tm *TeamMember) CanInvite() bool {
	return tm.MemberType == TeamMemberTypeLeader || tm.MemberType == TeamMemberTypeMember
}

// CanKick checks if this member can kick others
// Only Leader can kick members
func (tm *TeamMember) CanKick() bool {
	return tm.MemberType == TeamMemberTypeLeader
}

// CanDeleteTeam checks if this member can delete the team
// Only Leader can delete the team
func (tm *TeamMember) CanDeleteTeam() bool {
	return tm.MemberType == TeamMemberTypeLeader
}

func (tm *TeamMember) Validate() error {
	if tm.TeamID <= 0 {
		return exception.ErrInvalidTeamID
	}

	if tm.UserID <= 0 {
		return exception.ErrInvalidUserID
	}

	if !tm.MemberType.IsValid() {
		return exception.ErrInvalidTeamMemberType
	}

	return nil
}
