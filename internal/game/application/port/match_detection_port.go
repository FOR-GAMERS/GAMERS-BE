package port

import "time"

// ValorantMatch represents a match from the Valorant API (summary level)
type ValorantMatch struct {
	MatchID   string    `json:"match_id"`
	MapName   string    `json:"map_name"`
	GameMode  string    `json:"game_mode"`
	GameStart time.Time `json:"game_start"`
	GameLength int      `json:"game_length"` // seconds
}

// ValorantMatchDetail contains full match data including players and teams
type ValorantMatchDetail struct {
	MatchID      string              `json:"match_id"`
	MapName      string              `json:"map_name"`
	GameMode     string              `json:"game_mode"`
	GameStart    time.Time           `json:"game_start"`
	GameLength   int                 `json:"game_length"`
	RoundsPlayed int                 `json:"rounds_played"`
	Teams        []ValorantTeamData  `json:"teams"`
	Players      []ValorantPlayerData `json:"players"`
}

// ValorantTeamData represents one side's data in a Valorant match
type ValorantTeamData struct {
	TeamID    string `json:"team_id"` // "Red" or "Blue"
	HasWon    bool   `json:"has_won"`
	RoundsWon int    `json:"rounds_won"`
}

// ValorantPlayerData represents a single player's data in a Valorant match
type ValorantPlayerData struct {
	PUUID     string `json:"puuid"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	TeamID    string `json:"team_id"` // "Red" or "Blue"
	Agent     string `json:"agent"`
	Kills     int    `json:"kills"`
	Deaths    int    `json:"deaths"`
	Assists   int    `json:"assists"`
	Score     int    `json:"score"`
	Headshots int    `json:"headshots"`
	Bodyshots int    `json:"bodyshots"`
	Legshots  int    `json:"legshots"`
}

// MatchDetectionPort defines the interface for fetching match data from Valorant API
type MatchDetectionPort interface {
	// GetRecentMatches fetches recent match history for a player
	GetRecentMatches(region, name, tag string) ([]ValorantMatch, error)

	// GetMatchDetail fetches full details for a specific match
	GetMatchDetail(matchID string) (*ValorantMatchDetail, error)
}
