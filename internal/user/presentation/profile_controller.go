package presentation

import (
	"GAMERS-BE/internal/common/response"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
	"GAMERS-BE/internal/user/domain"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	profileService *application.ProfileService
}

func NewProfileController(profileService *application.ProfileService) *ProfileController {
	return &ProfileController{
		profileService: profileService,
	}
}

func (c *ProfileController) RegisterRoutes(router *gin.Engine) {
	profileGroup := router.Group("/api/profiles")
	{
		profileGroup.GET("/:id", c.GetProfile)
		profileGroup.PATCH("/:id", c.UpdateProfile)
		profileGroup.DELETE("/:id", c.DeleteProfile)
	}
}

// GetProfile godoc
// @Summary Get a profile by ID
// @Description Get profile details by profile ID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Success 200 {object} response.Response{data=dto.ProfileResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /profiles/{id} [get]
func (c *ProfileController) GetProfile(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid profile ID"))
		return
	}

	profile, err := c.profileService.GetProfile(id)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	response.JSON(ctx, response.Success(profile, "Profile retrieved successfully"))
}

// UpdateProfile godoc
// @Summary Update a profile
// @Description Update profile information by profile ID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param profile body dto.UpdateProfileRequest true "Profile update request"
// @Success 200 {object} response.Response{data=dto.ProfileResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /profiles/{id} [patch]
func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid profile ID"))
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	profile, err := c.profileService.UpdateProfile(id, req)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	response.JSON(ctx, response.Success(profile, "Profile updated successfully"))
}

// DeleteProfile godoc
// @Summary Delete a profile
// @Description Delete profile by profile ID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /profiles/{id} [delete]
func (c *ProfileController) DeleteProfile(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid profile ID"))
		return
	}

	if err := c.profileService.DeleteProfile(id); err != nil {
		c.handleError(ctx, err)
		return
	}

	response.SendNoContent(ctx)
}

func (c *ProfileController) handleError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrProfileNotFound):
		response.JSON(ctx, response.NotFound(err.Error()))
	case errors.Is(err, domain.ErrProfileAlreadyExists):
		response.JSON(ctx, response.Conflict(err.Error()))
	case errors.Is(err, domain.ErrUsernameEmpty), errors.Is(err, domain.ErrUsernameTooLong), errors.Is(err, domain.ErrUsernameInvalidChar):
		response.JSON(ctx, response.BadRequest(err.Error()))
	case errors.Is(err, domain.ErrTagEmpty), errors.Is(err, domain.ErrTagTooLong), errors.Is(err, domain.ErrTagInvalidChar):
		response.JSON(ctx, response.BadRequest(err.Error()))
	case errors.Is(err, domain.ErrBioTooLong):
		response.JSON(ctx, response.BadRequest(err.Error()))
	default:
		response.JSON(ctx, response.InternalServerError("Internal server error"))
	}
}
