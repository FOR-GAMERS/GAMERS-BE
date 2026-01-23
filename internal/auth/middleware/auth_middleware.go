package middleware

import (
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/response"
	"GAMERS-BE/internal/global/security/jwt/application"
	jwtdomain "GAMERS-BE/internal/global/security/jwt/domain"
	userdomain "GAMERS-BE/internal/user/domain"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserQueryPort is the interface for querying users (to avoid circular dependency)
type UserQueryPort interface {
	FindById(id int64) (*userdomain.User, error)
}

type AuthMiddleware struct {
	tokenService  *application.TokenService
	userQueryPort UserQueryPort
}

func NewAuthMiddleware(tokenService *application.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// SetUserQueryPort sets the user query port for admin middleware
func (m *AuthMiddleware) SetUserQueryPort(port UserQueryPort) {
	m.userQueryPort = port
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.JSON(c, response.Error(401, "Authorization header required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.JSON(c, response.Error(401, "Invalid authorization format"))
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := m.tokenService.Validate(jwtdomain.TokenTypeAccess, token)
		if err != nil {
			response.JSON(c, response.Error(401, err.Error()))
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)

		c.Next()
	}
}

// RequireAdmin is a middleware that checks if the user has admin role
// Must be used after RequireAuth middleware
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.userQueryPort == nil {
			response.JSON(c, response.InternalServerError("admin middleware not configured"))
			c.Abort()
			return
		}

		userId, ok := GetUserIdFromContext(c)
		if !ok {
			response.JSON(c, response.Error(401, "user not authenticated"))
			c.Abort()
			return
		}

		user, err := m.userQueryPort.FindById(userId)
		if err != nil {
			response.JSON(c, response.Error(401, "user not found"))
			c.Abort()
			return
		}

		if !user.IsAdmin() {
			c.JSON(exception.ErrAdminRequired.Status, exception.ErrAdminRequired)
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetUserIdFromContext(c *gin.Context) (int64, bool) {
	userId, exists := c.Get("userId")
	if !exists {
		return 0, false
	}

	userIdInt, ok := userId.(int64)
	return userIdInt, ok
}

func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("userEmail")
	if !exists {
		return "", false
	}

	emailStr, ok := email.(string)
	return emailStr, ok
}
