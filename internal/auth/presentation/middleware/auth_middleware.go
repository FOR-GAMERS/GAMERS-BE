package middleware

import (
	"GAMERS-BE/internal/auth/infra/jwt"
	"GAMERS-BE/internal/global/response"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenManager *jwt.TokenManager
}

func NewAuthMiddleware(tokenManager *jwt.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenManager: tokenManager,
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

		claims, err := m.tokenManager.ValidateAccessToken(token)
		if err != nil {
			response.JSON(c, response.Error(401, "Invalid or expired token"))
			c.Abort()
			return
		}

		c.Set("userId", claims.UserId)

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
