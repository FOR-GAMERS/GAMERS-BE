package presentation

import (
	"GAMERS-BE/internal/auth/middleware"
	"GAMERS-BE/internal/comment/application"
	"GAMERS-BE/internal/comment/application/dto"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	router  *router.Router
	service *application.CommentService
	helper  *handler.ControllerHelper
}

func NewCommentController(
	router *router.Router,
	service *application.CommentService,
	helper *handler.ControllerHelper,
) *CommentController {
	return &CommentController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *CommentController) RegisterRoutes() {
	publicGroup := c.router.PublicGroup("/api/contests/:id/comments")
	publicGroup.GET("", c.GetComments)
	publicGroup.GET("/:commentId", c.GetCommentByID)

	privateGroup := c.router.ProtectedGroup("/api/contests/:id/comments")
	privateGroup.POST("", c.CreateComment)
	privateGroup.PATCH("/:commentId", c.UpdateComment)
	privateGroup.DELETE("/:commentId", c.DeleteComment)
}

// CreateComment godoc
// @Summary Create a new comment
// @Description Create a new comment for a contest
// @Tags contest-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param comment body dto.CreateCommentRequest true "Comment creation request"
// @Success 201 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/comments [post]
func (c *CommentController) CreateComment(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	var req dto.CreateCommentRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	comment, err := c.service.CreateComment(contestId, userId, &req)
	c.helper.RespondCreated(ctx, comment, err, "comment created successfully")
}

// GetComments godoc
// @Summary Get comments for a contest
// @Description Get all comments for a contest with pagination
// @Tags contest-comments
// @Accept json
// @Produce json
// @Param id path int true "Contest ID"
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Page size (default: 10, max: 100)" minimum(1) maximum(100)
// @Param sort_by query string false "Sort field (created_at, modified_at)" default(created_at)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} response.Response{data=commonDto.PaginationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/comments [get]
func (c *CommentController) GetComments(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pagination := commonDto.NewPaginationRequest(page, pageSize)

	sortBy := ctx.DefaultQuery("sort_by", "created_at")
	order := ctx.DefaultQuery("order", "desc")
	sort := commonDto.NewSortRequest(sortBy, order, []string{"created_at", "modified_at"})

	result, err := c.service.GetCommentsByContestID(contestId, pagination, sort)
	c.helper.RespondOK(ctx, result, err, "comments retrieved successfully")
}

// GetCommentByID godoc
// @Summary Get a comment by ID
// @Description Get a specific comment by its ID
// @Tags contest-comments
// @Accept json
// @Produce json
// @Param id path int true "Contest ID"
// @Param commentId path int true "Comment ID"
// @Success 200 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/comments/{commentId} [get]
func (c *CommentController) GetCommentByID(ctx *gin.Context) {
	_, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	commentId, err := strconv.ParseInt(ctx.Param("commentId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid comment id"))
		return
	}

	comment, err := c.service.GetCommentByID(commentId)
	c.helper.RespondOK(ctx, comment, err, "comment retrieved successfully")
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update a comment (owner only)
// @Tags contest-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param commentId path int true "Comment ID"
// @Param comment body dto.UpdateCommentRequest true "Comment update request"
// @Success 200 {object} response.Response{data=dto.CommentResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/comments/{commentId} [patch]
func (c *CommentController) UpdateComment(ctx *gin.Context) {
	_, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	commentId, err := strconv.ParseInt(ctx.Param("commentId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid comment id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	var req dto.UpdateCommentRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	comment, err := c.service.UpdateComment(commentId, userId, &req)
	c.helper.RespondOK(ctx, comment, err, "comment updated successfully")
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment (owner only)
// @Tags contest-comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param commentId path int true "Comment ID"
// @Success 204 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/comments/{commentId} [delete]
func (c *CommentController) DeleteComment(ctx *gin.Context) {
	_, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	commentId, err := strconv.ParseInt(ctx.Param("commentId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid comment id"))
		return
	}

	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.service.DeleteComment(commentId, userId)
	c.helper.RespondNoContent(ctx, err)
}
