package domain

import "github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"

// GameTeam represents a many-to-many relationship between Game and Team
type GameTeam struct {
	GameTeamID int64 `gorm:"column:game_team_id;primaryKey;autoIncrement" json:"game_team_id"`
	GameID     int64 `gorm:"column:game_id;type:bigint;not null" json:"game_id"`
	TeamID     int64 `gorm:"column:team_id;type:bigint;not null" json:"team_id"`
	Grade      *int  `gorm:"column:grade;type:int" json:"grade,omitempty"`
}

func NewGameTeam(gameID, teamID int64) *GameTeam {
	return &GameTeam{
		GameID: gameID,
		TeamID: teamID,
		Grade:  nil,
	}
}

func NewGameTeamWithGrade(gameID, teamID int64, grade int) *GameTeam {
	return &GameTeam{
		GameID: gameID,
		TeamID: teamID,
		Grade:  &grade,
	}
}

func (gt *GameTeam) TableName() string {
	return "game_teams"
}

func (gt *GameTeam) SetGrade(grade int) {
	gt.Grade = &grade
}

func (gt *GameTeam) ClearGrade() {
	gt.Grade = nil
}

func (gt *GameTeam) HasGrade() bool {
	return gt.Grade != nil
}

// ValidateGrade checks if the grade is within valid range (1 to maxTeamCount)
func (gt *GameTeam) ValidateGrade(maxTeamCount int) error {
	if gt.Grade == nil {
		return nil
	}

	if *gt.Grade < 1 {
		return exception.ErrInvalidGradeMin
	}

	if maxTeamCount > 0 && *gt.Grade > maxTeamCount {
		return exception.ErrGradeExceedsMaxTeamCount
	}

	return nil
}

func (gt *GameTeam) Validate() error {
	if gt.GameID <= 0 {
		return exception.ErrInvalidGameID
	}

	if gt.TeamID <= 0 {
		return exception.ErrInvalidTeamID
	}

	return nil
}

// ValidateWithMaxTeamCount validates the game team including grade constraint
func (gt *GameTeam) ValidateWithMaxTeamCount(maxTeamCount int) error {
	if err := gt.Validate(); err != nil {
		return err
	}

	if err := gt.ValidateGrade(maxTeamCount); err != nil {
		return err
	}

	return nil
}
