package port

import "GAMERS-BE/internal/game/domain"

type GameDatabasePort interface {
	Save(game *domain.Game) (*domain.Game, error)
	SaveBatch(games []*domain.Game) error
	GetByID(gameID int64) (*domain.Game, error)
	GetByContestID(contestID int64) ([]*domain.Game, error)
	GetByContestAndRound(contestID int64, round int) ([]*domain.Game, error)
	Update(game *domain.Game) error
	Delete(gameID int64) error
	DeleteByContestID(contestID int64) error

	// Scheduler queries for match detection
	GetGamesReadyToStart() ([]*domain.Game, error)
	GetGamesInDetection() ([]*domain.Game, error)
}
