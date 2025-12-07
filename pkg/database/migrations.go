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

	return nil
}
