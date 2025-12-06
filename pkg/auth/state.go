package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type oauthStateClaims struct {
	Shop string `json:"shop"`
	jwt.RegisteredClaims
}

// GenerateStateToken creates a signed JWT used as the OAuth state parameter.
func GenerateStateToken(shop string, signingKey []byte, ttl time.Duration) (string, error) {
	claims := oauthStateClaims{
		Shop: shop,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

// ParseStateToken validates the state token and returns the embedded shop domain.
func ParseStateToken(tokenString string, signingKey []byte) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &oauthStateClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*oauthStateClaims); ok && token.Valid {
		return claims.Shop, nil
	}

	return "", jwt.ErrTokenInvalidClaims
}
