package exception

import "net/http"

var (
	// Discord validation errors
	ErrDiscordGuildNotFound       = NewBusinessError(http.StatusNotFound, "discord guild not found", "DC001")
	ErrBotNotInGuild              = NewBusinessError(http.StatusBadRequest, "GAMERS bot is not in the specified guild", "DC002")
	ErrUserNotInGuild             = NewBusinessError(http.StatusForbidden, "user is not a member of the specified guild", "DC003")
	ErrInvalidDiscordChannel      = NewBusinessError(http.StatusBadRequest, "invalid discord channel", "DC004")
	ErrChannelNotInGuild          = NewBusinessError(http.StatusBadRequest, "channel is not in the specified guild", "DC005")
	ErrDiscordAPIError            = NewBusinessError(http.StatusBadGateway, "discord API error", "DC006")
	ErrDiscordAccessTokenRequired = NewBusinessError(http.StatusUnauthorized, "discord access token is required", "DC007")
	ErrDiscordGuildRequired       = NewBusinessError(http.StatusBadRequest, "discord guild id is required for tournament contests", "DC008")
	ErrDiscordChannelRequired     = NewBusinessError(http.StatusBadRequest, "discord text channel id is required when guild is specified", "DC009")
	ErrDiscordAccountNotFound     = NewBusinessError(http.StatusNotFound, "discord account not linked to this user", "DC010")
	ErrDiscordTokenNotFound       = NewBusinessError(http.StatusUnauthorized, "discord token not found, please re-authenticate with Discord", "DC011")
	ErrDiscordTokenExpired        = NewBusinessError(http.StatusUnauthorized, "discord token expired, please re-authenticate with Discord", "DC012")
)
