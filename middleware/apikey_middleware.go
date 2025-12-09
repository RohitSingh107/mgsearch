package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"mgsearch/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type APIKeyMiddleware struct {
	clientRepo *repositories.ClientRepository
}

func NewAPIKeyMiddleware(clientRepo *repositories.ClientRepository) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		clientRepo: clientRepo,
	}
}

// RequireAPIKey validates API key and sets client context
func (m *APIKeyMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing API key",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Hash the API key for lookup
		apiKeyHash := hashAPIKey(apiKey)

		// Find client by API key
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		client, err := m.clientRepo.FindByAPIKey(ctx, apiKeyHash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Find the specific API key to update last_used_at and check expiration
		var apiKeyID primitive.ObjectID
		var isExpired bool
		for _, key := range client.APIKeys {
			if key.Key == apiKeyHash && key.IsActive {
				apiKeyID = key.ID
				if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now().UTC()) {
					isExpired = true
				}
				break
			}
		}

		if isExpired {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key has expired",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Update last used timestamp (async, don't block request)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = m.clientRepo.UpdateAPIKeyLastUsed(ctx, client.ID, apiKeyID)
		}()

		// Verify client_name in URL matches the client that owns the API key
		clientNameParam := c.Param("client_name")
		if clientNameParam != "" && clientNameParam != client.Name {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "API key does not belong to this client",
				"code":  "FORBIDDEN",
			})
			return
		}

		// Set client information in context
		c.Set(ContextClientIDKey, client.ID.Hex())
		c.Set("client_name", client.Name)

		c.Next()
	}
}

// extractAPIKey extracts API key from Authorization header or X-API-Key header
func extractAPIKey(c *gin.Context) string {
	// Try Authorization header first (Bearer token format)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			return strings.TrimSpace(authHeader[7:])
		}
	}

	// Try X-API-Key header
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	return ""
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
