package dto

import gameDomain "GAMERS-BE/internal/game/domain"

type CreateGameTeamRequest struct {
	GameID int64 `json:"game_id" binding:"required"`
	TeamID int64 `json:"team_id" binding:"required"`
	Grade  *int  `json:"grade,omitempty"`
}

type GameTeamResponse struct {
	GameTeamID int64 `json:"game_team_id"`
	GameID     int64 `json:"game_id"`
	TeamID     int64 `json:"team_id"`
	Grade      *int  `json:"grade,omitempty"`
}

func ToGameTeamResponse(gt *gameDomain.GameTeam) *GameTeamResponse {
	return &GameTeamResponse{
		GameTeamID: gt.GameTeamID,
		GameID:     gt.GameID,
		TeamID:     gt.TeamID,
		Grade:      gt.Grade,
	}
}

func ToGameTeamResponseList(gameTeams []*gameDomain.GameTeam) []*GameTeamResponse {
	responses := make([]*GameTeamResponse, len(gameTeams))
	for i, gt := range gameTeams {
		responses[i] = ToGameTeamResponse(gt)
	}
	return responses
}
