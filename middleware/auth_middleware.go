package middleware

import (
	"net/http"
	"strings"

	"mgsearch/pkg/auth"

	"github.com/gin-gonic/gin"
)

const (
	contextStoreIDKey = "store_id"
	contextShopKey    = "shop_domain"
)

type AuthMiddleware struct {
	signingKey []byte
}

func NewAuthMiddleware(signingKey string) *AuthMiddleware {
	return &AuthMiddleware{
		signingKey: []byte(signingKey),
	}
}

func (m *AuthMiddleware) RequireStoreSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		token := strings.TrimSpace(authHeader[7:])
		claims, err := auth.ParseSessionToken(token, m.signingKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(contextStoreIDKey, claims.StoreID)
		c.Set(contextShopKey, claims.Shop)
		c.Next()
	}
}

func GetStoreID(c *gin.Context) (string, bool) {
	value, ok := c.Get(contextStoreIDKey)
	if !ok {
		return "", false
	}
	storeID, ok := value.(string)
	return storeID, ok
}
