package point

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/presentation"

	"gorm.io/gorm"
)

type Dependencies struct {
	ValorantController  *presentation.ValorantController
	ScoreTableRepository port.ValorantScoreTableDatabasePort
}

func ProvidePointDependencies(
	db *gorm.DB,
	router *router.Router,
) *Dependencies {
	controllerHelper := handler.NewControllerHelper()

	// Adapters
	valorantScoreTableAdapter := adapter.NewValorantScoreTableDatabaseAdapter(db)

	// Services
	valorantService := application.NewValorantService(
		valorantScoreTableAdapter,
	)

	// Controllers
	valorantController := presentation.NewValorantController(
		router,
		valorantService,
		controllerHelper,
	)

	return &Dependencies{
		ValorantController:   valorantController,
		ScoreTableRepository: valorantScoreTableAdapter,
	}
}
