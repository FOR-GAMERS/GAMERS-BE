package presentation

import (
	"GAMERS-BE/internal/game/application"
	gameDto "GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GameTeamController struct {
	router  *router.Router
	service *application.GameTeamService
	helper  *handler.ControllerHelper
}

func NewGameTeamController(
	router *router.Router,
	service *application.GameTeamService,
	helper *handler.ControllerHelper,
) *GameTeamController {
	return &GameTeamController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *GameTeamController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/game-teams")
	{
		privateGroup.POST("", c.CreateGameTeam)
		privateGroup.DELETE("/:id", c.DeleteGameTeam)
	}

	publicGroup := c.router.PublicGroup("/api/game-teams")
	{
		publicGroup.GET("/:id", c.GetGameTeam)
	}

	gameTeamsGroup := c.router.PublicGroup("/api/games/:id/game-teams")
	{
		gameTeamsGroup.GET("", c.GetGameTeamsByGame)
	}
}

// CreateGameTeam godoc
// @Summary Create a new game-team relationship
// @Description Register a team to participate in a game with optional grade
// @Tags game-teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param gameTeam body gameDto.CreateGameTeamRequest true "Game team creation request"
// @Success 201 {object} response.Response{data=gameDto.GameTeamResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/game-teams [post]
func (c *GameTeamController) CreateGameTeam(ctx *gin.Context) {
	var req gameDto.CreateGameTeamRequest

	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	gameTeam, err := c.service.CreateGameTeam(&req)
	c.helper.RespondCreated(ctx, gameDto.ToGameTeamResponse(gameTeam), err, "game team created successfully")
}

// GetGameTeam godoc
// @Summary Get a game-team relationship by ID
// @Description Get game-team details by game team ID
// @Tags game-teams
// @Accept json
// @Produce json
// @Param id path int true "Game Team ID"
// @Success 200 {object} response.Response{data=gameDto.GameTeamResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/game-teams/{id} [get]
func (c *GameTeamController) GetGameTeam(ctx *gin.Context) {
	gameTeamID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game team id"))
		return
	}

	gameTeam, err := c.service.GetGameTeam(gameTeamID)
	c.helper.RespondOK(ctx, gameDto.ToGameTeamResponse(gameTeam), err, "game team retrieved successfully")
}

// GetGameTeamsByGame godoc
// @Summary Get all game-teams for a game
// @Description Get all teams participating in a specific game
// @Tags game-teams
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=[]gameDto.GameTeamResponse}
// @Failure 400 {object} response.Response
// @Router /api/games/{id}/game-teams [get]
func (c *GameTeamController) GetGameTeamsByGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	gameTeams, err := c.service.GetGameTeamsByGame(gameID)
	if err != nil {
		c.helper.RespondOK(ctx, nil, err, "")
		return
	}

	c.helper.RespondOK(ctx, gameDto.ToGameTeamResponseList(gameTeams), nil, "game teams retrieved successfully")
}

// DeleteGameTeam godoc
// @Summary Delete a game-team relationship
// @Description Remove a team from a game
// @Tags game-teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game Team ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/game-teams/{id} [delete]
func (c *GameTeamController) DeleteGameTeam(ctx *gin.Context) {
	gameTeamID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game team id"))
		return
	}

	err = c.service.DeleteGameTeam(gameTeamID)
	c.helper.RespondNoContent(ctx, err)
}
