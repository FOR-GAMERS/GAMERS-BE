package exception

import "net/http"

var (
	ErrDiscordCannotGetUserInfo = NewBusinessError(http.StatusUnauthorized, "failed to get user info from Discord", "OA001")
	ErrDiscordUserCannotFound   = NewBusinessError(http.StatusNotFound, "failed to find user", "OA002")
	ErrDiscordTokenExchange     = NewBusinessError(http.StatusUnauthorized, "failed to exchange token", "OA003")
)
