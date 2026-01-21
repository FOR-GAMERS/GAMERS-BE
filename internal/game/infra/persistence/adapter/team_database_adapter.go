package adapter

import (
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type TeamDatabaseAdapter struct {
	db *gorm.DB
}

func NewTeamDatabaseAdapter(db *gorm.DB) *TeamDatabaseAdapter {
	return &TeamDatabaseAdapter{db: db}
}

// Team operations

func (a *TeamDatabaseAdapter) Save(team *domain.Team) (*domain.Team, error) {
	if err := team.Validate(); err != nil {
		return nil, err
	}

	if err := a.db.Create(team).Error; err != nil {
		return nil, a.translateError(err)
	}

	return team, nil
}

func (a *TeamDatabaseAdapter) GetByID(teamID int64) (*domain.Team, error) {
	var team domain.Team
	result := a.db.Where("team_id = ?", teamID).First(&team)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &team, nil
}

func (a *TeamDatabaseAdapter) GetByContestID(contestID int64) ([]*domain.Team, error) {
	var teams []*domain.Team
	result := a.db.Where("contest_id = ?", contestID).Find(&teams)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return teams, nil
}

func (a *TeamDatabaseAdapter) GetByContestAndName(contestID int64, teamName string) (*domain.Team, error) {
	var team domain.Team
	result := a.db.Where("contest_id = ? AND team_name = ?", contestID, teamName).First(&team)

	if result.Error != nil {
		return nil, a.translateError(result.Error)
	}

	return &team, nil
}

func (a *TeamDatabaseAdapter) CountByContestID(contestID int64) (int, error) {
	var count int64
	result := a.db.Model(&domain.Team{}).Where("contest_id = ?", contestID).Count(&count)

	if result.Error != nil {
		return 0, a.translateError(result.Error)
	}

	return int(count), nil
}

func (a *TeamDatabaseAdapter) Update(team *domain.Team) error {
	if err := team.Validate(); err != nil {
		return err
	}

	result := a.db.Save(team)
	if result.Error != nil {
		return a.translateError(result.Error)
	}

	return nil
}

func (a *TeamDatabaseAdapter) Delete(teamID int64) error {
	result := a.db.Where("team_id = ?", teamID).Delete(&domain.Team{})

	if result.Error != nil {
		return a.translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrTeamNotFound
	}

	return nil
}

func (a *TeamDatabaseAdapter) DeleteByContestID(contestID int64) error {
	result := a.db.Where("contest_id = ?", contestID).Delete(&domain.Team{})

	if result.Error != nil {
		return a.translateError(result.Error)
	}

	return nil
}

// TeamMember operations

func (a *TeamDatabaseAdapter) SaveMember(member *domain.TeamMember) (*domain.TeamMember, error) {
	if err := member.Validate(); err != nil {
		return nil, err
	}

	if err := a.db.Create(member).Error; err != nil {
		return nil, a.translateMemberError(err)
	}

	return member, nil
}

func (a *TeamDatabaseAdapter) SaveMemberBatch(members []*domain.TeamMember) error {
	if len(members) == 0 {
		return nil
	}

	for _, member := range members {
		if err := member.Validate(); err != nil {
			return err
		}
	}

	if err := a.db.Create(&members).Error; err != nil {
		return a.translateMemberError(err)
	}

	return nil
}

func (a *TeamDatabaseAdapter) GetMemberByID(id int64) (*domain.TeamMember, error) {
	var member domain.TeamMember
	result := a.db.First(&member, id)

	if result.Error != nil {
		return nil, a.translateMemberError(result.Error)
	}

	return &member, nil
}

func (a *TeamDatabaseAdapter) GetMembersByTeamID(teamID int64) ([]*domain.TeamMember, error) {
	var members []*domain.TeamMember
	result := a.db.Where("team_id = ?", teamID).Find(&members)

	if result.Error != nil {
		return nil, a.translateMemberError(result.Error)
	}

	return members, nil
}

func (a *TeamDatabaseAdapter) GetMemberByTeamAndUser(teamID, userID int64) (*domain.TeamMember, error) {
	var member domain.TeamMember
	result := a.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member)

	if result.Error != nil {
		return nil, a.translateMemberError(result.Error)
	}

	return &member, nil
}

func (a *TeamDatabaseAdapter) GetMemberCountByTeamID(teamID int64) (int, error) {
	var count int64
	result := a.db.Model(&domain.TeamMember{}).Where("team_id = ?", teamID).Count(&count)

	if result.Error != nil {
		return 0, a.translateMemberError(result.Error)
	}

	return int(count), nil
}

func (a *TeamDatabaseAdapter) GetLeaderByTeamID(teamID int64) (*domain.TeamMember, error) {
	var member domain.TeamMember
	result := a.db.Where("team_id = ? AND member_type = ?", teamID, domain.TeamMemberTypeLeader).First(&member)

	if result.Error != nil {
		return nil, a.translateMemberError(result.Error)
	}

	return &member, nil
}

func (a *TeamDatabaseAdapter) UpdateMember(member *domain.TeamMember) error {
	if err := member.Validate(); err != nil {
		return err
	}

	result := a.db.Save(member)
	if result.Error != nil {
		return a.translateMemberError(result.Error)
	}

	return nil
}

func (a *TeamDatabaseAdapter) DeleteMember(id int64) error {
	result := a.db.Delete(&domain.TeamMember{}, id)

	if result.Error != nil {
		return a.translateMemberError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrTeamMemberNotFound
	}

	return nil
}

