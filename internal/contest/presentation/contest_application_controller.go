package presentation

import (
	"GAMERS-BE/internal/auth/middleware"
	"GAMERS-BE/internal/contest/application"
	"GAMERS-BE/internal/contest/application/dto"
	commonDto "GAMERS-BE/internal/global/common/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/response"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContestApplicationController struct {
	router  *router.Router
	service *application.ContestApplicationService
	helper  *handler.ControllerHelper
}

func NewContestApplicationController(
	router *router.Router,
	service *application.ContestApplicationService,
	helper *handler.ControllerHelper,
) *ContestApplicationController {
	return &ContestApplicationController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *ContestApplicationController) RegisterRoute() {
	privateGroup := c.router.ProtectedGroup("/api/contests/:id/applications")
	privateGroup.POST("", c.RequestParticipate)
	privateGroup.GET("", c.GetPendingApplications)
	privateGroup.GET("/me", c.GetMyApplication)
	privateGroup.POST("/:userId/accept", c.AcceptApplication)
	privateGroup.POST("/:userId/reject", c.RejectApplication)
	privateGroup.DELETE("/cancel", c.CancelApplication)
	privateGroup.DELETE("/withdraw", c.WithdrawFromContest)

	// Contest members endpoints
	membersGroup := c.router.ProtectedGroup("/api/contests/:id/members")
	membersGroup.GET("", c.GetContestMembers)
	membersGroup.PATCH("/:userId/role", c.ChangeMemberRole)

	// Contest status endpoint
	statusGroup := c.router.ProtectedGroup("/api/contests/:id/status")
	statusGroup.GET("/me", c.GetMyContestStatus)
}

// RequestParticipate godoc
// @Summary Request to participate in a contest
// @Description Apply to participate in a contest
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response{data=dto.DiscordLinkRequiredResponse}
// @Failure 409 {object} response.Response
// @Router /api/contests/{contestId}/applications [post]
func (c *ContestApplicationController) RequestParticipate(ctx *gin.Context) {
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

	discordLinkRequired, err := c.service.RequestParticipate(ctx.Request.Context(), contestId, userId)

	// Handle Discord link required error
	if errors.Is(err, exception.ErrDiscordLinkRequired) {
		response.JSON(ctx, response.Forbidden(discordLinkRequired, "discord linking required"))
		return
	}

	c.helper.RespondCreated(ctx, nil, err, "application submitted successfully")
}

// GetPendingApplications godoc
// @Summary Get pending applications for a contest
// @Description Get all pending applications for a contest (Leader only)
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/contests/{contestId}/applications [get]
func (c *ContestApplicationController) GetPendingApplications(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	applications, err := c.service.GetPendingApplications(ctx.Request.Context(), contestId)
	c.helper.RespondOK(ctx, applications, err, "applications retrieved successfully")
}

// GetMyApplication godoc
// @Summary Get my application status
// @Description Get current user's application status for a contest
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/applications/me [get]
func (c *ContestApplicationController) GetMyApplication(ctx *gin.Context) {
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

	myApplication, err := c.service.GetMyApplication(ctx.Request.Context(), contestId, userId)
	c.helper.RespondOK(ctx, myApplication, err, "application retrieved successfully")
}

// AcceptApplication godoc
// @Summary Accept a contest application
// @Description Accept a user's application to join the contest (Leader only)
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param userId path int true "User ID to accept"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/contests/{contestId}/applications/{userId}/accept [post]
func (c *ContestApplicationController) AcceptApplication(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid user id"))
		return
	}

	leaderUserId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.service.AcceptApplication(ctx.Request.Context(), contestId, userId, leaderUserId)
	c.helper.RespondOK(ctx, nil, err, "application accepted successfully")
}

// RejectApplication godoc
// @Summary Reject a contest application
// @Description Reject a user's application to join the contest (Leader only)
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param userId path int true "User ID to reject"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/contests/{contestId}/applications/{userId}/reject [post]
func (c *ContestApplicationController) RejectApplication(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid user id"))
		return
	}

	leaderUserId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.service.RejectApplication(ctx.Request.Context(), contestId, userId, leaderUserId)
	c.helper.RespondOK(ctx, nil, err, "application rejected successfully")
}

// CancelApplication godoc
// @Summary Cancel my pending application
// @Description Cancel current user's pending application for a contest
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/applications/cancel [delete]
func (c *ContestApplicationController) CancelApplication(ctx *gin.Context) {
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

	err = c.service.CancelApplication(ctx.Request.Context(), contestId, userId)
	c.helper.RespondOK(ctx, nil, err, "application cancelled successfully")
}

