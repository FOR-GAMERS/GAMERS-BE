package domain

import "time"

// MatchResult stores the detailed result of a detected Valorant match for a tournament game
type MatchResult struct {
	MatchResultID   int64     `gorm:"column:match_result_id;primaryKey;autoIncrement" json:"match_result_id"`
	GameID          int64     `gorm:"column:game_id;type:bigint unsigned;not null;uniqueIndex:idx_match_results_game" json:"game_id"`
	ValorantMatchID string    `gorm:"column:valorant_match_id;type:varchar(255);not null;index:idx_match_results_valorant" json:"valorant_match_id"`
	MapName         string    `gorm:"column:map_name;type:varchar(50)" json:"map_name"`
	RoundsPlayed    int       `gorm:"column:rounds_played;type:int;not null" json:"rounds_played"`
	WinnerTeamID    int64     `gorm:"column:winner_team_id;type:bigint unsigned;not null" json:"winner_team_id"`
	LoserTeamID     int64     `gorm:"column:loser_team_id;type:bigint unsigned;not null" json:"loser_team_id"`
	WinnerScore     int       `gorm:"column:winner_score;type:int;not null" json:"winner_score"`
	LoserScore      int       `gorm:"column:loser_score;type:int;not null" json:"loser_score"`
	GameStartedAt   time.Time `gorm:"column:game_started_at;type:datetime;not null" json:"game_started_at"`
	GameDuration    int       `gorm:"column:game_duration;type:int;not null" json:"game_duration"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func NewMatchResult(
	gameID int64,
	valorantMatchID string,
	mapName string,
	roundsPlayed int,
	winnerTeamID, loserTeamID int64,
	winnerScore, loserScore int,
	gameStartedAt time.Time,
	gameDuration int,
) *MatchResult {
	return &MatchResult{
		GameID:          gameID,
		ValorantMatchID: valorantMatchID,
		MapName:         mapName,
		RoundsPlayed:    roundsPlayed,
		WinnerTeamID:    winnerTeamID,
		LoserTeamID:     loserTeamID,
		WinnerScore:     winnerScore,
		LoserScore:      loserScore,
		GameStartedAt:   gameStartedAt,
		GameDuration:    gameDuration,
		CreatedAt:       time.Now(),
	}
}

func (m *MatchResult) TableName() string {
	return "match_results"
}
