package dto

import "time"

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
	Tag      string `json:"tag" binding:"required"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

type UpdateUserRequest struct {
	Password string `json:"password" binding:"required"`
}

type UpdateUserInfoRequest struct {
	Username *string `json:"username"`
	Tag      *string `json:"tag"`
	Bio      *string `json:"bio"`
	Avatar   *string `json:"avatar"`
}

type UserResponse struct {
	Id         int64     `json:"user_id"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type MyUserResponse struct {
	Id                 int64      `json:"user_id"`
	Email              string     `json:"email"`
	Username           string     `json:"username"`
	Tag                string     `json:"tag"`
	Bio                string     `json:"bio"`
	Avatar             string     `json:"avatar"`
	ProfileKey         *string    `json:"profile_key,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ModifiedAt         time.Time  `json:"modified_at"`
	RiotName           *string    `json:"riot_name,omitempty"`
	RiotTag            *string    `json:"riot_tag,omitempty"`
	Region             *string    `json:"region,omitempty"`
	CurrentTier        *int       `json:"current_tier,omitempty"`
	CurrentTierPatched *string    `json:"current_tier_patched,omitempty"`
	Elo                *int       `json:"elo,omitempty"`
	RankingInTier      *int       `json:"ranking_in_tier,omitempty"`
	PeakTier           *int       `json:"peak_tier,omitempty"`
	PeakTierPatched    *string    `json:"peak_tier_patched,omitempty"`
	ValorantUpdatedAt  *time.Time `json:"valorant_updated_at,omitempty"`
}
