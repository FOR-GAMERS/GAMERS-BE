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

	response.JSON(ctx, response.NoContent("User deleted successfully"))
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
