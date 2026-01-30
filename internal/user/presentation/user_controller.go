package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/auth/middleware"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/user/application/dto"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	router      *router.Router
	userService *application.UserService
}

func NewUserController(router *router.Router, userService *application.UserService) *UserController {
	return &UserController{
		router:      router,
		userService: userService,
	}
}

func (c *UserController) RegisterRoutes() {
	privateGroup := c.router.ProtectedGroup("/api/users")
	{
		privateGroup.GET("/my", c.GetMyInfo)
		privateGroup.GET("/:id", c.GetUser)
		privateGroup.PUT("/:id", c.UpdateUserInfo)
		privateGroup.PATCH("/:id", c.UpdateUser)
		privateGroup.DELETE("/:id", c.DeleteUser)
	}

	userGroup := c.router.PublicGroup("/api/users")
	{
		userGroup.POST("", c.CreateUser)
	}
}

// GetMyInfo godoc
// @Summary Get my user information
// @Description Get authenticated user's information from access token
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.MyUserResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/users/my [get]
func (c *UserController) GetMyInfo(ctx *gin.Context) {
	userId, ok := middleware.GetUserIdFromContext(ctx)
	if !ok {
		response.JSON(ctx, response.Error(401, "User not authenticated"))
		return
	}

	user, err := c.userService.GetMyInfo(userId)
	if err != nil {
		var businessErr *exception.BusinessError
		if errors.As(err, &businessErr) {
			ctx.JSON(businessErr.Status, businessErr)
			return
		}
		response.JSON(ctx, response.InternalServerError("Internal server error"))
		return
	}

	response.JSON(ctx, response.Success(user, "User information retrieved successfully"))
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/users [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	user, err := c.userService.CreateUser(req)
	if err != nil {
		var businessErr *exception.BusinessError
		if errors.As(err, &businessErr) {
			ctx.JSON(businessErr.Status, businessErr)
			return
		}
		response.JSON(ctx, response.BadRequest("cannot create user"))
		return
	}

	response.JSON(ctx, response.Created(user, "User created successfully"))
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/users/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid user ID"))
		return
	}

	user, err := c.userService.GetUserById(id)
	if err != nil {
		var businessErr *exception.BusinessError
		if errors.As(err, &businessErr) {
			ctx.JSON(businessErr.Status, businessErr)
			return
		}
		response.JSON(ctx, response.InternalServerError("Internal server error"))
		return
	}

	response.JSON(ctx, response.Success(user, "User retrieved successfully"))
}

// UpdateUserInfo godoc
// @Summary Update user information
// @Description Update user profile information (username, tag, bio, avatar)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserInfoRequest true "User info update request"
// @Success 200 {object} response.Response{data=dto.MyUserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /api/users/{id} [put]
func (c *UserController) UpdateUserInfo(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid user ID"))
		return
	}

	var req dto.UpdateUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	user, err := c.userService.UpdateUserInfo(id, req)
	if err != nil {
		var businessErr *exception.BusinessError
		if errors.As(err, &businessErr) {
			ctx.JSON(businessErr.Status, businessErr)
			return
		}
		response.JSON(ctx, response.InternalServerError("Internal server error"))
		return
	}

	response.JSON(ctx, response.Success(user, "User information updated successfully"))
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user password by user ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/users/{id} [patch]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid user ID"))
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	user, err := c.userService.UpdateUser(id, req)
	if err != nil {
		var businessErr *exception.BusinessError
		if errors.As(err, &businessErr) {
			ctx.JSON(businessErr.Status, businessErr)
			return
		}
		response.JSON(ctx, response.InternalServerError("Internal server error"))
		return
	}

	response.JSON(ctx, response.Success(user, "User updated successfully"))
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete user by user ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/users/{id} [delete]
func (c *UserController) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid user ID"))
		return
	}

	if err := c.userService.DeleteUser(id); err != nil {
		var businessErr *exception.BusinessError
		if errors.As(err, &businessErr) {
			ctx.JSON(businessErr.Status, businessErr)
			return
		}
		response.JSON(ctx, response.InternalServerError("Internal server error"))
		return
	}

	response.SendNoContent(ctx)
}
