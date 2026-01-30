package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/banner/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/banner/application/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/handler"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// swagger type alias
var _ = dto.BannerResponse{}
var _ = dto.BannerListResponse{}

type BannerController struct {
	router  *router.Router
	service *application.BannerService
	helper  *handler.ControllerHelper
}

func NewBannerController(router *router.Router, service *application.BannerService, helper *handler.ControllerHelper) *BannerController {
	return &BannerController{
		router:  router,
		service: service,
		helper:  helper,
	}
}

func (c *BannerController) RegisterRoutes() {
	// Public routes
	publicGroup := c.router.PublicGroup("/api/banners")
	publicGroup.GET("", c.GetActiveBanners)

	// Admin routes
	adminGroup := c.router.AdminGroup("/api/admin/banners")
	adminGroup.GET("", c.GetAllBanners)
	adminGroup.POST("", c.CreateBanner)
	adminGroup.PATCH("/:id", c.UpdateBanner)
	adminGroup.DELETE("/:id", c.DeleteBanner)
}

// GetActiveBanners godoc
// @Summary Get active banners
// @Description Retrieves all active banners ordered by display_order (for homepage display)
// @Tags banners
// @Produce json
// @Success 200 {object} response.Response{data=dto.BannerListResponse}
// @Router /api/banners [get]
func (c *BannerController) GetActiveBanners(ctx *gin.Context) {
	result, err := c.service.GetActiveBanners()
	c.helper.RespondOK(ctx, result, err, "active banners retrieved successfully")
}

// GetAllBanners godoc
// @Summary Get all banners (Admin)
// @Description Retrieves all banners including inactive ones (Admin only)
// @Tags banners
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.BannerListResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /api/admin/banners [get]
func (c *BannerController) GetAllBanners(ctx *gin.Context) {
	result, err := c.service.GetAllBanners()
	c.helper.RespondOK(ctx, result, err, "all banners retrieved successfully")
}

// CreateBanner godoc
// @Summary Create a new banner (Admin)
// @Description Creates a new banner (maximum 5 banners allowed, Admin only)
// @Tags banners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBannerRequest true "Banner creation request"
// @Success 201 {object} response.Response{data=dto.BannerResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 409 {object} response.Response "Maximum banner limit exceeded"
// @Router /api/admin/banners [post]
func (c *BannerController) CreateBanner(ctx *gin.Context) {
	var req dto.CreateBannerRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	result, err := c.service.CreateBanner(&req)
	c.helper.RespondCreated(ctx, result, err, "banner created successfully")
}

// UpdateBanner godoc
// @Summary Update a banner (Admin)
// @Description Updates an existing banner (Admin only)
// @Tags banners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Banner ID"
// @Param request body dto.UpdateBannerRequest true "Banner update request"
// @Success 200 {object} response.Response{data=dto.BannerResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/admin/banners/{id} [patch]
func (c *BannerController) UpdateBanner(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid banner id"))
		return
	}

	var req dto.UpdateBannerRequest
	if !c.helper.BindJSON(ctx, &req) {
		return
	}

	result, err := c.service.UpdateBanner(id, &req)
	c.helper.RespondOK(ctx, result, err, "banner updated successfully")
}

// DeleteBanner godoc
// @Summary Delete a banner (Admin)
// @Description Deletes an existing banner (Admin only)
// @Tags banners
// @Produce json
// @Security BearerAuth
// @Param id path int true "Banner ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/admin/banners/{id} [delete]
func (c *BannerController) DeleteBanner(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("invalid banner id"))
		return
	}

	err = c.service.DeleteBanner(id)
	c.helper.RespondNoContent(ctx, err)
}
