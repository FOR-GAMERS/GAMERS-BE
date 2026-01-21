package domain

import (
	"GAMERS-BE/internal/global/exception"
	"time"
)

type GameStatus string

const (
	GameStatusPending   GameStatus = "PENDING"
	GameStatusActive    GameStatus = "ACTIVE"
	GameStatusFinished  GameStatus = "FINISHED"
	GameStatusCancelled GameStatus = "CANCELLED"
)

type GameTeamType string

const (
	GameTeamTypeSingle GameTeamType = "SINGLE" // 1 member
	GameTeamTypeDuo    GameTeamType = "DUO"    // 2 members
	GameTeamTypeTrio   GameTeamType = "TRIO"   // 3 members
	GameTeamTypeFull   GameTeamType = "FULL"   // 4 members
	GameTeamTypeHurupa GameTeamType = "HURUPA" // 5 members
)

// GetMaxTeamMembers returns the maximum number of team members for this team type
func (t GameTeamType) GetMaxTeamMembers() int {
	switch t {
	case GameTeamTypeSingle:
		return 1
	case GameTeamTypeDuo:
		return 2
	case GameTeamTypeTrio:
		return 3
	case GameTeamTypeFull:
		return 4
	case GameTeamTypeHurupa:
		return 5
	default:
		return 0
	}
}

func (t GameTeamType) IsValid() bool {
	switch t {
	case GameTeamTypeSingle, GameTeamTypeDuo, GameTeamTypeTrio, GameTeamTypeFull, GameTeamTypeHurupa:
		return true
	default:
		return false
	}
}

type Game struct {
	GameID          int64        `gorm:"column:game_id;primaryKey;autoIncrement" json:"game_id"`
	ContestID       int64        `gorm:"column:contest_id;type:bigint;not null" json:"contest_id"`
	GameStatus      GameStatus   `gorm:"column:game_status;type:varchar(16);not null" json:"game_status"`
	GameTeamType    GameTeamType `gorm:"column:game_team_type;type:varchar(16);not null" json:"game_team_type"`
	StartedAt       time.Time    `gorm:"column:started_at;type:datetime;not null" json:"started_at"`
	EndedAt         time.Time    `gorm:"column:ended_at;type:datetime;not null" json:"ended_at"`
	Round           *int         `gorm:"column:round;type:int" json:"round,omitempty"`
	MatchNumber     *int         `gorm:"column:match_number;type:int" json:"match_number,omitempty"`
	NextGameID      *int64       `gorm:"column:next_game_id;type:bigint" json:"next_game_id,omitempty"`
	BracketPosition *int         `gorm:"column:bracket_position;type:int" json:"bracket_position,omitempty"`
	CreatedAt       time.Time    `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ModifiedAt      time.Time    `gorm:"column:modified_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"modified_at"`
}

func NewGame(
	contestID int64,
	gameTeamType GameTeamType,
	startedAt, endedAt time.Time,
) *Game {
	return &Game{
		ContestID:    contestID,
		GameStatus:   GameStatusPending,
		GameTeamType: gameTeamType,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
	}
}

// NewTournamentGame creates a new tournament game with bracket information
func NewTournamentGame(
	contestID int64,
	gameTeamType GameTeamType,
	round, matchNumber, bracketPosition int,
) *Game {
	return &Game{
		ContestID:       contestID,
		GameStatus:      GameStatusPending,
		GameTeamType:    gameTeamType,
		Round:           &round,
		MatchNumber:     &matchNumber,
		BracketPosition: &bracketPosition,
	}
}

// SetNextGame sets the next game (winner advances to)
func (g *Game) SetNextGame(nextGameID int64) {
	g.NextGameID = &nextGameID
}

// IsTournamentGame checks if this game is part of a tournament bracket
func (g *Game) IsTournamentGame() bool {
	return g.Round != nil && g.MatchNumber != nil
}

// GetRound returns the round number (0 if not set)
func (g *Game) GetRound() int {
	if g.Round == nil {
		return 0
	}
	return *g.Round
}

// GetMatchNumber returns the match number (0 if not set)
func (g *Game) GetMatchNumber() int {
	if g.MatchNumber == nil {
		return 0
	}
	return *g.MatchNumber
}

func (g *Game) TableName() string {
	return "games"
}

// State machine transitions
var gameAllowedTransitions = map[GameStatus][]GameStatus{
	GameStatusPending: {
		GameStatusActive,
		GameStatusCancelled,
	},
	GameStatusActive: {
		GameStatusFinished,
		GameStatusCancelled,
	},
	GameStatusFinished:  {},
	GameStatusCancelled: {},
}

func (g *Game) CanTransitionTo(targetStatus GameStatus) bool {
	allowedTargets, exists := gameAllowedTransitions[g.GameStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if allowed == targetStatus {
			return true
		}
	}
	return false
}

func (g *Game) TransitionTo(targetStatus GameStatus) error {
	if !g.CanTransitionTo(targetStatus) {
		return exception.ErrInvalidGameStatusTransition
	}

	g.GameStatus = targetStatus
	return nil
}

func (g *Game) IsTerminalState() bool {
	return g.GameStatus == GameStatusFinished || g.GameStatus == GameStatusCancelled
}

func (g *Game) IsPending() bool {
	return g.GameStatus == GameStatusPending
}

func (g *Game) CanInviteTeam() bool {
	return g.GameStatus == GameStatusPending
}

func (g *Game) IsValidStatus() bool {
	switch g.GameStatus {
	case GameStatusPending, GameStatusActive, GameStatusFinished, GameStatusCancelled:
		return true
	default:
		return false
	}
}

func (g *Game) IsValidTeamType() bool {
	return g.GameTeamType.IsValid()
}

// ValidateDates checks if the game dates are valid
// EndedAt must be after StartedAt and within 2 hours
// For tournament games, dates may be set later
func (g *Game) ValidateDates() error {
	// Tournament games don't require dates initially
	if g.IsTournamentGame() && g.StartedAt.IsZero() && g.EndedAt.IsZero() {
		return nil
	}

	if g.StartedAt.IsZero() {
		return exception.ErrGameStartTimeRequired
	}

	if g.EndedAt.IsZero() {
		return exception.ErrGameEndTimeRequired
	}

	if !g.StartedAt.Before(g.EndedAt) {
		return exception.ErrInvalidGameDates
	}

	// EndedAt must be within 2 hours of StartedAt
	maxEndTime := g.StartedAt.Add(2 * time.Hour)
	if g.EndedAt.After(maxEndTime) {
		return exception.ErrGameDurationExceeded
	}

	return nil
}

func (g *Game) Validate() error {
	if g.ContestID <= 0 {
		return exception.ErrInvalidContestID
	}

	if !g.IsValidStatus() {
		return exception.ErrInvalidGameStatus
	}

	// GameTeamType may not be set for tournament placeholder games
	if g.GameTeamType != "" && !g.IsValidTeamType() {
		return exception.ErrInvalidGameTeamType
	}

	if err := g.ValidateDates(); err != nil {
		return err
	}

	return nil
}

// ValidateForTournament validates tournament-specific fields
func (g *Game) ValidateForTournament() error {
	if err := g.Validate(); err != nil {
		return err
	}

	if g.Round == nil || *g.Round < 1 {
		return exception.ErrInvalidTournamentRound
	}

	if g.MatchNumber == nil || *g.MatchNumber < 1 {
		return exception.ErrInvalidMatchNumber
	}

	return nil
}
