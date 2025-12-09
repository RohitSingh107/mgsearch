package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Email        string               `bson:"email" json:"email"`
	PasswordHash string               `bson:"password_hash" json:"-"`
	FirstName    string               `bson:"first_name" json:"first_name"`
	LastName     string               `bson:"last_name" json:"last_name"`
	ClientIDs    []primitive.ObjectID `bson:"client_ids" json:"client_ids"`
	IsActive     bool                 `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
}

// ToPublicView returns user data without sensitive information
func (u *User) ToPublicView() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID.Hex(),
		"email":      u.Email,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"client_ids": u.ClientIDs,
		"is_active":  u.IsActive,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}
