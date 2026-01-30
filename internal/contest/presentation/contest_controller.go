package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/middleware"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	commonDto "github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContestController struct {
	router  *router.Router
	service *application.ContestService
	helper  *handler.ControllerHelper
}

func NewContestController(router *router.Router, service *application.ContestService, helper *handler.ControllerHelper) *ContestController {
	return &ContestController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *ContestController) RegisterRoute() {
	privateGroup := c.router.ProtectedGroup("/api/contests")
	privateGroup.POST("", c.SaveContest)
	privateGroup.GET("/me", c.GetMyContests)
	privateGroup.PATCH("/:id", c.UpdateContest)
	privateGroup.DELETE("/:id", c.DeleteContest)
	privateGroup.POST("/:id/start", c.StartContest)
	privateGroup.POST("/:id/stop", c.StopContest)

	publicGroup := c.router.PublicGroup("/api/contests")
	publicGroup.GET("", c.GetAllContests)
	publicGroup.GET("/:id", c.GetContestById)
}

// SaveContest godoc
// @Summary Create a new contest
// @Description Create a new contest with contest details
// @Tags contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contest body dto.CreateContestRequest true "Contest creation request"
// @Success 201 {object} response.Response{data=dto.ContestResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response{data=dto.DiscordLinkRequiredResponse}
// @Router /api/contests [post]
func (c *ContestController) SaveContest(ctx *gin.Context) {
	var req dto.CreateContestRequest

	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	contest, discordLinkRequired, err := c.service.SaveContest(&req, userId)

	// Handle Discord link required error
	if errors.Is(err, exception.ErrDiscordLinkRequired) {
		response.JSON(ctx, response.Forbidden(discordLinkRequired, "discord linking required"))
		return
	}

	c.helper.RespondCreated(ctx, contest, err, "contest created successfully")
}

// GetAllContests godoc
// @Summary Get all contests with pagination
// @Description Get all contests with pagination, sorting, and title search support
// @Tags contests
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param sort_by query string false "Sort field: created_at, started_at, ended_at (default: created_at)"
// @Param order query string false "Sort order: asc, desc (default: desc)"
// @Param title query string false "Search by contest title (partial match)"
// @Success 200 {object} response.Response{data=commonDto.PaginationResponse}
// @Failure 400 {object} response.Response
// @Router /api/contests [get]
func (c *ContestController) GetAllContests(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	order := ctx.DefaultQuery("order", "desc")
	titleParam := ctx.Query("title")

	var title *string
	if titleParam != "" {
		title = &titleParam
	}

	paginationReq := commonDto.NewPaginationRequest(page, pageSize)
	allowedSortFields := []string{"created_at", "started_at", "ended_at"}
	sortReq := commonDto.NewSortRequest(sortBy, order, allowedSortFields)

	contests, totalCount, err := c.service.GetAllContests(paginationReq.GetOffset(), paginationReq.GetLimit(), sortReq, title)
	if err != nil {
		response.JSON(ctx, response.Error(400, err.Error()))
		return
	}

	paginationResp := commonDto.NewPaginationResponse(contests, paginationReq.Page, paginationReq.PageSize, totalCount)
	c.helper.RespondOK(ctx, paginationResp, nil, "contests retrieved successfully")
}

// GetContestById godoc
// @Summary Get a contest by ID
// @Description Get contest details by contest ID
// @Tags contests
// @Accept json
// @Produce json
// @Param id path int true "Contest ID"
// @Success 200 {object} response.Response{data=dto.ContestResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id} [get]
func (c *ContestController) GetContestById(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	contest, err := c.service.GetContestById(contestId)
	c.helper.RespondOK(ctx, contest, err, "contest retrieved successfully")
}

// UpdateContest godoc
// @Summary Update a contest
// @Description Update contest details by contest ID
// @Tags contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param contest body dto.UpdateContestRequest true "Contest update request"
// @Success 200 {object} response.Response{data=dto.ContestResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id} [patch]
func (c *ContestController) UpdateContest(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)

	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	var req dto.UpdateContestRequest

	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	contest, err := c.service.UpdateContest(id, &req)
	c.helper.RespondOK(ctx, contest, err, "contest updated successfully")
}

// DeleteContest godoc
// @Summary Delete a contest
// @Description Delete contest by contest ID
// @Tags contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id} [delete]
func (c *ContestController) DeleteContest(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)

	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	err = c.service.DeleteContestById(id)
	c.helper.RespondNoContent(ctx, err)
}

// StartContest godoc
// @Summary Start a contest
// @Description Start a contest and migrate accepted applications to config (Leader only)
// @Tags contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Success 200 {object} response.Response{data=dto.ContestResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/start [post]
func (c *ContestController) StartContest(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	contest, err := c.service.StartContest(ctx.Request.Context(), id, userId)
	c.helper.RespondOK(ctx, contest, err, "contest started successfully")
}

