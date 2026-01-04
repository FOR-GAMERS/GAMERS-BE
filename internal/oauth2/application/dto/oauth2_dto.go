package dto

type DiscordCallbackRequest struct {
	Code string `form:"code" binding:"required"`
}

type DiscordUserInfo struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Verified bool   `json:"verified"`
}

type OAuth2LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IsNewUser    bool   `json:"is_new_user"`
}
