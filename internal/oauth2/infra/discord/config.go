package discord

import (
	"GAMERS-BE/internal/global/utils"

	"golang.org/x/oauth2"
)

const (
	ClientIDKey     = "DISCORD_CLIENT_ID"
	ClientSecretKey = "DISCORD_CLIENT_SECRET"
	RedirectURLKey  = "DISCORD_REDIRECT_URL"

	AuthUrl  = "https://discord.com/oauth2/authorize"
	TokenUrl = "https://discord.com/api/oauth2/token"
)

type Config struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUri  string `json:"redirect_uri"`
}

func NewConfigFromEnv() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     utils.GetEnv(ClientIDKey, ""),
		ClientSecret: utils.GetEnv(ClientSecretKey, ""),
		RedirectURL:  utils.GetEnv(RedirectURLKey, ""),
		Scopes:       []string{"identify", "email", "guilds"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthUrl,
			TokenURL: TokenUrl,
		},
	}
}
