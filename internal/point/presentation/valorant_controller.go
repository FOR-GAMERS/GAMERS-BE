package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/point/application"
	pointDto "github.com/FOR-GAMERS/GAMERS-BE/internal/point/application/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ValorantController struct {
	router  *router.Router
	service *application.ValorantService
	helper  *handler.ControllerHelper
}

func NewValorantController(
	router *router.Router,
	service *application.ValorantService,
	helper *handler.ControllerHelper,
) *ValorantController {
	return &ValorantController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *ValorantController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/valorant/score-tables")
	{
		privateGroup.POST("", c.CreateScoreTable)
		privateGroup.DELETE("/:id", c.DeleteScoreTable)
	}

	publicGroup := c.router.PublicGroup("/api/valorant/score-tables")
	{
		publicGroup.GET("", c.GetAllScoreTables)
		publicGroup.GET("/:id", c.GetScoreTable)
	}
}

// CreateScoreTable godoc
// @Summary Create a new Valorant score table
// @Description Create a new score table with points for each rank tier
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scoreTable body pointDto.CreateValorantScoreTableDto true "Score table creation request"
// @Success 201 {object} response.Response{data=pointDto.ValorantScoreTableResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/valorant/score-tables [post]
func (c *ValorantController) CreateScoreTable(ctx *gin.Context) {
	var req pointDto.CreateValorantScoreTableDto

	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	scoreTable, err := c.service.CreateScoreTable(&req)
	c.helper.RespondCreated(ctx, pointDto.ToValorantScoreTableResponse(scoreTable), err, "score table created successfully")
}

// GetScoreTable godoc
// @Summary Get a Valorant score table by ID
// @Description Get score table details by score table ID
// @Tags valorant
// @Accept json
// @Produce json
// @Param id path int true "Score Table ID"
// @Success 200 {object} response.Response{data=pointDto.ValorantScoreTableResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/valorant/score-tables/{id} [get]
func (c *ValorantController) GetScoreTable(ctx *gin.Context) {
	scoreTableID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid score table id"))
		return
	}

	scoreTable, err := c.service.GetScoreTable(scoreTableID)
	c.helper.RespondOK(ctx, pointDto.ToValorantScoreTableResponse(scoreTable), err, "score table retrieved successfully")
}

// GetAllScoreTables godoc
// @Summary Get all Valorant score tables
// @Description Get all available score tables
// @Tags valorant
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]pointDto.ValorantScoreTableResponse}
// @Router /api/valorant/score-tables [get]
func (c *ValorantController) GetAllScoreTables(ctx *gin.Context) {
	scoreTables, err := c.service.GetAllScoreTables()
	if err != nil {
		c.helper.RespondOK(ctx, nil, err, "")
		return
	}

	c.helper.RespondOK(ctx, pointDto.ToValorantScoreTableResponseList(scoreTables), nil, "score tables retrieved successfully")
}

// DeleteScoreTable godoc
// @Summary Delete a Valorant score table
// @Description Delete a score table by score table ID
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Score Table ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/valorant/score-tables/{id} [delete]
func (c *ValorantController) DeleteScoreTable(ctx *gin.Context) {
	scoreTableID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid score table id"))
		return
	}

	err = c.service.DeleteScoreTable(scoreTableID)
	c.helper.RespondNoContent(ctx, err)
}
