package comment

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/comment/presentation"
	contestPort "github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"

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
