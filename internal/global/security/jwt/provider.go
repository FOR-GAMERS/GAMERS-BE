package jwt

import (
	"GAMERS-BE/internal/global/security/jwt/application"
	"GAMERS-BE/internal/global/security/jwt/domain"
	"GAMERS-BE/internal/global/security/jwt/infra"
)

func ProvideJwtService() *application.TokenService {
	config := domain.NewConfigFromEnv()
	accessStrategy := infra.NewAccessTokenStrategy(config)
	refreshStrategy := infra.NewRefreshTokenStrategy(config)

	tokenService := application.NewTokenService()
	tokenService.RegisterStrategy(accessStrategy)
	tokenService.RegisterStrategy(refreshStrategy)

	return tokenService
}
