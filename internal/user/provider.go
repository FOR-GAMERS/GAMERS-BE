package user

import (
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/security/password"
	oauth2Port "GAMERS-BE/internal/oauth2/application/port"
	"GAMERS-BE/internal/user/application"
	userCommandPort "GAMERS-BE/internal/user/application/port/command"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"
	userCommand "GAMERS-BE/internal/user/infra/persistence/command"
	userQuery "GAMERS-BE/internal/user/infra/persistence/query"
	"GAMERS-BE/internal/user/presentation"

	"gorm.io/gorm"
)

type Dependencies struct {
	Controller      *presentation.UserController
	UserQueryRepo   userQueryPort.UserQueryPort
	UserCommandRepo userCommandPort.UserCommandPort
}

func ProvideUserDependencies(db *gorm.DB, router *router.Router, oauth2DbPort oauth2Port.OAuth2DatabasePort) *Dependencies {
	passwordHasher := password.NewBcryptPasswordHasher()

	userQueryAdapter := userQuery.NewMysqlUserRepository(db)
	userCommandAdapter := userCommand.NewMySQLUserRepository(db)

	userService := application.NewUserService(
		userQueryAdapter,
		userCommandAdapter,
		passwordHasher,
		oauth2DbPort,
	)

	userController := presentation.NewUserController(router, userService)

	return &Dependencies{
		Controller:      userController,
		UserQueryRepo:   userQueryAdapter,
		UserCommandRepo: userCommandAdapter,
	}
}
