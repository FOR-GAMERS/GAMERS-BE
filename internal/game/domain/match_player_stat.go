package domain

// MatchPlayerStat stores individual player statistics from a detected Valorant match
type MatchPlayerStat struct {
	MatchPlayerStatID int64  `gorm:"column:match_player_stat_id;primaryKey;autoIncrement" json:"match_player_stat_id"`
	MatchResultID     int64  `gorm:"column:match_result_id;type:bigint unsigned;not null;index:idx_match_player_stats_result" json:"match_result_id"`
	UserID            int64  `gorm:"column:user_id;type:bigint unsigned;not null;index:idx_match_player_stats_user" json:"user_id"`
	TeamID            int64  `gorm:"column:team_id;type:bigint unsigned;not null" json:"team_id"`
	AgentName         string `gorm:"column:agent_name;type:varchar(50)" json:"agent_name"`
	Kills             int    `gorm:"column:kills;type:int;not null;default:0" json:"kills"`
	Deaths            int    `gorm:"column:deaths;type:int;not null;default:0" json:"deaths"`
	Assists           int    `gorm:"column:assists;type:int;not null;default:0" json:"assists"`
	Score             int    `gorm:"column:score;type:int;not null;default:0" json:"score"`
	Headshots         int    `gorm:"column:headshots;type:int;not null;default:0" json:"headshots"`
	Bodyshots         int    `gorm:"column:bodyshots;type:int;not null;default:0" json:"bodyshots"`
	Legshots          int    `gorm:"column:legshots;type:int;not null;default:0" json:"legshots"`
}

func NewMatchPlayerStat(
	matchResultID, userID, teamID int64,
	agentName string,
	kills, deaths, assists, score int,
	headshots, bodyshots, legshots int,
) *MatchPlayerStat {
	return &MatchPlayerStat{
		MatchResultID: matchResultID,
		UserID:        userID,
		TeamID:        teamID,
		AgentName:     agentName,
		Kills:         kills,
		Deaths:        deaths,
		Assists:       assists,
		Score:         score,
		Headshots:     headshots,
		Bodyshots:     bodyshots,
		Legshots:      legshots,
	}
}

func (m *MatchPlayerStat) TableName() string {
	return "match_player_stats"
}
