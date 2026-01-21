package dto

import (
	gameDomain "GAMERS-BE/internal/game/domain"
	"errors"
	"time"
)

type CreateGameRequest struct {
	ContestID    int64               `json:"contest_id" binding:"required"`
	GameTeamType gameDomain.GameTeamType `json:"game_team_type" binding:"required"`
	StartedAt    time.Time           `json:"started_at" binding:"required"`
	EndedAt      time.Time           `json:"ended_at" binding:"required"`
}

type UpdateGameRequest struct {
	GameStatus   *gameDomain.GameStatus   `json:"game_status,omitempty"`
	GameTeamType *gameDomain.GameTeamType `json:"game_team_type,omitempty"`
	StartedAt    *time.Time           `json:"started_at,omitempty"`
	EndedAt      *time.Time           `json:"ended_at,omitempty"`
}

type GameResponse struct {
	GameID       int64               `json:"game_id"`
	ContestID    int64               `json:"contest_id"`
	GameStatus   gameDomain.GameStatus   `json:"game_status"`
	GameTeamType gameDomain.GameTeamType `json:"game_team_type"`
	StartedAt    time.Time           `json:"started_at"`
	EndedAt      time.Time           `json:"ended_at"`
	CreatedAt    time.Time           `json:"created_at"`
	ModifiedAt   time.Time           `json:"modified_at"`
}

func (req *UpdateGameRequest) ApplyTo(game *gameDomain.Game) {
	if req.GameStatus != nil {
		game.GameStatus = *req.GameStatus
	}
	if req.GameTeamType != nil {
		game.GameTeamType = *req.GameTeamType
	}
	if req.StartedAt != nil {
		game.StartedAt = *req.StartedAt
	}
	if req.EndedAt != nil {
		game.EndedAt = *req.EndedAt
	}
}

func (req *UpdateGameRequest) HasChanges() bool {
	return req.GameStatus != nil ||
		req.GameTeamType != nil ||
		req.StartedAt != nil ||
		req.EndedAt != nil
}

func (req *UpdateGameRequest) Validate() error {
	if req.StartedAt != nil && req.EndedAt != nil {
		if req.EndedAt.Before(*req.StartedAt) {
			return errors.New("end time must be after start time")
		}

		// EndedAt must be within 2 hours of StartedAt
		maxEndTime := req.StartedAt.Add(2 * time.Hour)
		if req.EndedAt.After(maxEndTime) {
			return errors.New("game duration cannot exceed 2 hours")
		}
	}

	if req.GameTeamType != nil && !req.GameTeamType.IsValid() {
		return errors.New("invalid game team type")
	}

	return nil
}

func ToGameResponse(game *gameDomain.Game) *GameResponse {
	return &GameResponse{
		GameID:       game.GameID,
		ContestID:    game.ContestID,
		GameStatus:   game.GameStatus,
		GameTeamType: game.GameTeamType,
		StartedAt:    game.StartedAt,
		EndedAt:      game.EndedAt,
		CreatedAt:    game.CreatedAt,
		ModifiedAt:   game.ModifiedAt,
	}
}
