package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OptionalAPIKeyMiddleware validates an optional API key from Authorization header.
// If SESSION_API_KEY is set, it requires the header. If not set, it allows requests through.
func OptionalAPIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no API key is configured, allow all requests
		if apiKey == "" {
			c.Next()
			return
		}

		// If API key is configured, require it
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing authentication token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		token := strings.TrimSpace(authHeader[7:])
		if token != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or missing authentication token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		c.Next()
	}
}

