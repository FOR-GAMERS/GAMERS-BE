package domain

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
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

// DetectionStatus represents the match detection state for a tournament game
type DetectionStatus string

const (
	DetectionStatusNone      DetectionStatus = "NONE"
	DetectionStatusDetecting DetectionStatus = "DETECTING"
	DetectionStatusDetected  DetectionStatus = "DETECTED"
	DetectionStatusFailed    DetectionStatus = "FAILED"
	DetectionStatusManual    DetectionStatus = "MANUAL"
)

func (d DetectionStatus) IsValid() bool {
	switch d {
	case DetectionStatusNone, DetectionStatusDetecting, DetectionStatusDetected,
		DetectionStatusFailed, DetectionStatusManual:
		return true
	default:
		return false
	}
}

// detectionAllowedTransitions defines valid DetectionStatus state transitions
var detectionAllowedTransitions = map[DetectionStatus][]DetectionStatus{
	DetectionStatusNone:      {DetectionStatusDetecting},
	DetectionStatusDetecting: {DetectionStatusDetected, DetectionStatusFailed, DetectionStatusManual},
	DetectionStatusFailed:    {DetectionStatusManual, DetectionStatusDetecting},
	DetectionStatusDetected:  {},
	DetectionStatusManual:    {},
}

type Game struct {
	GameID                 int64           `gorm:"column:game_id;primaryKey;autoIncrement" json:"game_id"`
	ContestID              int64           `gorm:"column:contest_id;type:bigint;not null" json:"contest_id"`
	GameStatus             GameStatus      `gorm:"column:game_status;type:varchar(16);not null" json:"game_status"`
	GameTeamType           GameTeamType    `gorm:"column:game_team_type;type:varchar(16);not null" json:"game_team_type"`
	StartedAt              *time.Time      `gorm:"column:started_at;type:datetime" json:"started_at,omitempty"`
	EndedAt                *time.Time      `gorm:"column:ended_at;type:datetime" json:"ended_at,omitempty"`
	Round                  *int            `gorm:"column:round;type:int" json:"round,omitempty"`
	MatchNumber            *int            `gorm:"column:match_number;type:int" json:"match_number,omitempty"`
	NextGameID             *int64          `gorm:"column:next_game_id;type:bigint" json:"next_game_id,omitempty"`
	BracketPosition        *int            `gorm:"column:bracket_position;type:int" json:"bracket_position,omitempty"`
	ScheduledStartTime     *time.Time      `gorm:"column:scheduled_start_time;type:datetime" json:"scheduled_start_time,omitempty"`
	DetectionWindowMinutes int             `gorm:"column:detection_window_minutes;type:int;not null;default:120" json:"detection_window_minutes"`
	DetectedMatchID        *string         `gorm:"column:detected_match_id;type:varchar(255)" json:"detected_match_id,omitempty"`
	DetectionStatus        DetectionStatus `gorm:"column:detection_status;type:varchar(20);not null;default:'NONE'" json:"detection_status"`
	CreatedAt              time.Time       `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ModifiedAt             time.Time       `gorm:"column:modified_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"modified_at"`
}

func NewGame(
	contestID int64,
	gameTeamType GameTeamType,
	startedAt, endedAt *time.Time,
) *Game {
	now := time.Now()
	return &Game{
		ContestID:    contestID,
		GameStatus:   GameStatusPending,
		GameTeamType: gameTeamType,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		CreatedAt:    now,
		ModifiedAt:   now,
	}
}

