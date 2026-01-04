package jwt

import (
	"GAMERS-BE/internal/global/utils"
	"time"
)

type Config struct {
	SecretKey            string
	RefreshSecretKey     string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

func NewConfigFromEnv() *Config {
	return &Config{
		SecretKey:            utils.GetEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		RefreshSecretKey:     utils.GetEnv("JWT_REFRESH_SECRET", "your-secret-key-change-in-production"),
		AccessTokenDuration:  utils.GetDurationEnv("JWT_ACCESS_DURATION", 30*time.Minute),
		RefreshTokenDuration: utils.GetDurationEnv("JWT_REFRESH_DURATION", 7*24*time.Hour),
		Issuer:               utils.GetEnv("JWT_ISSUER", "gamers-api"),
	}
}
