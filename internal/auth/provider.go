package auth

import (
	"GAMERS-BE/internal/auth/application"
	"GAMERS-BE/internal/auth/infra/persistence/adapter"
	"GAMERS-BE/internal/auth/presentation"
	"GAMERS-BE/internal/global/common/router"
	jwtProvider "GAMERS-BE/internal/global/security/jwt"
	"GAMERS-BE/internal/global/security/password"
	authUserQuery "GAMERS-BE/internal/user/infra/persistence/query"
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Controller *presentation.AuthController
}

func ProvideAuthDependencies(db *gorm.DB, redisClient *redis.Client, ctx *context.Context, router *router.Router) *Dependencies {
	passwordHasher := password.NewBcryptPasswordHasher()

	authUserQueryAdapter := authUserQuery.NewAuthUserQueryAdapter(db)
	refreshCacheAdapter := adapter.NewRefreshTokenCacheAdapter(ctx, redisClient)
	tokenService := jwtProvider.ProvideJwtService()

	authService := application.NewAuthService(
		authUserQueryAdapter,
		refreshCacheAdapter,
		*tokenService,
		passwordHasher,
	)

	authController := presentation.NewAuthController(router, authService)

	return &Dependencies{
		Controller: authController,
	}
}
