package dto

type CreateProfileRequest struct {
	UserId   int64  `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	Tag      string `json:"tag" binding:"required"`
	Bio      string `json:"bio"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
	Tag      string `json:"tag" binding:"required"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

type ProfileResponse struct {
	Id       int64  `json:"profile_id"`
	Username string `json:"username"`
	Tag      string `json:"tag"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}
