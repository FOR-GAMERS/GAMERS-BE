package port

import "github.com/FOR-GAMERS/GAMERS-BE/internal/game/domain"

// MatchResultDatabasePort defines the interface for match result persistence
type MatchResultDatabasePort interface {
	Save(result *domain.MatchResult) (*domain.MatchResult, error)
	GetByGameID(gameID int64) (*domain.MatchResult, error)
	SavePlayerStats(stats []*domain.MatchPlayerStat) error
	GetPlayerStatsByMatchResult(matchResultID int64) ([]*domain.MatchPlayerStat, error)
}
