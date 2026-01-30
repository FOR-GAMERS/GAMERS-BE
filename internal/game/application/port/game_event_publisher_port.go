package port

import (
	"context"
	"time"
)

// GameEventType defines the type of game events
type GameEventType string

const (
	GameEventScheduled          GameEventType = "game.scheduled"
	GameEventActivated          GameEventType = "game.activated"
	GameEventMatchDetecting     GameEventType = "game.match.detecting"
	GameEventMatchDetected      GameEventType = "game.match.detected"
	GameEventMatchFailed        GameEventType = "game.match.failed"
	GameEventFinished           GameEventType = "game.finished"
	GameEventManualResult       GameEventType = "game.result.manual"
)

// GameEvent is the base event structure for game-related events
type GameEvent struct {
	EventID   string        `json:"event_id"`
	EventType GameEventType `json:"event_type"`
	Timestamp time.Time     `json:"timestamp"`
	ContestID int64         `json:"contest_id"`
	GameID    int64         `json:"game_id"`
	Round     int           `json:"round,omitempty"`
	MatchNumber int         `json:"match_number,omitempty"`
}

// MatchDetectedEvent is published when a Valorant match is detected for a game
type MatchDetectedEvent struct {
	GameEvent
	ValorantMatchID string `json:"valorant_match_id"`
	WinnerTeamID    int64  `json:"winner_team_id"`
	WinnerTeamName  string `json:"winner_team_name"`
	LoserTeamID     int64  `json:"loser_team_id"`
	LoserTeamName   string `json:"loser_team_name"`
	Score           string `json:"score"`
	MapName         string `json:"map_name"`
}

// GameEventPublisherPort defines the interface for publishing game events to RabbitMQ
type GameEventPublisherPort interface {
	PublishGameEvent(ctx context.Context, event *GameEvent) error
	PublishMatchDetectedEvent(ctx context.Context, event *MatchDetectedEvent) error
}
