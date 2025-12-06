package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type SessionClaims struct {
	StoreID string `json:"store_id"`
	Shop    string `json:"shop"`
	jwt.RegisteredClaims
}

func GenerateSessionToken(storeID, shop string, signingKey []byte, ttl time.Duration) (string, error) {
	claims := SessionClaims{
		StoreID: storeID,
		Shop:    shop,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

func ParseSessionToken(tokenString string, signingKey []byte) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SessionClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
