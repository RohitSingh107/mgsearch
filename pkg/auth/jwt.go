package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	ClientID string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userID, email string, signingKey []byte, duration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

// GenerateJWTWithClient generates a JWT token with client context
func GenerateJWTWithClient(userID, email, clientID string, signingKey []byte, duration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Email:    email,
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

// ParseJWT parses and validates a JWT token
func ParseJWT(tokenString string, signingKey []byte) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
