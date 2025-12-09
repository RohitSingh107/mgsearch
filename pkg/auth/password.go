package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a password with its hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
