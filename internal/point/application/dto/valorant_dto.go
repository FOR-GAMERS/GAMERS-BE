package dto

import "GAMERS-BE/internal/point/domain"

type CreateValorantScoreTableDto struct {
	Radiant    int `json:"radiant" binding:"required"`
	Immortal3  int `json:"immortal_3" binding:"required"`
	Immortal2  int `json:"immortal_2" binding:"required"`
	Immortal1  int `json:"immortal_1" binding:"required"`
	Ascendant3 int `json:"ascendant_3" binding:"required"`
	Ascendant2 int `json:"ascendant_2" binding:"required"`
	Ascendant1 int `json:"ascendant_1" binding:"required"`
	Diamond3   int `json:"diamond_3" binding:"required"`
	Diamond2   int `json:"diamond_2" binding:"required"`
	Diamond1   int `json:"diamond_1" binding:"required"`
	Platinum3  int `json:"platinum_3" binding:"required"`
	Platinum2  int `json:"platinum_2" binding:"required"`
	Platinum1  int `json:"platinum_1" binding:"required"`
	Gold3      int `json:"gold_3" binding:"required"`
	Gold2      int `json:"gold_2" binding:"required"`
	Gold1      int `json:"gold_1" binding:"required"`
	Silver3    int `json:"silver_3" binding:"required"`
	Silver2    int `json:"silver_2" binding:"required"`
	Silver1    int `json:"silver_1" binding:"required"`
	Bronze3    int `json:"bronze_3" binding:"required"`
	Bronze2    int `json:"bronze_2" binding:"required"`
	Bronze1    int `json:"bronze_1" binding:"required"`
	Iron3      int `json:"iron_3" binding:"required"`
	Iron2      int `json:"iron_2" binding:"required"`
	Iron1      int `json:"iron_1" binding:"required"`
}

type ValorantScoreTableResponse struct {
	ScoreTableID int64 `json:"score_table_id"`
	Radiant      int   `json:"radiant"`
	Immortal3    int   `json:"immortal_3"`
	Immortal2    int   `json:"immortal_2"`
	Immortal1    int   `json:"immortal_1"`
	Ascendant3   int   `json:"ascendant_3"`
	Ascendant2   int   `json:"ascendant_2"`
	Ascendant1   int   `json:"ascendant_1"`
	Diamond3     int   `json:"diamond_3"`
	Diamond2     int   `json:"diamond_2"`
	Diamond1     int   `json:"diamond_1"`
	Platinum3    int   `json:"platinum_3"`
	Platinum2    int   `json:"platinum_2"`
	Platinum1    int   `json:"platinum_1"`
	Gold3        int   `json:"gold_3"`
	Gold2        int   `json:"gold_2"`
	Gold1        int   `json:"gold_1"`
	Silver3      int   `json:"silver_3"`
	Silver2      int   `json:"silver_2"`
	Silver1      int   `json:"silver_1"`
	Bronze3      int   `json:"bronze_3"`
	Bronze2      int   `json:"bronze_2"`
	Bronze1      int   `json:"bronze_1"`
	Iron3        int   `json:"iron_3"`
	Iron2        int   `json:"iron_2"`
	Iron1        int   `json:"iron_1"`
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
		Immortal3:    table.Immortal3,
		Immortal2:    table.Immortal2,
		Immortal1:    table.Immortal1,
		Ascendant3:   table.Ascendant3,
		Ascendant2:   table.Ascendant2,
		Ascendant1:   table.Ascendant1,
		Diamond3:     table.Diamond3,
		Diamond2:     table.Diamond2,
		Diamond1:     table.Diamond1,
		Platinum3:    table.Platinum3,
		Platinum2:    table.Platinum2,
		Platinum1:    table.Platinum1,
		Gold3:        table.Gold3,
		Gold2:        table.Gold2,
		Gold1:        table.Gold1,
		Silver3:      table.Silver3,
		Silver2:      table.Silver2,
		Silver1:      table.Silver1,
		Bronze3:      table.Bronze3,
		Bronze2:      table.Bronze2,
		Bronze1:      table.Bronze1,
		Iron3:        table.Iron3,
		Iron2:        table.Iron2,
		Iron1:        table.Iron1,
	}
}

func ToValorantScoreTableResponseList(tables []*domain.ValorantScoreTable) []*ValorantScoreTableResponse {
	responses := make([]*ValorantScoreTableResponse, len(tables))
	for i, table := range tables {
		responses[i] = ToValorantScoreTableResponse(table)
	}
	return responses
}
