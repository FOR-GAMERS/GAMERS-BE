package dto

import "time"

// RegisterValorantRequest - Riot ID registration request
type RegisterValorantRequest struct {
	RiotName string `json:"riot_name" binding:"required,max=32" example:"PlayerName"`
	RiotTag  string `json:"riot_tag" binding:"required,max=8" example:"KR1"`
	Region   string `json:"region" binding:"required,oneof=ap br eu kr latam na" example:"ap"`
}

// ValorantInfoResponse - Valorant information response
type ValorantInfoResponse struct {
	RiotName           string     `json:"riot_name" example:"PlayerName"`
	RiotTag            string     `json:"riot_tag" example:"KR1"`
	Region             string     `json:"region" example:"ap"`
	CurrentTier        int        `json:"current_tier" example:"21"`
	CurrentTierPatched string     `json:"current_tier_patched" example:"Diamond 1"`
	Elo                int        `json:"elo" example:"1850"`
	RankingInTier      int        `json:"ranking_in_tier" example:"50"`
	PeakTier           int        `json:"peak_tier" example:"24"`
	PeakTierPatched    string     `json:"peak_tier_patched" example:"Immortal 1"`
	UpdatedAt          *time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	RefreshNeeded      bool       `json:"refresh_needed" example:"false"`
}

// ContestPointResponse - Contest point calculation response
type ContestPointResponse struct {
	UserID             int64  `json:"user_id" example:"123"`
	RiotName           string `json:"riot_name" example:"PlayerName"`
	RiotTag            string `json:"riot_tag" example:"KR1"`
	CurrentTierPatched string `json:"current_tier_patched" example:"Diamond 1"`
	CurrentTierPoint   int    `json:"current_tier_point" example:"100"`
	PeakTierPatched    string `json:"peak_tier_patched" example:"Immortal 1"`
	PeakTierPoint      int    `json:"peak_tier_point" example:"150"`
	FinalPoint         int    `json:"final_point" example:"125"`
	RefreshNeeded      bool   `json:"refresh_needed" example:"false"`
	RefreshMessage     string `json:"refresh_message,omitempty" example:"마지막 갱신 후 24시간이 지났습니다. 갱신이 필요합니다."`
}
