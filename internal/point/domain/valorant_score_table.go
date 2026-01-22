package domain

import "time"

type ValorantScoreTable struct {
	ScoreTableID int64     `gorm:"column:score_table_id;primaryKey;autoIncrement" json:"score_table_id"`
	Radiant      int       `gorm:"column:radiant;type:int;not null" json:"radiant"`
	Immortal3    int       `gorm:"column:immortal_3;type:int;not null" json:"immortal_3"`
	Immortal2    int       `gorm:"column:immortal_2;type:int;not null" json:"immortal_2"`
	Immortal1    int       `gorm:"column:immortal_1;type:int;not null" json:"immortal_1"`
	Ascendant3   int       `gorm:"column:ascendant_3;type:int;not null" json:"ascendant_3"`
	Ascendant2   int       `gorm:"column:ascendant_2;type:int;not null" json:"ascendant_2"`
	Ascendant1   int       `gorm:"column:ascendant_1;type:int;not null" json:"ascendant_1"`
	Diamond3     int       `gorm:"column:diamond_3;type:int;not null" json:"diamond_3"`
	Diamond2     int       `gorm:"column:diamond_2;type:int;not null" json:"diamond_2"`
	Diamond1     int       `gorm:"column:diamond_1;type:int;not null" json:"diamond_1"`
	Platinum3    int       `gorm:"column:platinum_3;type:int;not null" json:"platinum_3"`
	Platinum2    int       `gorm:"column:platinum_2;type:int;not null" json:"platinum_2"`
	Platinum1    int       `gorm:"column:platinum_1;type:int;not null" json:"platinum_1"`
	Gold3        int       `gorm:"column:gold_3;type:int;not null" json:"gold_3"`
	Gold2        int       `gorm:"column:gold_2;type:int;not null" json:"gold_2"`
	Gold1        int       `gorm:"column:gold_1;type:int;not null" json:"gold_1"`
	Silver3      int       `gorm:"column:silver_3;type:int;not null" json:"silver_3"`
	Silver2      int       `gorm:"column:silver_2;type:int;not null" json:"silver_2"`
	Silver1      int       `gorm:"column:silver_1;type:int;not null" json:"silver_1"`
	Bronze3      int       `gorm:"column:bronze_3;type:int;not null" json:"bronze_3"`
	Bronze2      int       `gorm:"column:bronze_2;type:int;not null" json:"bronze_2"`
	Bronze1      int       `gorm:"column:bronze_1;type:int;not null" json:"bronze_1"`
	Iron3        int       `gorm:"column:iron_3;type:int;not null" json:"iron_3"`
	Iron2        int       `gorm:"column:iron_2;type:int;not null" json:"iron_2"`
	Iron1        int       `gorm:"column:iron_1;type:int;not null" json:"iron_1"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ModifiedAt   time.Time `gorm:"column:modified_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"modified_at"`
}

func (v *ValorantScoreTable) TableName() string {
	return "valorant_score_tables"
}

func NewValorantScoreTable(
	radiant int,
	immortal3, immortal2, immortal1 int,
	ascendant3, ascendant2, ascendant1 int,
	diamond3, diamond2, diamond1 int,
	platinum3, platinum2, platinum1 int,
	gold3, gold2, gold1 int,
	silver3, silver2, silver1 int,
	bronze3, bronze2, bronze1 int,
	iron3, iron2, iron1 int,
) *ValorantScoreTable {
	return &ValorantScoreTable{
		Radiant:    radiant,
		Immortal3:  immortal3,
		Immortal2:  immortal2,
		Immortal1:  immortal1,
		Ascendant3: ascendant3,
		Ascendant2: ascendant2,
		Ascendant1: ascendant1,
		Diamond3:   diamond3,
		Diamond2:   diamond2,
		Diamond1:   diamond1,
		Platinum3:  platinum3,
		Platinum2:  platinum2,
		Platinum1:  platinum1,
		Gold3:      gold3,
		Gold2:      gold2,
		Gold1:      gold1,
		Silver3:    silver3,
		Silver2:    silver2,
		Silver1:    silver1,
		Bronze3:    bronze3,
		Bronze2:    bronze2,
		Bronze1:    bronze1,
		Iron3:      iron3,
		Iron2:      iron2,
		Iron1:      iron1,
	}
}
