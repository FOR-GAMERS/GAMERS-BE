//go:build wireinject
// +build wireinject

package main

import (
	"GAMERS-BE/internal/common/security/password"
	"GAMERS-BE/internal/user/application"
	commandport "GAMERS-BE/internal/user/application/port/command"
	queryport "GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/user/infra/persistence/command"
	"GAMERS-BE/internal/user/infra/persistence/query"
	"GAMERS-BE/internal/user/presentation"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// InitializeUserController creates UserController with all dependencies
func InitializeUserController(db *gorm.DB) *presentation.UserController {
	wire.Build(
		// Password Hasher
		password.NewBcryptPasswordHasher,
		wire.Bind(new(password.PasswordHasher), new(*password.BcryptPasswordHasher)),

		// User repositories
		query.NewMysqlUserRepository,
		wire.Bind(new(queryport.UserQueryPort), new(*query.MYSQLUserRepository)),

		command.NewMySQLUserRepository,
		wire.Bind(new(commandport.UserCommandPort), new(*command.MySQLUserRepository)),

		// Profile repositories
		command.NewMysqlProfileCommandAdapter,
		wire.Bind(new(commandport.ProfileCommandPort), new(*command.MYSQLProfileCommandAdapter)),

		// User Service
		application.NewUserService,

		// User Controller
		presentation.NewUserController,
	)
	return nil
}

// InitializeProfileController creates ProfileController with all dependencies
func InitializeProfileController(db *gorm.DB) *presentation.ProfileController {
	wire.Build(
		// Profile repositories
		query.NewMysqlProfileQueryAdapter,
		wire.Bind(new(queryport.ProfileQueryPort), new(*query.MYSQLProfileQueryAdapter)),

		command.NewMysqlProfileCommandAdapter,
		wire.Bind(new(commandport.ProfileCommandPort), new(*command.MYSQLProfileCommandAdapter)),

		// Profile Service
		application.NewProfileService,

		// Profile Controller
		presentation.NewProfileController,
	)
	return nil
}
