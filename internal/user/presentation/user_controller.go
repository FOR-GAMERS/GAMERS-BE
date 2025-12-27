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

type UserController struct {
	userService *application.UserService
}

func NewUserController(userService *application.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) RegisterRoutes(router *gin.Engine) {
	userGroup := router.Group("/api/users")
	{
		userGroup.POST("", c.CreateUser)
		userGroup.GET("/:id", c.GetUser)
		userGroup.PATCH("/:id", c.UpdateUser)
		userGroup.DELETE("/:id", c.DeleteUser)
	}
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
// @Router /users [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSON(ctx, response.BadRequest("Invalid request body"))
		return
	}

	user, err := c.userService.CreateUser(req)
	if err != nil {
		c.handleError(ctx, err)
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
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid user ID"))
		return
	}

	user, err := c.userService.GetUserById(id)
	if err != nil {
		c.handleError(ctx, err)
		return
	}

	response.JSON(ctx, response.Success(user, "User retrieved successfully"))
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user password by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [patch]
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
		c.handleError(ctx, err)
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
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [delete]
func (c *UserController) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.JSON(ctx, response.BadRequest("Invalid user ID"))
		return
	}

	if err := c.userService.DeleteUser(id); err != nil {
		c.handleError(ctx, err)
		return
	}

	response.SendNoContent(ctx)
}

func (c *UserController) handleError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		response.JSON(ctx, response.NotFound(err.Error()))
	case errors.Is(err, domain.ErrUserAlreadyExists):
		response.JSON(ctx, response.Conflict(err.Error()))
	case errors.Is(err, domain.ErrEmailCannotChange):
		response.JSON(ctx, response.BadRequest(err.Error()))
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrPasswordTooShort), errors.Is(err, domain.ErrPasswordTooWeak):
		response.JSON(ctx, response.BadRequest(err.Error()))
	default:
		response.JSON(ctx, response.InternalServerError("Internal server error"))
	}
}