func (a *TeamDatabaseAdapter) DeleteMemberByTeamAndUser(teamID, userID int64) error {
	result := a.db.Where("team_id = ? AND user_id = ?", teamID, userID).Delete(&domain.TeamMember{})

	if result.Error != nil {
		return a.translateMemberError(result.Error)
	}

	if result.RowsAffected == 0 {
		return exception.ErrTeamMemberNotFound
	}

	return nil
}

func (a *TeamDatabaseAdapter) DeleteAllMembersByTeamID(teamID int64) error {
	result := a.db.Where("team_id = ?", teamID).Delete(&domain.TeamMember{})

	if result.Error != nil {
		return a.translateMemberError(result.Error)
	}

	return nil
}

// Contest-based queries

func (a *TeamDatabaseAdapter) GetTeamsByContestWithMembers(contestID int64) ([]*port.TeamWithMembers, error) {
	teams, err := a.GetByContestID(contestID)
	if err != nil {
		return nil, err
	}

	result := make([]*port.TeamWithMembers, 0, len(teams))
	for _, team := range teams {
		members, err := a.GetMembersByTeamID(team.TeamID)
		if err != nil {
			return nil, err
		}
		result = append(result, &port.TeamWithMembers{
			Team:    team,
			Members: members,
		})
	}

	return result, nil
}

func (a *TeamDatabaseAdapter) GetUserTeamInContest(contestID, userID int64) (*domain.Team, error) {
	var team domain.Team
	result := a.db.
		Joins("JOIN team_members ON team_members.team_id = teams.team_id").
		Where("teams.contest_id = ? AND team_members.user_id = ?", contestID, userID).
		First(&team)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrTeamNotFound
		}
		return nil, a.translateError(result.Error)
	}

	return &team, nil
}

// Game-based queries

func (a *TeamDatabaseAdapter) GetTeamByGameID(gameID int64) (*port.TeamWithMembers, error) {
	// Find team via game_teams table
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ?", gameID).First(&gameTeam)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrTeamNotFound
		}
		return nil, a.translateError(result.Error)
	}

	// Get team
	team, err := a.GetByID(gameTeam.TeamID)
	if err != nil {
		return nil, err
	}

	// Get members
	members, err := a.GetMembersByTeamID(gameTeam.TeamID)
	if err != nil {
		return nil, err
	}

	return &port.TeamWithMembers{
		Team:    team,
		Members: members,
	}, nil
}

func (a *TeamDatabaseAdapter) GetNextTeamID(contestID int64) (int64, error) {
	var maxTeamID int64
	result := a.db.Model(&domain.Team{}).
		Where("contest_id = ?", contestID).
		Select("COALESCE(MAX(team_id), 0)").
		Scan(&maxTeamID)

	if result.Error != nil {
		return 0, a.translateError(result.Error)
	}

	return maxTeamID + 1, nil
}

func (a *TeamDatabaseAdapter) GetMembersByGameID(gameID int64) ([]*domain.TeamMember, error) {
	// Find team via game_teams table first
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ?", gameID).First(&gameTeam)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrTeamNotFound
		}
		return nil, a.translateError(result.Error)
	}

	return a.GetMembersByTeamID(gameTeam.TeamID)
}

func (a *TeamDatabaseAdapter) GetByGameAndUser(gameID, userID int64) (*domain.TeamMember, error) {
	// Find team via game_teams table first
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ?", gameID).First(&gameTeam)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrTeamNotFound
		}
		return nil, a.translateError(result.Error)
	}

	return a.GetMemberByTeamAndUser(gameTeam.TeamID, userID)
}

func (a *TeamDatabaseAdapter) GetByGameAndTeamID(gameID, teamID int64) ([]*domain.TeamMember, error) {
	// Verify that the team is associated with this game
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ? AND team_id = ?", gameID, teamID).First(&gameTeam)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, exception.ErrTeamNotFound
		}
		return nil, a.translateError(result.Error)
	}

	return a.GetMembersByTeamID(teamID)
}

func (a *TeamDatabaseAdapter) DeleteAllByGameID(gameID int64) error {
	// Find team via game_teams table first
	var gameTeam domain.GameTeam
	result := a.db.Where("game_id = ?", gameID).First(&gameTeam)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil // No team to delete
		}
		return a.translateError(result.Error)
	}

	// Delete all members
	if err := a.DeleteAllMembersByTeamID(gameTeam.TeamID); err != nil {
		return err
	}

	// Delete the game_team relationship
	if err := a.db.Delete(&gameTeam).Error; err != nil {
		return a.translateError(err)
	}

	// Delete the team
	return a.Delete(gameTeam.TeamID)
}

// Error translation

func (a *TeamDatabaseAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrTeamNotFound
	}

	if a.isDuplicateKeyError(err) {
		return exception.ErrTeamNameAlreadyExists
	}

	if a.isForeignKeyError(err) {
		return exception.ErrContestNotFound
	}

	if a.isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (a *TeamDatabaseAdapter) translateMemberError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return exception.ErrTeamMemberNotFound
	}

	if a.isDuplicateKeyError(err) {
		return exception.ErrTeamMemberAlreadyExists
	}

	if a.isForeignKeyError(err) {
		return exception.ErrTeamNotFound
	}

	if a.isConnectionError(err) {
		return exception.ErrDBConnection
	}

	return err
}

func (a *TeamDatabaseAdapter) isDuplicateKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "1062") ||
		strings.Contains(errMsg, "duplicate key value") ||
		strings.Contains(errMsg, "23505")
}

func (a *TeamDatabaseAdapter) isForeignKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "foreign key constraint") ||
		strings.Contains(errMsg, "1452") ||
		strings.Contains(errMsg, "23503")
}

func (a *TeamDatabaseAdapter) isConnectionError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "refused")
}
