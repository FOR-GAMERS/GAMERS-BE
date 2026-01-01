package jwt

import (
	"os"
	"strconv"
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
		SecretKey:            getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		RefreshSecretKey:     getEnv("JWT_REFRESH_SECRET", "your-secret-key-change-in-production"),
		AccessTokenDuration:  getDurationEnv("JWT_ACCESS_DURATION", 30*time.Minute),
		RefreshTokenDuration: getDurationEnv("JWT_REFRESH_DURATION", 7*24*time.Hour), // 7 days
		Issuer:               getEnv("JWT_ISSUER", "gamers-api"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}
