package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	ContextUserIDKey   = "user_id"
	ContextUserEmail   = "user_email"
	ContextClientIDKey = "client_id"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	ClientID string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

type JWTMiddleware struct {
	signingKey []byte
}

func NewJWTMiddleware(signingKey string) *JWTMiddleware {
	return &JWTMiddleware{
		signingKey: []byte(signingKey),
	}
}

// RequireAuth validates JWT token and sets user context
func (m *JWTMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		tokenString := strings.TrimSpace(authHeader[7:])
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.signingKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Set user information in context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		if claims.ClientID != "" {
			c.Set(ContextClientIDKey, claims.ClientID)
		}

		c.Next()
	}
}

// GetUserID retrieves the user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	value, ok := c.Get(ContextUserIDKey)
	if !ok {
		return "", false
	}
	userID, ok := value.(string)
	return userID, ok
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	value, ok := c.Get(ContextUserEmail)
	if !ok {
		return "", false
	}
	email, ok := value.(string)
	return email, ok
}

// GetClientID retrieves the client ID from context
func GetClientID(c *gin.Context) (string, bool) {
	value, ok := c.Get(ContextClientIDKey)
	if !ok {
		return "", false
	}
	clientID, ok := value.(string)
	return clientID, ok
}
