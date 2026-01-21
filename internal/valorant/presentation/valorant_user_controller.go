package presentation

import (
	"GAMERS-BE/internal/global/common/handler"
	"GAMERS-BE/internal/global/common/router"
	"GAMERS-BE/internal/global/response"
	"GAMERS-BE/internal/valorant/application"
	"GAMERS-BE/internal/valorant/application/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ValorantUserController struct {
	router   *router.Router
	service  *application.ValorantUserService
	helper   *handler.ControllerHelper
}

func NewValorantUserController(
	router *router.Router,
	service *application.ValorantUserService,
) *ValorantUserController {
	return &ValorantUserController{
		router:  router,
		service: service,
		helper:  handler.NewControllerHelper(),
	}
}

func (c *ValorantUserController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/users/valorant")
	{
		privateGroup.POST("", c.RegisterValorant)
		privateGroup.GET("", c.GetValorantInfo)
		privateGroup.POST("/refresh", c.RefreshValorant)
		privateGroup.DELETE("", c.UnlinkValorant)
	}

	contestGroup := c.router.ProtectedGroup("/api/contests")
	{
		contestGroup.GET("/:id/valorant-point", c.GetContestPoint)
	}
}

// RegisterValorant godoc
// @Summary Register Valorant account
// @Description Link a Valorant account to the user by providing Riot ID
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RegisterValorantRequest true "Riot ID registration request"
// @Success 201 {object} response.Response{data=dto.ValorantInfoResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 502 {object} response.Response
// @Router /api/users/valorant [post]
func (c *ValorantUserController) RegisterValorant(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		response.JSON(ctx, response.Unauthorized("User not authenticated"))
		return
	}

	var req dto.RegisterValorantRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	result, err := c.service.RegisterValorant(userId.(int64), &req)
	c.helper.RespondCreated(ctx, result, err, "Valorant account registered successfully")
}

// GetValorantInfo godoc
// @Summary Get Valorant info
// @Description Get the stored Valorant information for the authenticated user
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.ValorantInfoResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/users/valorant [get]
func (c *ValorantUserController) GetValorantInfo(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		response.JSON(ctx, response.Unauthorized("User not authenticated"))
		return
	}

	result, err := c.service.GetValorantInfo(userId.(int64))
	c.helper.RespondOK(ctx, result, err, "Valorant info retrieved successfully")
}

// RefreshValorant godoc
// @Summary Refresh Valorant data
// @Description Refresh the Valorant MMR data from Valorant API
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.ValorantInfoResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 502 {object} response.Response
// @Router /api/users/valorant/refresh [post]
func (c *ValorantUserController) RefreshValorant(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		response.JSON(ctx, response.Unauthorized("User not authenticated"))
		return
	}

	result, err := c.service.RefreshValorant(userId.(int64))
	c.helper.RespondOK(ctx, result, err, "Valorant data refreshed successfully")
}

// UnlinkValorant godoc
// @Summary Unlink Valorant account
// @Description Remove the Valorant account link from the user
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/users/valorant [delete]
func (c *ValorantUserController) UnlinkValorant(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		response.JSON(ctx, response.Unauthorized("User not authenticated"))
		return
	}

	err := c.service.UnlinkValorant(userId.(int64))
	c.helper.RespondNoContent(ctx, err)
}

// GetContestPoint godoc
// @Summary Get contest point
// @Description Calculate and return the contest point for the authenticated user based on Valorant rank
// @Tags valorant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contest ID"
// @Param scoreTableId query int true "Score Table ID"
// @Success 200 {object} response.Response{data=dto.ContestPointResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/contests/{id}/valorant-point [get]
func (c *ValorantUserController) GetContestPoint(ctx *gin.Context) {
	userId, exists := ctx.Get("userId")
	if !exists {
		response.JSON(ctx, response.Unauthorized("User not authenticated"))
		return
	}

	scoreTableIdStr := ctx.Query("scoreTableId")
	if scoreTableIdStr == "" {
		response.JSON(ctx, response.BadRequest("scoreTableId is required"))
		return
	}

	scoreTableId, err := strconv.ParseInt(scoreTableIdStr, 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid scoreTableId"))
		return
	}

	result, err := c.service.CalculateContestPoint(userId.(int64), scoreTableId)
	c.helper.RespondOK(ctx, result, err, "Contest point calculated successfully")
}
