package user

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/password"
	oauth2Port "github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user/application"
	userCommandPort "github.com/FOR-GAMERS/GAMERS-BE/internal/user/application/port/command"
	userQueryPort "github.com/FOR-GAMERS/GAMERS-BE/internal/user/application/port/port"
	userCommand "github.com/FOR-GAMERS/GAMERS-BE/internal/user/infra/persistence/command"
	userQuery "github.com/FOR-GAMERS/GAMERS-BE/internal/user/infra/persistence/query"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user/presentation"

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
