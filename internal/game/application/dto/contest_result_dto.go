package dto

// ContestResultResponse represents the full tournament bracket result
type ContestResultResponse struct {
	ContestID     int64         `json:"contest_id"`
	Title         string        `json:"title"`
	ContestStatus string        `json:"contest_status"`
	TotalRounds   int           `json:"total_rounds"`
	Champion      *TeamSummary  `json:"champion,omitempty"`
	Rounds        []RoundResult `json:"rounds"`
}

// RoundResult represents a single round in the tournament
type RoundResult struct {
	Round     int          `json:"round"`
	RoundName string       `json:"round_name"`
	Games     []GameResult `json:"games"`
}

// GameResult represents a single game in a round
type GameResult struct {
	GameID          int64               `json:"game_id"`
	MatchNumber     int                 `json:"match_number"`
	GameStatus      string              `json:"game_status"`
	DetectionStatus string              `json:"detection_status"`
	Teams           []GameTeamResult    `json:"teams"`
	MatchResult     *MatchResultSummary `json:"match_result,omitempty"`
}

// GameTeamResult represents a team in a game
type GameTeamResult struct {
	TeamID   int64  `json:"team_id"`
	TeamName string `json:"team_name"`
	Grade    *int   `json:"grade,omitempty"`
}

// MatchResultSummary represents the result of a finished game
type MatchResultSummary struct {
	WinnerTeamID int64  `json:"winner_team_id"`
	LoserTeamID  int64  `json:"loser_team_id"`
	WinnerScore  int    `json:"winner_score"`
	LoserScore   int    `json:"loser_score"`
	MapName      string `json:"map_name,omitempty"`
}

// TeamSummary represents a team summary
type TeamSummary struct {
	TeamID   int64  `json:"team_id"`
	TeamName string `json:"team_name"`
}
