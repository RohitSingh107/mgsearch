package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client represents a client/tenant in the system
type Client struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	UserIDs     []primitive.ObjectID `bson:"user_ids" json:"user_ids"`
	APIKeys     []APIKey             `bson:"api_keys" json:"api_keys"`
	IsActive    bool                 `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

// APIKey represents an API key for client authentication
type APIKey struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key         string             `bson:"key" json:"key"`           // The actual API key (hashed)
	Name        string             `bson:"name" json:"name"`         // Human-readable name
	KeyPrefix   string             `bson:"key_prefix" json:"prefix"` // First few characters for identification
	Permissions []string           `bson:"permissions" json:"permissions"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	LastUsedAt  *time.Time         `bson:"last_used_at,omitempty" json:"last_used_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt   *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// ToPublicView returns client data for public consumption
func (c *Client) ToPublicView() map[string]interface{} {
	apiKeys := make([]map[string]interface{}, len(c.APIKeys))
	for i, key := range c.APIKeys {
		apiKeys[i] = map[string]interface{}{
			"id":           key.ID.Hex(),
			"name":         key.Name,
			"prefix":       key.KeyPrefix,
			"permissions":  key.Permissions,
			"is_active":    key.IsActive,
			"last_used_at": key.LastUsedAt,
			"created_at":   key.CreatedAt,
			"expires_at":   key.ExpiresAt,
		}
	}

	return map[string]interface{}{
		"id":          c.ID.Hex(),
		"name":        c.Name,
		"description": c.Description,
		"user_ids":    c.UserIDs,
		"api_keys":    apiKeys,
		"is_active":   c.IsActive,
		"created_at":  c.CreatedAt,
		"updated_at":  c.UpdatedAt,
	}
}
