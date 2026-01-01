package presentation

import (
	"GAMERS-BE/internal/auth/presentation/middleware"
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/response"
	"GAMERS-BE/internal/user/application"
	"GAMERS-BE/internal/user/application/dto"
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

func (c *UserController) RegisterRoutes(router *gin.Engine, authMiddleware ...*middleware.AuthMiddleware) {
	userGroup := router.Group("/api/users")
	{
		// Public endpoint - no auth required for user creation (signup)
		userGroup.POST("", c.CreateUser)

		// Apply auth middleware to protected endpoints if provided
		if len(authMiddleware) > 0 && authMiddleware[0] != nil {
			protected := userGroup.Group("")
			protected.Use(authMiddleware[0].RequireAuth())
			{
				protected.GET("/:id", c.GetUser)
				protected.PATCH("/:id", c.UpdateUser)
				protected.DELETE("/:id", c.DeleteUser)
			}
		} else {
			// Backward compatibility: if no middleware provided, endpoints are public
			userGroup.GET("/:id", c.GetUser)
			userGroup.PATCH("/:id", c.UpdateUser)
			userGroup.DELETE("/:id", c.DeleteUser)
		}
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
// @Router /users/{id} [get]
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
// @Router /users/{id} [delete]
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
