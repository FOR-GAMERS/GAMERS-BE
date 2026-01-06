package middleware

import (
	"GAMERS-BE/internal/global/response"
	"GAMERS-BE/internal/global/security/jwt/application"
	"GAMERS-BE/internal/global/security/jwt/domain"
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

		claims, err := m.tokenService.Validate(domain.TokenTypeAccess, token)
		if err != nil {
			response.JSON(c, response.Error(401, err.Error()))
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)

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
