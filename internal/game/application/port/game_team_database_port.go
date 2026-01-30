package port

import "github.com/FOR-GAMERS/GAMERS-BE/internal/game/domain"

type GameTeamDatabasePort interface {
	Save(gameTeam *domain.GameTeam) (*domain.GameTeam, error)
	GetByID(gameTeamID int64) (*domain.GameTeam, error)
	GetByGameID(gameID int64) ([]*domain.GameTeam, error)
	GetByGameAndTeam(gameID, teamID int64) (*domain.GameTeam, error)
	GetByGrade(gameID int64, grade int) (*domain.GameTeam, error)
	Delete(gameTeamID int64) error
	DeleteByGameID(gameID int64) error
}
