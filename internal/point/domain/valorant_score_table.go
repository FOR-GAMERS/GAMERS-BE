package domain

import "time"

type ValorantScoreTable struct {
	ScoreTableID int64     `gorm:"column:score_table_id;primaryKey;autoIncrement" json:"score_table_id"`
	Radiant      int       `gorm:"column:radiant;type:int;not null" json:"radiant"`
	Immortal     int       `gorm:"column:immortal;type:int;not null" json:"immortal"`
	Ascendant    int       `gorm:"column:ascendant;type:int;not null" json:"ascendant"`
	Diamond      int       `gorm:"column:diamond;type:int;not null" json:"diamond"`
	Platinum     int       `gorm:"column:platinum;type:int;not null" json:"platinum"`
	Gold         int       `gorm:"column:gold;type:int;not null" json:"gold"`
	Silver       int       `gorm:"column:silver;type:int;not null" json:"silver"`
	Bronze       int       `gorm:"column:bronze;type:int;not null" json:"bronze"`
	Iron         int       `gorm:"column:iron;type:int;not null" json:"iron"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ModifiedAt   time.Time `gorm:"column:modified_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"modified_at"`
}

func (v *ValorantScoreTable) TableName() string {
	return "valorant_score_tables"
}

func NewValorantScoreTable(
	radiant, immortal, ascendant, diamond, platinum, gold, silver, bronze, iron int,
) *ValorantScoreTable {
	return &ValorantScoreTable{
		Radiant:   radiant,
		Immortal:  immortal,
		Ascendant: ascendant,
		Diamond:   diamond,
		Platinum:  platinum,
		Gold:      gold,
		Silver:    silver,
		Bronze:    bronze,
		Iron:      iron,
	}
}
