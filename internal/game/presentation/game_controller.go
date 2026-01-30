package presentation

import (
	"GAMERS-BE/internal/auth/middleware"
	"GAMERS-BE/internal/game/application"
	gameDto "GAMERS-BE/internal/game/application/dto"
	gameDomain "GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GameController struct {
	router                  *router.Router
	service                 *application.GameService
	matchDetectionSvc       *application.MatchDetectionService
	tournamentResultSvc     *application.TournamentResultService
	helper                  *handler.ControllerHelper
}

func NewGameController(
	router *router.Router,
	service *application.GameService,
	matchDetectionSvc *application.MatchDetectionService,
	tournamentResultSvc *application.TournamentResultService,
	helper *handler.ControllerHelper,
) *GameController {
	return &GameController{
		router:              router,
		service:             service,
		matchDetectionSvc:   matchDetectionSvc,
		tournamentResultSvc: tournamentResultSvc,
		helper:              helper,
	}
}

func (c *GameController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/games")
	{
		privateGroup.POST("", c.CreateGame)
		privateGroup.PATCH("/:id", c.UpdateGame)
		privateGroup.DELETE("/:id", c.DeleteGame)
		privateGroup.POST("/:id/start", c.StartGame)
		privateGroup.POST("/:id/finish", c.FinishGame)
		privateGroup.POST("/:id/cancel", c.CancelGame)
	}

	publicGroup := c.router.PublicGroup("/api/games")
	{
		publicGroup.GET("/:id", c.GetGame)
	}

	contestGamesGroup := c.router.PublicGroup("/api/contests")
	{
		contestGamesGroup.GET("/:id/games", c.GetGamesByContest)
	}

	// Match Detection endpoints
	contestGamesProtected := c.router.ProtectedGroup("/api/contests")
	{
		contestGamesProtected.PUT("/:id/games/:gameId/schedule", c.ScheduleGame)
		contestGamesProtected.POST("/:id/games/:gameId/result", c.SubmitManualResult)
		contestGamesProtected.POST("/:id/games/:gameId/detect", c.TriggerDetection)
	}

	contestGamesPublic := c.router.PublicGroup("/api/contests")
	{
		contestGamesPublic.GET("/:id/games/:gameId/detection-status", c.GetDetectionStatus)
		contestGamesPublic.GET("/:id/games/:gameId/result", c.GetMatchResult)
		contestGamesPublic.GET("/:id/games/:gameId/result/stats", c.GetMatchResultWithStats)
		contestGamesPublic.GET("/:id/result", c.GetContestResult)
	}
}

// CreateGame godoc
// @Summary Create a new game
// @Description Create a new game under a contest (creator becomes the leader)
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param game body gameDto.CreateGameRequest true "Game creation request"
// @Success 201 {object} response.Response{data=gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/games [post]
func (c *GameController) CreateGame(ctx *gin.Context) {
	var req gameDto.CreateGameRequest

	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	game, err := c.service.CreateGame(&req)
	c.helper.RespondCreated(ctx, gameDto.ToGameResponse(game), err, "game created successfully")
}

// GetGame godoc
// @Summary Get a game by ID
// @Description Get game details by game ID
// @Tags games
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id} [get]
func (c *GameController) GetGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	game, err := c.service.GetGame(gameID)
	c.helper.RespondOK(ctx, gameDto.ToGameResponse(game), err, "game retrieved successfully")
}

// GetGamesByContest godoc
// @Summary Get all games for a contest
// @Description Get all games under a specific contest
// @Tags games
// @Accept json
// @Produce json
// @Param id path int true "Contest ID"
// @Success 200 {object} response.Response{data=[]gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Router /api/contests/{id}/games [get]
func (c *GameController) GetGamesByContest(ctx *gin.Context) {
	contestID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	games, err := c.service.GetGamesByContest(contestID)
	if err != nil {
		c.helper.RespondOK(ctx, nil, err, "")
		return
	}

	responses := make([]*gameDto.GameResponse, len(games))
	for i, game := range games {
		responses[i] = gameDto.ToGameResponse(game)
	}

	c.helper.RespondOK(ctx, responses, nil, "games retrieved successfully")
}

// UpdateGame godoc
// @Summary Update a game
// @Description Update game details by game ID (Leader only, PENDING status only)
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param game body gameDto.UpdateGameRequest true "Game update request"
// @Success 200 {object} response.Response{data=gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id} [patch]
func (c *GameController) UpdateGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	var req gameDto.UpdateGameRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	game, err := c.service.UpdateGame(gameID, &req)
	c.helper.RespondOK(ctx, gameDto.ToGameResponse(game), err, "game updated successfully")
}

// DeleteGame godoc
// @Summary Delete a game
// @Description Delete a game by game ID (Leader only, PENDING status only)
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id} [delete]
func (c *GameController) DeleteGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.service.DeleteGame(gameID, userID)
	c.helper.RespondNoContent(ctx, err)
}

