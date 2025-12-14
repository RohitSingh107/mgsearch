package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Index represents a Meilisearch index belonging to a client
type Index struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ClientID   primitive.ObjectID `bson:"client_id" json:"client_id"`
	Name       string             `bson:"name" json:"name"` // User friendly name (e.g. "movies")
	UID        string             `bson:"uid" json:"uid"`   // Meilisearch UID (e.g. "client_name__movies")
	PrimaryKey string             `bson:"primary_key,omitempty" json:"primary_key,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateIndexRequest represents the request body for creating an index
type CreateIndexRequest struct {
	Name       string `json:"name" binding:"required"`
	PrimaryKey string `json:"primary_key,omitempty"`
}
