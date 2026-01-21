package dto

import "GAMERS-BE/internal/point/domain"

type CreateValorantScoreTableDto struct {
	Radiant   int `json:"radiant" binding:"required"`
	Immortal  int `json:"immortal" binding:"required"`
	Ascendant int `json:"ascendant" binding:"required"`
	Diamond   int `json:"diamond" binding:"required"`
	Platinum  int `json:"platinum" binding:"required"`
	Gold      int `json:"gold" binding:"required"`
	Silver    int `json:"silver" binding:"required"`
	Bronze    int `json:"bronze" binding:"required"`
	Iron      int `json:"iron" binding:"required"`
}

type ValorantScoreTableResponse struct {
	ScoreTableID int64 `json:"score_table_id"`
	Radiant      int   `json:"radiant"`
	Immortal     int   `json:"immortal"`
	Ascendant    int   `json:"ascendant"`
	Diamond      int   `json:"diamond"`
	Platinum     int   `json:"platinum"`
	Gold         int   `json:"gold"`
	Silver       int   `json:"silver"`
	Bronze       int   `json:"bronze"`
	Iron         int   `json:"iron"`
}

type GetContestPointDto struct {
	Region   string `json:"region" binding:"required"`
	Username string `json:"username" binding:"required"`
	Tag      string `json:"tag" binding:"required"`
}

func ToValorantScoreTableResponse(table *domain.ValorantScoreTable) *ValorantScoreTableResponse {
	if table == nil {
		return nil
	}

	return &ValorantScoreTableResponse{
		ScoreTableID: table.ScoreTableID,
		Radiant:      table.Radiant,
		Immortal:     table.Immortal,
		Ascendant:    table.Ascendant,
		Diamond:      table.Diamond,
		Platinum:     table.Platinum,
		Gold:         table.Gold,
		Silver:       table.Silver,
		Bronze:       table.Bronze,
		Iron:         table.Iron,
	}
}

func ToValorantScoreTableResponseList(tables []*domain.ValorantScoreTable) []*ValorantScoreTableResponse {
	responses := make([]*ValorantScoreTableResponse, len(tables))
	for i, table := range tables {
		responses[i] = ToValorantScoreTableResponse(table)
	}
	return responses
}