// NewTournamentGame creates a new tournament game with bracket information
func NewTournamentGame(
	contestID int64,
	gameTeamType GameTeamType,
	round, matchNumber, bracketPosition int,
) *Game {
	now := time.Now()
	return &Game{
		ContestID:       contestID,
		GameStatus:      GameStatusPending,
		GameTeamType:    gameTeamType,
		Round:           &round,
		MatchNumber:     &matchNumber,
		BracketPosition: &bracketPosition,
		CreatedAt:       now,
		ModifiedAt:      now,
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
	if g.IsTournamentGame() && g.StartedAt == nil && g.EndedAt == nil {
		return nil
	}

	// If neither date is set, it's valid (dates can be set later)
	if g.StartedAt == nil && g.EndedAt == nil {
		return nil
	}

	if g.StartedAt == nil {
		return exception.ErrGameStartTimeRequired
	}

	if g.EndedAt == nil {
		return exception.ErrGameEndTimeRequired
	}

	if !g.StartedAt.Before(*g.EndedAt) {
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

// CanTransitionDetectionTo checks if the detection status can transition to the target
func (g *Game) CanTransitionDetectionTo(target DetectionStatus) bool {
	allowed, exists := detectionAllowedTransitions[g.DetectionStatus]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// TransitionDetectionTo transitions the detection status to the target
func (g *Game) TransitionDetectionTo(target DetectionStatus) error {
	if !g.CanTransitionDetectionTo(target) {
		return exception.ErrInvalidDetectionStatusTransition
	}
	g.DetectionStatus = target
	g.ModifiedAt = time.Now()
	return nil
}

// SetSchedule sets the scheduled start time and detection window
func (g *Game) SetSchedule(scheduledStartTime time.Time, detectionWindowMinutes int) error {
	if !g.IsPending() {
		return exception.ErrGameNotPending
	}
	if scheduledStartTime.Before(time.Now()) {
		return exception.ErrScheduledTimeInPast
	}
	if detectionWindowMinutes <= 0 {
		detectionWindowMinutes = 120
	}
	g.ScheduledStartTime = &scheduledStartTime
	g.DetectionWindowMinutes = detectionWindowMinutes
	g.ModifiedAt = time.Now()
	return nil
}

// IsReadyToActivate checks if scheduled start time has arrived and game is pending
func (g *Game) IsReadyToActivate() bool {
	return g.GameStatus == GameStatusPending &&
		g.ScheduledStartTime != nil &&
		!g.ScheduledStartTime.After(time.Now())
}

// IsDetectionWindowExpired checks if the detection window has passed
func (g *Game) IsDetectionWindowExpired() bool {
	if g.ScheduledStartTime == nil {
		return false
	}
	windowEnd := g.ScheduledStartTime.Add(time.Duration(g.DetectionWindowMinutes) * time.Minute)
	return time.Now().After(windowEnd)
}

// GetDetectionWindowEnd returns the end time of the detection window
func (g *Game) GetDetectionWindowEnd() time.Time {
	if g.ScheduledStartTime == nil {
		return time.Time{}
	}
	return g.ScheduledStartTime.Add(time.Duration(g.DetectionWindowMinutes) * time.Minute)
}

// ActivateForDetection transitions game to ACTIVE and starts detection
func (g *Game) ActivateForDetection() error {
	if err := g.TransitionTo(GameStatusActive); err != nil {
		return err
	}
	now := time.Now()
	g.StartedAt = &now
	return g.TransitionDetectionTo(DetectionStatusDetecting)
}

// MarkDetected records the detected match ID and transitions statuses
func (g *Game) MarkDetected(matchID string) error {
	if err := g.TransitionDetectionTo(DetectionStatusDetected); err != nil {
		return err
	}
	g.DetectedMatchID = &matchID
	return nil
}

// MarkDetectionFailed transitions detection to FAILED status
func (g *Game) MarkDetectionFailed() error {
	return g.TransitionDetectionTo(DetectionStatusFailed)
}

// MarkManualResult transitions detection to MANUAL status
func (g *Game) MarkManualResult() error {
	return g.TransitionDetectionTo(DetectionStatusManual)
}

// FinishGame transitions game to FINISHED and records end time
func (g *Game) FinishGame() error {
	if err := g.TransitionTo(GameStatusFinished); err != nil {
		return err
	}
	now := time.Now()
	g.EndedAt = &now
	g.ModifiedAt = now
	return nil
}

// IsDetecting returns true if the game is actively detecting matches
func (g *Game) IsDetecting() bool {
	return g.GameStatus == GameStatusActive && g.DetectionStatus == DetectionStatusDetecting
}
