package comment

import (
	"GAMERS-BE/internal/comment/application"
	"GAMERS-BE/internal/comment/infra/persistence/adapter"
	"GAMERS-BE/internal/comment/presentation"
	contestPort "GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"

	"gorm.io/gorm"
)

type Dependencies struct {
	Controller *presentation.CommentController
}

func ProvideCommentDependencies(
	db *gorm.DB,
	router *router.Router,
	contestRepository contestPort.ContestDatabasePort,
) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	commentDatabaseAdapter := adapter.NewCommentDatabaseAdapter(db)

	commentService := application.NewCommentService(
		commentDatabaseAdapter,
		contestRepository,
	)

	commentController := presentation.NewCommentController(
		router,
		commentService,
		controllerHelper,
	)

	return &Dependencies{
		Controller: commentController,
	}
}
