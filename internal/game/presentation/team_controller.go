package presentation

import (
	"GAMERS-BE/internal/auth/middleware"
	"GAMERS-BE/internal/game/application"
	gameDto "GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TeamController struct {
	router  *router.Router
	service *application.TeamService
	helper  *handler.ControllerHelper
}

func NewTeamController(
	router *router.Router,
	service *application.TeamService,
	helper *handler.ControllerHelper,
) *TeamController {
	return &TeamController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *TeamController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/games/:id/team")
	{
		privateGroup.GET("", c.GetTeam)
		privateGroup.POST("/invite", c.InviteMember)
		privateGroup.POST("/invite/accept", c.AcceptInvite)
		privateGroup.POST("/invite/reject", c.RejectInvite)
		privateGroup.POST("/kick", c.KickMember)
		privateGroup.POST("/leave", c.LeaveTeam)
		privateGroup.POST("/transfer", c.TransferLeadership)
		privateGroup.POST("/finalize", c.FinalizeTeam)
		privateGroup.DELETE("", c.DeleteTeam)
	}

	membersGroup := c.router.ProtectedGroup("/api/games/:id/members")
	{
		membersGroup.GET("", c.GetMembers)
		membersGroup.GET("/:userId", c.GetMember)
	}
}

// GetTeam godoc
// @Summary Get team information
// @Description Get all members and team details for a game (from Redis cache or DB)
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team [get]
func (c *TeamController) GetTeam(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	team, err := c.service.GetTeam(ctx.Request.Context(), gameID)
	c.helper.RespondOK(ctx, team, err, "team retrieved successfully")
}

// InviteMember godoc
// @Summary Invite a user to the team
// @Description Invite a user to join the game team. Sends Discord notification if configured.
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param request body gameDto.InviteMemberRequest true "Invite member request"
// @Success 201 {object} response.Response{data=gameDto.TeamInviteResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/games/{id}/team/invite [post]
func (c *TeamController) InviteMember(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	var req gameDto.InviteMemberRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	invite, err := c.service.InviteMember(ctx.Request.Context(), gameID, userID, req.UserID)
	if err != nil {
		c.helper.RespondCreated(ctx, nil, err, "")
		return
	}

	c.helper.RespondCreated(ctx, gameDto.ToTeamInviteResponse(invite), nil, "member invited successfully")
}

// AcceptInvite godoc
// @Summary Accept a team invitation
// @Description Accept a pending team invitation and join the team
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=gameDto.CachedMemberResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/games/{id}/team/invite/accept [post]
func (c *TeamController) AcceptInvite(ctx *gin.Context) {
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

	member, err := c.service.AcceptInvite(ctx.Request.Context(), gameID, userID)
	if err != nil {
		c.helper.RespondOK(ctx, nil, err, "")
		return
	}

	c.helper.RespondOK(ctx, gameDto.ToCachedMemberResponse(member), nil, "invite accepted successfully")
}

// RejectInvite godoc
// @Summary Reject a team invitation
// @Description Reject a pending team invitation
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team/invite/reject [post]
func (c *TeamController) RejectInvite(ctx *gin.Context) {
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

	err = c.service.RejectInvite(ctx.Request.Context(), gameID, userID)
	c.helper.RespondNoContent(ctx, err)
}

// KickMember godoc
// @Summary Kick a member from the team
// @Description Remove a member from the game team (Leader only)
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param request body gameDto.KickMemberRequest true "Kick member request"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team/kick [post]
func (c *TeamController) KickMember(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	var req gameDto.KickMemberRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.service.KickMember(ctx.Request.Context(), gameID, userID, req.UserID)
	c.helper.RespondNoContent(ctx, err)
}

// LeaveTeam godoc
// @Summary Leave the team
// @Description Leave the game team voluntarily (Members only, Leader cannot leave)
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team/leave [post]
func (c *TeamController) LeaveTeam(ctx *gin.Context) {
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

	err = c.service.LeaveTeam(ctx.Request.Context(), gameID, userID)
	c.helper.RespondNoContent(ctx, err)
}

// TransferLeadership godoc
// @Summary Transfer leadership to another member
// @Description Transfer team leadership to another member (Leader only)
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param request body gameDto.TransferLeadershipRequest true "Transfer leadership request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team/transfer [post]
func (c *TeamController) TransferLeadership(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	var req gameDto.TransferLeadershipRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	userID, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	err = c.service.TransferLeadership(ctx.Request.Context(), gameID, userID, req.NewLeaderUserID)
	c.helper.RespondOK(ctx, nil, err, "leadership transferred successfully")
}

// FinalizeTeam godoc
// @Summary Finalize the team
// @Description Finalize the team and persist to database (Leader only, team must be full)
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team/finalize [post]
func (c *TeamController) FinalizeTeam(ctx *gin.Context) {
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

	err = c.service.FinalizeTeam(ctx.Request.Context(), gameID, userID)
	c.helper.RespondOK(ctx, nil, err, "team finalized successfully")
}

// DeleteTeam godoc
// @Summary Delete the entire team
// @Description Delete all team members (Leader only)
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/team [delete]
func (c *TeamController) DeleteTeam(ctx *gin.Context) {
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

	err = c.service.DeleteTeam(ctx.Request.Context(), gameID, userID)
	c.helper.RespondNoContent(ctx, err)
}

// GetMembers godoc
// @Summary Get all team members
// @Description Get all members of a game team
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} response.Response{data=[]gameDto.CachedMemberResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/members [get]
func (c *TeamController) GetMembers(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	members, err := c.service.GetMembers(ctx.Request.Context(), gameID)
	if err != nil {
		c.helper.RespondOK(ctx, nil, err, "")
		return
	}

	c.helper.RespondOK(ctx, gameDto.ToCachedMemberResponses(members), nil, "members retrieved successfully")
}

// GetMember godoc
// @Summary Get a specific team member
// @Description Get a specific member of a game team by user ID
// @Tags teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param userId path int true "User ID"
// @Success 200 {object} response.Response{data=gameDto.CachedMemberResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/games/{id}/members/{userId} [get]
func (c *TeamController) GetMember(ctx *gin.Context) {
	gameID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid game id"))
		return
	}

	userID, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid user id"))
		return
	}

	member, err := c.service.GetMember(ctx.Request.Context(), gameID, userID)
	if err != nil {
		c.helper.RespondOK(ctx, nil, err, "")
		return
	}

	c.helper.RespondOK(ctx, gameDto.ToCachedMemberResponse(member), nil, "member retrieved successfully")
}