// StartGame godoc
// @Summary Start a game
// @Description Transition game status from PENDING to ACTIVE (Leader only)
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/start [post]
func (c *GameController) StartGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	game, err := c.service.StartGame(gameID)
	c.helper.RespondOK(ctx, gameDto.ToGameResponse(game), err, "game started successfully")
}

// FinishGame godoc
// @Summary Finish a game
// @Description Transition game status from ACTIVE to FINISHED (Leader only)
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/finish [post]
func (c *GameController) FinishGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	game, err := c.service.FinishGame(gameID)
	c.helper.RespondOK(ctx, gameDto.ToGameResponse(game), err, "game finished successfully")
}

// CancelGame godoc
// @Summary Cancel a game
// @Description Transition game status to CANCELLED (Leader only)
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.GameResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/cancel [post]
func (c *GameController) CancelGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	game, err := c.service.CancelGame(gameID, userID)
	c.helper.RespondOK(ctx, gameDto.ToGameResponse(game), err, "game cancelled successfully")
}

// ScheduleGame godoc
// @Summary Set game scheduled start time
// @Description Staff sets the scheduled start time and detection window for a tournament game
// @Tags games, match-detection
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param gameId path int true "Game ID"
// @Param body body gameDto.ScheduleGameRequest true "Schedule request"
// @Success 200 {object} response.Response{data=gameDto.ScheduleGameResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/games/{gameId}/schedule [put]
func (c *GameController) ScheduleGame(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("gameId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	var req gameDto.ScheduleGameRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	game, err := c.service.ScheduleGame(gameID, &req)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	c.helper.RespondOK(ctx, gameDto.ToScheduleGameResponse(game), nil, "game scheduled successfully")
}

// SubmitManualResult godoc
// @Summary Submit manual game result
// @Description Staff manually inputs the game result (fallback when auto-detection fails)
// @Tags games, match-detection
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param gameId path int true "Game ID"
// @Param body body gameDto.ManualResultRequest true "Manual result request"
// @Success 201 {object} response.Response{data=gameDto.MatchResultResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/games/{gameId}/result [post]
func (c *GameController) SubmitManualResult(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("gameId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	var req gameDto.ManualResultRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	result, err := c.matchDetectionSvc.SubmitManualResult(gameID, &req)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	c.helper.RespondCreated(ctx, gameDto.ToMatchResultResponse(result, gameDomain.DetectionStatusManual), nil, "manual result submitted successfully")
}

// GetDetectionStatus godoc
// @Summary Get match detection status
// @Description Returns the current detection status for a game
// @Tags games, match-detection
// @Produce json
// @Param contestId path int true "Contest ID"
// @Param gameId path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.ScheduleGameResponse}
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/games/{gameId}/detection-status [get]
func (c *GameController) GetDetectionStatus(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("gameId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	game, err := c.service.GetDetectionStatus(gameID)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	c.helper.RespondOK(ctx, gameDto.ToScheduleGameResponse(game), nil, "detection status retrieved")
}

// GetMatchResult godoc
// @Summary Get match result
// @Description Returns the match result for a finished game
// @Tags games, match-detection
// @Produce json
// @Param contestId path int true "Contest ID"
// @Param gameId path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.MatchResultResponse}
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/games/{gameId}/result [get]
func (c *GameController) GetMatchResult(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("gameId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	result, err := c.matchDetectionSvc.GetMatchResult(gameID)
	c.helper.RespondOK(ctx, result, err, "match result retrieved")
}

// GetMatchResultWithStats godoc
// @Summary Get match result with player stats
// @Description Returns the match result and individual player stats
// @Tags games, match-detection
// @Produce json
// @Param contestId path int true "Contest ID"
// @Param gameId path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.MatchResultResponse}
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/games/{gameId}/result/stats [get]
func (c *GameController) GetMatchResultWithStats(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("gameId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	result, err := c.matchDetectionSvc.GetMatchResultWithStats(gameID)
	c.helper.RespondOK(ctx, result, err, "match result with stats retrieved")
}

// TriggerDetection godoc
// @Summary Manually trigger match detection
// @Description Staff manually triggers match detection for a specific game
// @Tags games, match-detection
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param gameId path int true "Game ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/games/{gameId}/detect [post]
func (c *GameController) TriggerDetection(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("gameId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.matchDetectionSvc.DetectMatchForGame(gameID)
	if err != nil {
		c.helper.HandleError(ctx, err)
		return
	}

	c.helper.RespondOK(ctx, nil, nil, "match detection triggered")
}

// GetContestResult godoc
// @Summary Get tournament contest result
// @Description Returns the full tournament bracket with game results and champion
// @Tags games, tournaments
// @Produce json
// @Param id path int true "Contest ID"
// @Success 200 {object} response.Response{data=gameDto.ContestResultResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/result [get]
func (c *GameController) GetContestResult(ctx *gin.Context) {
	contestID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	result, err := c.tournamentResultSvc.GetContestResult(contestID)
	c.helper.RespondOK(ctx, result, err, "contest result retrieved")
}