// WithdrawFromContest godoc
// @Summary Withdraw from a contest
// @Description Withdraw current user from a contest (Leader cannot withdraw)
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/applications/withdraw [delete]
func (c *ContestApplicationController) WithdrawFromContest(ctx *gin.Context) {
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

	err = c.service.WithdrawFromContest(contestId, userId)
	c.helper.RespondOK(ctx, nil, err, "successfully withdrawn from contest")
}

// ChangeMemberRole godoc
// @Summary Change a member's role (Leader only)
// @Description Promote a member to STAFF or demote STAFF to NORMAL (Leader only, cannot change leader's role)
// @Tags contest-members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param userId path int true "Target User ID"
// @Param request body dto.ChangeMemberRoleRequest true "New member role"
// @Success 200 {object} response.Response{data=dto.ChangeMemberRoleResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/members/{userId}/role [patch]
func (c *ContestApplicationController) ChangeMemberRole(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	targetUserId, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid user id"))
		return
	}

	leaderUserId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	var req dto.ChangeMemberRoleRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	// Validate member type
	if !req.MemberType.IsValid() {
		response.JSON(ctx, response.BadRequest("invalid member type, must be STAFF or NORMAL"))
		return
	}

	result, err := c.service.ChangeMemberRole(contestId, targetUserId, leaderUserId, req.MemberType)
	c.helper.RespondOK(ctx, result, err, "member role changed successfully")
}

// GetMyContestStatus godoc
// @Summary Get my status in a contest
// @Description Get current user's status in a contest (is_leader, is_member, has_applied, application_status)
// @Tags contest-applications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Success 200 {object} response.Response{data=dto.UserContestStatusResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/status/me [get]
func (c *ContestApplicationController) GetMyContestStatus(ctx *gin.Context) {
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

	status, err := c.service.GetMyContestStatus(ctx.Request.Context(), contestId, userId)
	c.helper.RespondOK(ctx, status, err, "contest status retrieved successfully")
}

// GetContestMembers godoc
// @Summary Get contest members with pagination
// @Description Get all members of a contest with user information and points, sorted by points
// @Tags contest-members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contestId path int true "Contest ID"
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param page_size query int false "Page size (default: 10, max: 100)" minimum(1) maximum(100)
// @Param sort_by query string false "Sort field (point, username)" default(point)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} response.Response{data=commonDto.PaginationResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{contestId}/members [get]
func (c *ContestApplicationController) GetContestMembers(ctx *gin.Context) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pagination := commonDto.NewPaginationRequest(page, pageSize)

	// Parse sort parameters
	sortBy := ctx.DefaultQuery("sort_by", "point")
	order := ctx.DefaultQuery("order", "desc")
	sort := commonDto.NewSortRequest(sortBy, order, []string{"point", "username"})

	result, err := c.service.GetContestMembers(ctx.Request.Context(), contestId, pagination, sort)
	c.helper.RespondOK(ctx, result, err, "members retrieved successfully")
}

// ==================== Testable Handler Functions ====================
// These functions are exposed for unit testing without router dependency

func HandleGetPendingApplications(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	applications, err := service.GetPendingApplications(ctx.Request.Context(), contestId)
	helper.RespondOK(ctx, applications, err, "applications retrieved successfully")
}

func HandleGetMyApplication(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
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

	myApplication, err := service.GetMyApplication(ctx.Request.Context(), contestId, userId)
	helper.RespondOK(ctx, myApplication, err, "application retrieved successfully")
}

func HandleAcceptApplication(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid user id"))
		return
	}

	leaderUserId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = service.AcceptApplication(ctx.Request.Context(), contestId, userId, leaderUserId)
	helper.RespondOK(ctx, nil, err, "application accepted successfully")
}

func HandleRejectApplication(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	userId, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid user id"))
		return
	}

	leaderUserId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = service.RejectApplication(ctx.Request.Context(), contestId, userId, leaderUserId)
	helper.RespondOK(ctx, nil, err, "application rejected successfully")
}

func HandleCancelApplication(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
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

	err = service.CancelApplication(ctx.Request.Context(), contestId, userId)
	helper.RespondOK(ctx, nil, err, "application cancelled successfully")
}

func HandleGetMyContestStatus(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
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

	status, err := service.GetMyContestStatus(ctx.Request.Context(), contestId, userId)
	helper.RespondOK(ctx, status, err, "contest status retrieved successfully")
}

func HandleGetContestMembers(ctx *gin.Context, service *application.ContestApplicationService, helper *handler.ControllerHelper) {
	contestId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest id"))
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	pagination := commonDto.NewPaginationRequest(page, pageSize)

	sortBy := ctx.DefaultQuery("sort_by", "point")
	order := ctx.DefaultQuery("order", "desc")
	sort := commonDto.NewSortRequest(sortBy, order, []string{"point", "username"})

	result, err := service.GetContestMembers(ctx.Request.Context(), contestId, pagination, sort)
	helper.RespondOK(ctx, result, err, "members retrieved successfully")
}