// StopContest godoc
// @Summary Stop a contest
// @Description Stop an active contest and transition it to finished status (Leader only)
// @Tags contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Success 200 {object} response.Response{data=dto.ContestResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/stop [post]
func (c *ContestController) StopContest(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	contest, err := c.service.StopContest(ctx.Request.Context(), id, userId)
	c.helper.RespondOK(ctx, contest, err, "contest stopped successfully")
}

// GetMyContests godoc
// @Summary Get contests I have joined
// @Description Get all contests that the authenticated user has joined with pagination, sorting, and filtering support
// @Tags contests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param sort_by query string false "Sort field: created_at, started_at, ended_at, point, contest_status (default: created_at)"
// @Param order query string false "Sort order: asc, desc (default: desc)"
// @Param status query string false "Filter by contest status: PENDING, ACTIVE, FINISHED, CANCELLED"
// @Success 200 {object} response.Response{data=commonDto.PaginationResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/contests/me [get]
func (c *ContestController) GetMyContests(ctx *gin.Context) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	order := ctx.DefaultQuery("order", "desc")
	statusParam := ctx.Query("status")

	// Parse and validate status filter
	var status *domain.ContestStatus
	if statusParam != "" {
		normalizedStatus := domain.ContestStatus(strings.ToUpper(statusParam))
		switch normalizedStatus {
		case domain.ContestStatusPending, domain.ContestStatusActive, domain.ContestStatusFinished, domain.ContestStatusCancelled:
			status = &normalizedStatus
		default:
			response.JSON(ctx, response.BadRequest("invalid status value. allowed: PENDING, ACTIVE, FINISHED, CANCELLED"))
			return
		}
	}

	paginationReq := commonDto.NewPaginationRequest(page, pageSize)
	allowedSortFields := []string{"created_at", "started_at", "ended_at", "point", "contest_status"}
	sortReq := commonDto.NewSortRequest(sortBy, order, allowedSortFields)

	contests, totalCount, err := c.service.GetMyContests(userId, paginationReq, sortReq, status)
	if err != nil {
		response.JSON(ctx, response.Error(400, err.Error()))
		return
	}

	contestResponses := dto.ToMyContestResponses(contests)
	paginationResp := commonDto.NewPaginationResponse(contestResponses, paginationReq.Page, paginationReq.PageSize, totalCount)
	c.helper.RespondOK(ctx, paginationResp, nil, "my contests retrieved successfully")
}

// ==================== Testable Handler Functions ====================
// These functions are exposed for unit testing without router dependency

func HandleGetContestById(ctx *gin.Context, service *application.ContestService, helper *handler.ControllerHelper) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	contest, err := service.GetContestById(contestId)
	helper.RespondOK(ctx, contest, err, "contest retrieved successfully")
}

func HandleGetAllContests(ctx *gin.Context, service *application.ContestService, helper *handler.ControllerHelper) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	order := ctx.DefaultQuery("order", "desc")
	titleParam := ctx.Query("title")

	var title *string
	if titleParam != "" {
		title = &titleParam
	}

	paginationReq := commonDto.NewPaginationRequest(page, pageSize)
	allowedSortFields := []string{"created_at", "started_at", "ended_at"}
	sortReq := commonDto.NewSortRequest(sortBy, order, allowedSortFields)

	contests, totalCount, err := service.GetAllContests(paginationReq.GetOffset(), paginationReq.GetLimit(), sortReq, title)
	if err != nil {
		response.JSON(ctx, response.Error(400, err.Error()))
		return
	}

	paginationResp := commonDto.NewPaginationResponse(contests, paginationReq.Page, paginationReq.PageSize, totalCount)
	helper.RespondOK(ctx, paginationResp, nil, "contests retrieved successfully")
}

func HandleDeleteContest(ctx *gin.Context, service *application.ContestService, helper *handler.ControllerHelper) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	err = service.DeleteContestById(id)
	helper.RespondNoContent(ctx, err)
}

func HandleUpdateContest(ctx *gin.Context, service *application.ContestService, helper *handler.ControllerHelper) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	var req dto.UpdateContestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest(err.Error()))
		return
	}

	contest, err := service.UpdateContest(id, &req)
	helper.RespondOK(ctx, contest, err, "contest updated successfully")
}

func HandleStartContest(ctx *gin.Context, service *application.ContestService, helper *handler.ControllerHelper) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	contest, err := service.StartContest(ctx.Request.Context(), id, userId)
	helper.RespondOK(ctx, contest, err, "contest started successfully")
}

func HandleStopContest(ctx *gin.Context, service *application.ContestService, helper *handler.ControllerHelper) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	contest, err := service.StopContest(ctx.Request.Context(), id, userId)
	helper.RespondOK(ctx, contest, err, "contest stopped successfully")
}
