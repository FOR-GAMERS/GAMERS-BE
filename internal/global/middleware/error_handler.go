package middleware

import (
	"GAMERS-BE/internal/global/exception"
	"errors"

	"github.com/gin-gonic/gin"
)

func GlobalErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var businessErr *exception.BusinessError
			if errors.As(err, &businessErr) {
				c.AbortWithStatusJSON(businessErr.Status, businessErr)
				return
			}

			c.AbortWithStatusJSON(500, gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal Server Error is occurred",
			})
		}
	}
}
