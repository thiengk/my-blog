package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MealAuth returns a Gin middleware that validates the X-Group-Secret header
// against the configured meal group secret.
func MealAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		provided := c.GetHeader("X-Group-Secret")
		if provided == "" || provided != secret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
