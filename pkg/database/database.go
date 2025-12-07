package database

import (
	"context"
	"fmt"
	"time"

	"mgsearch/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewClient initializes a MongoDB client using application configuration.
func NewClient(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	clientOptions := options.Client().ApplyURI(cfg.DatabaseURL)

	if cfg.DatabaseMaxConns > 0 {
		maxPoolSize := uint64(cfg.DatabaseMaxConns)
		clientOptions.SetMaxPoolSize(maxPoolSize)
	}

	clientOptions.SetMaxConnIdleTime(5 * time.Minute)
	clientOptions.SetMaxConnecting(10)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	return client, nil
}

// Ping ensures the database connection is healthy.
func Ping(ctx context.Context, client *mongo.Client) error {
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// GetDatabase returns the database instance from the client.
func GetDatabase(client *mongo.Client, dbName string) *mongo.Database {
	return client.Database(dbName)
}
