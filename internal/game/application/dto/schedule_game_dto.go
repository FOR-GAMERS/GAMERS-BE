package dto

import (
	gameDomain "github.com/FOR-GAMERS/GAMERS-BE/internal/game/domain"
	"time"
)

// ScheduleGameRequest is the request body for setting a game's scheduled start time
type ScheduleGameRequest struct {
	ScheduledStartTime     time.Time `json:"scheduledStartTime" binding:"required"`
	DetectionWindowMinutes int       `json:"detectionWindowMinutes"`
}

// ScheduleGameResponse is the response after scheduling a game
type ScheduleGameResponse struct {
	GameID                 int64                      `json:"gameId"`
	ContestID              int64                      `json:"contestId"`
	Round                  int                        `json:"round,omitempty"`
	MatchNumber            int                        `json:"matchNumber,omitempty"`
	ScheduledStartTime     *time.Time                 `json:"scheduledStartTime"`
	DetectionWindowMinutes int                        `json:"detectionWindowMinutes"`
	GameStatus             gameDomain.GameStatus      `json:"gameStatus"`
	DetectionStatus        gameDomain.DetectionStatus  `json:"detectionStatus"`
}

func ToScheduleGameResponse(game *gameDomain.Game) *ScheduleGameResponse {
	resp := &ScheduleGameResponse{
		GameID:                 game.GameID,
		ContestID:              game.ContestID,
		ScheduledStartTime:     game.ScheduledStartTime,
		DetectionWindowMinutes: game.DetectionWindowMinutes,
		GameStatus:             game.GameStatus,
		DetectionStatus:        game.DetectionStatus,
	}
	if game.Round != nil {
		resp.Round = *game.Round
	}
	if game.MatchNumber != nil {
		resp.MatchNumber = *game.MatchNumber
	}
	return resp
}

// ManualResultRequest is the request body for staff manual result input
type ManualResultRequest struct {
	WinnerTeamID int64  `json:"winnerTeamId" binding:"required"`
	WinnerScore  int    `json:"winnerScore" binding:"required"`
	LoserScore   int    `json:"loserScore" binding:"required"`
	Note         string `json:"note,omitempty"`
}

// MatchResultResponse is the response for match result queries
type MatchResultResponse struct {
	MatchResultID   int64                      `json:"matchResultId"`
	GameID          int64                      `json:"gameId"`
	ValorantMatchID string                     `json:"valorantMatchId,omitempty"`
	MapName         string                     `json:"mapName,omitempty"`
	RoundsPlayed    int                        `json:"roundsPlayed"`
	WinnerTeamID    int64                      `json:"winnerTeamId"`
	LoserTeamID     int64                      `json:"loserTeamId"`
	WinnerScore     int                        `json:"winnerScore"`
	LoserScore      int                        `json:"loserScore"`
	GameStartedAt   *time.Time                 `json:"gameStartedAt,omitempty"`
	GameDuration    int                        `json:"gameDuration,omitempty"`
	DetectionStatus gameDomain.DetectionStatus `json:"detectionStatus"`
	PlayerStats     []*PlayerStatResponse      `json:"playerStats,omitempty"`
}

// PlayerStatResponse represents individual player stats
type PlayerStatResponse struct {
	UserID    int64  `json:"userId"`
	TeamID    int64  `json:"teamId"`
	AgentName string `json:"agentName,omitempty"`
	Kills     int    `json:"kills"`
	Deaths    int    `json:"deaths"`
	Assists   int    `json:"assists"`
	Score     int    `json:"score"`
	Headshots int    `json:"headshots"`
	Bodyshots int    `json:"bodyshots"`
	Legshots  int    `json:"legshots"`
}

func ToMatchResultResponse(result *gameDomain.MatchResult, detectionStatus gameDomain.DetectionStatus) *MatchResultResponse {
	return &MatchResultResponse{
		MatchResultID:   result.MatchResultID,
		GameID:          result.GameID,
		ValorantMatchID: result.ValorantMatchID,
		MapName:         result.MapName,
		RoundsPlayed:    result.RoundsPlayed,
		WinnerTeamID:    result.WinnerTeamID,
		LoserTeamID:     result.LoserTeamID,
		WinnerScore:     result.WinnerScore,
		LoserScore:      result.LoserScore,
		GameStartedAt:   &result.GameStartedAt,
		GameDuration:    result.GameDuration,
		DetectionStatus: detectionStatus,
	}
}

func ToPlayerStatResponse(stat *gameDomain.MatchPlayerStat) *PlayerStatResponse {
	return &PlayerStatResponse{
		UserID:    stat.UserID,
		TeamID:    stat.TeamID,
		AgentName: stat.AgentName,
		Kills:     stat.Kills,
		Deaths:    stat.Deaths,
		Assists:   stat.Assists,
		Score:     stat.Score,
		Headshots: stat.Headshots,
		Bodyshots: stat.Bodyshots,
		Legshots:  stat.Legshots,
	}
}
