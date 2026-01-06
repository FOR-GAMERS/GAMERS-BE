package presentation

import (
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/application/dto"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"strconv"

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
	privateGroup.PATCH("/:id", c.UpdateContest)
	privateGroup.DELETE("/:id", c.DeleteContest)

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
// @Router /api/contests [post]
func (c *ContestController) SaveContest(ctx *gin.Context) {
	var req dto.CreateContestRequest

	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	contest, err := c.service.SaveContest(&req)
	c.helper.RespondCreated(ctx, contest, err, "contest created successfully")
}

// GetAllContests godoc
// @Summary Get all contests with pagination
// @Description Get all contests with pagination and sorting support
// @Tags contests
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param sort_by query string false "Sort field: created_at, started_at, ended_at (default: created_at)"
// @Param order query string false "Sort order: asc, desc (default: desc)"
// @Success 200 {object} response.Response{data=commonDto.PaginationResponse}
// @Failure 400 {object} response.Response
// @Router /api/contests [get]
func (c *ContestController) GetAllContests(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	order := ctx.DefaultQuery("order", "desc")

	paginationReq := commonDto.NewPaginationRequest(page, pageSize)
	allowedSortFields := []string{"created_at", "started_at", "ended_at"}
	sortReq := commonDto.NewSortRequest(sortBy, order, allowedSortFields)

	contests, totalCount, err := c.service.GetAllContests(paginationReq.GetOffset(), paginationReq.GetLimit(), sortReq)
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
