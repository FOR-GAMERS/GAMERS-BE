package port

import "GAMERS-BE/internal/game/domain"

// TeamDatabasePort defines the interface for team config operations
type TeamDatabasePort interface {
	// Team operations
	Save(team *domain.Team) (*domain.Team, error)
	GetByID(teamID int64) (*domain.Team, error)
	GetByContestID(contestID int64) ([]*domain.Team, error)
	GetByContestAndName(contestID int64, teamName string) (*domain.Team, error)
	CountByContestID(contestID int64) (int, error)
	Update(team *domain.Team) error
	Delete(teamID int64) error
	DeleteByContestID(contestID int64) error

	// TeamMember operations
	SaveMember(member *domain.TeamMember) (*domain.TeamMember, error)
	SaveMemberBatch(members []*domain.TeamMember) error
	GetMemberByID(id int64) (*domain.TeamMember, error)
	GetMembersByTeamID(teamID int64) ([]*domain.TeamMember, error)
	GetMemberByTeamAndUser(teamID, userID int64) (*domain.TeamMember, error)
	GetMemberCountByTeamID(teamID int64) (int, error)
	GetLeaderByTeamID(teamID int64) (*domain.TeamMember, error)
	UpdateMember(member *domain.TeamMember) error
	DeleteMember(id int64) error
	DeleteMemberByTeamAndUser(teamID, userID int64) error
	DeleteAllMembersByTeamID(teamID int64) error

	// Contest-based queries
	GetTeamsByContestWithMembers(contestID int64) ([]*TeamWithMembers, error)
	GetUserTeamInContest(contestID, userID int64) (*domain.Team, error)

	// Game-based queries (via game_teams table)
	GetTeamByGameID(gameID int64) (*TeamWithMembers, error)
	GetNextTeamID(contestID int64) (int64, error)
	GetMembersByGameID(gameID int64) ([]*domain.TeamMember, error)
	GetByGameAndUser(gameID, userID int64) (*domain.TeamMember, error)
	GetByGameAndTeamID(gameID, teamID int64) ([]*domain.TeamMember, error)
	DeleteAllByGameID(gameID int64) error
}

// TeamWithMembers represents a team with its members
type TeamWithMembers struct {
	Team    *domain.Team
	Members []*domain.TeamMember
}
