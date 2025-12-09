package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RunMigrations creates MongoDB collections and indexes required for the service.
// For production, prefer using a dedicated migration tool, but this ensures
// local development works out-of-the-box.
func RunMigrations(ctx context.Context, client *mongo.Client, dbName string) error {
	db := client.Database(dbName)

	// Create stores collection and indexes
	storesCollection := db.Collection("stores")

	// Create unique indexes for stores
	storeIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"shop_domain": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    map[string]interface{}{"api_key_public": 1},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := storesCollection.Indexes().CreateMany(ctx, storeIndexes); err != nil {
		return fmt.Errorf("failed to create store indexes: %w", err)
	}

	// Create sessions collection and indexes
	sessionsCollection := db.Collection("sessions")

	// Create indexes for sessions
	sessionIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"shop": 1},
		},
		{
			Keys: map[string]interface{}{"expires": 1},
		},
	}

	if _, err := sessionsCollection.Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("failed to create session indexes: %w", err)
	}

	// Create users collection and indexes
	usersCollection := db.Collection("users")

	// Create unique index on email
	userIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"client_ids": 1},
		},
		{
			Keys: map[string]interface{}{"is_active": 1},
		},
	}

	if _, err := usersCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	// Create clients collection and indexes
	clientsCollection := db.Collection("clients")

	// Create unique index on name and indexes for API key lookups
	clientIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"user_ids": 1},
		},
		{
			Keys: map[string]interface{}{"is_active": 1},
		},
		{
			Keys: map[string]interface{}{"api_keys.key": 1},
		},
	}

	if _, err := clientsCollection.Indexes().CreateMany(ctx, clientIndexes); err != nil {
		return fmt.Errorf("failed to create client indexes: %w", err)
	}

	return nil
}
