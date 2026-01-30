package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/middleware"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/storage/application/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

// swagger type alias
var _ = dto.UploadResponse{}

type StorageController struct {
	router  *router.Router
	service *application.StorageService
	helper  *handler.ControllerHelper
}

func NewStorageController(router *router.Router, service *application.StorageService, helper *handler.ControllerHelper) *StorageController {
	return &StorageController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *StorageController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/v1/storage")
	privateGroup.POST("/contest-banner", c.UploadContestBanner)
	privateGroup.POST("/user-profile", c.UploadUserProfile)
}

// UploadContestBanner godoc
// @Summary Upload a contest banner image
// @Description Upload a banner image for a contest. Maximum file size is 5MB. Allowed formats: jpeg, png, webp.
// @Tags storage
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param contest_id formData int true "Contest ID"
// @Param file formData file true "Image file (max 5MB, jpeg/png/webp)"
// @Success 201 {object} response.Response{data=dto.UploadResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/storage/contest-banner [post]
func (c *StorageController) UploadContestBanner(ctx *gin.Context) {
	_, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	contestIdStr := ctx.PostForm("contest_id")
	if contestIdStr == "" {
		response.JSON(ctx, response.BadRequest("contest_id is required"))
		return
	}

	contestId, err := strconv.ParseInt(contestIdStr, 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid contest_id"))
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.JSON(ctx, response.BadRequest("file is required"))
		return
	}

	result, err := c.service.UploadContestBanner(ctx.Request.Context(), contestId, file)
	c.helper.RespondCreated(ctx, result, err, "uploaded successfully")
}

// UploadUserProfile godoc
// @Summary Upload a user profile image
// @Description Upload a profile image for the authenticated user. Maximum file size is 2MB. Allowed formats: jpeg, png, webp.
// @Tags storage
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file (max 2MB, jpeg/png/webp)"
// @Success 201 {object} response.Response{data=dto.UploadResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/storage/user-profile [post]
func (c *StorageController) UploadUserProfile(ctx *gin.Context) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "user not authenticated"))
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.JSON(ctx, response.BadRequest("file is required"))
		return
	}

	result, err := c.service.UploadUserProfile(ctx.Request.Context(), userId, file)
	c.helper.RespondCreated(ctx, result, err, "uploaded successfully")
}
