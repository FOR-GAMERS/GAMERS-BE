package user

import (
	"GAMERS-BE/internal/global/security/password"
	"GAMERS-BE/internal/user/application"
	userCommand "GAMERS-BE/internal/user/infra/persistence/command"
	userQuery "GAMERS-BE/internal/user/infra/persistence/query"
	"GAMERS-BE/internal/user/presentation"

	"gorm.io/gorm"
)

type Dependencies struct {
	Controller *presentation.UserController
}

func ProvideUserDependencies(db *gorm.DB) *Dependencies {
	passwordHasher := password.NewBcryptPasswordHasher()

	userQueryAdapter := userQuery.NewMysqlUserRepository(db)
	userCommandAdapter := userCommand.NewMySQLUserRepository(db)

	userService := application.NewUserService(
		userQueryAdapter,
		userCommandAdapter,
		passwordHasher,
	)

	userController := presentation.NewUserController(userService)

	return &Dependencies{
		Controller: userController,
	}
}
