package auth

import (
	"GAMERS-BE/internal/auth/application"
	"GAMERS-BE/internal/auth/infra/jwt"
	authCommand "GAMERS-BE/internal/auth/infra/persistence/command"
	authQuery "GAMERS-BE/internal/auth/infra/persistence/query"
	"GAMERS-BE/internal/auth/presentation"
	"GAMERS-BE/internal/auth/presentation/middleware"
	"GAMERS-BE/internal/global/security/password"
	authUserQuery "GAMERS-BE/internal/user/infra/persistence/query"
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Controller     *presentation.AuthController
	AuthMiddleware *middleware.AuthMiddleware
}

func ProvideAuthDependencies(db *gorm.DB, redisClient *redis.Client, ctx context.Context) *Dependencies {
	jwtConfig := jwt.NewConfigFromEnv()

	tokenManager := jwt.NewTokenManager(jwtConfig)

	tokenProvider := jwt.NewTokenProvider(tokenManager)

	passwordHasher := password.NewBcryptPasswordHasher()

	authUserQueryAdapter := authUserQuery.NewAuthUserQueryAdapter(db)
	refreshTokenCommandAdapter := authCommand.NewRefreshTokenRedisCommandAdapter(redisClient)
	refreshTokenQueryAdapter := authQuery.NewRefreshTokenRedisQueryAdapter(redisClient)

	authService := application.NewAuthService(
		ctx,
		authUserQueryAdapter,
		refreshTokenCommandAdapter,
		refreshTokenQueryAdapter,
		tokenProvider,
		passwordHasher,
	)

	authController := presentation.NewAuthController(authService)

	authMiddleware := middleware.NewAuthMiddleware(tokenManager)

	return &Dependencies{
		Controller:     authController,
		AuthMiddleware: authMiddleware,
	}
}
