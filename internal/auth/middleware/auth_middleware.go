package middleware

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/response"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/application"
	jwtdomain "github.com/FOR-GAMERS/GAMERS-BE/internal/global/security/jwt/domain"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenService *application.TokenService
}

func NewAuthMiddleware(tokenService *application.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
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
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RequireAdmin is a middleware that checks if the user has admin role
// Must be used after RequireAuth middleware
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRoleFromContext(c)
		if !ok {
			response.JSON(c, response.Error(401, "user not authenticated"))
			c.Abort()
			return
		}

		if role != "ADMIN" {
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

func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("userRole")
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}

func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("userEmail")
	if !exists {
		return "", false
	}

	emailStr, ok := email.(string)
	return emailStr, ok
}
