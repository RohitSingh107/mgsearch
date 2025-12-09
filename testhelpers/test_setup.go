package testhelpers

import (
	"context"
	"fmt"
	"os"
	"time"

	"mgsearch/config"
	"mgsearch/pkg/database"
	"mgsearch/repositories"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestConfig creates a test configuration
func TestConfig() *config.Config {
	return &config.Config{
		MeilisearchURL:      getEnv("TEST_MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchAPIKey:   getEnv("TEST_MEILISEARCH_API_KEY", "test-key"),
		ServerPort:          "8080",
		DatabaseURL:         getEnv("TEST_DATABASE_URL", "mongodb://localhost:27017/mgsearch_test"),
		DatabaseMaxConns:    10,
		ShopifyAPIKey:       "test-shopify-key",
		ShopifyAPISecret:    "test-shopify-secret",
		ShopifyAppURL:       "https://test-app.example.com",
		ShopifyScopes:       "read_products,write_products",
		JWTSigningKey:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		EncryptionKey:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		WebhookSharedSecret: "test-webhook-secret",
		SessionAPIKey:       "test-session-api-key",
	}
}

// SetupTestDatabase creates a test MongoDB database and returns client, database, and cleanup function
func SetupTestDatabase(ctx context.Context, cfg *config.Config) (*mongo.Client, *mongo.Database, func(), error) {
	client, err := database.NewClient(ctx, cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create test database client: %w", err)
	}

	// Extract database name
	dbName := "mgsearch_test"
	if cfg.DatabaseURL != "" {
		// Try to extract from URL - look for last /
		if idx := len(cfg.DatabaseURL) - 1; idx >= 0 {
			for i := len(cfg.DatabaseURL) - 1; i >= 0; i-- {
				if cfg.DatabaseURL[i] == '/' {
					if i < len(cfg.DatabaseURL)-1 {
						dbName = cfg.DatabaseURL[i+1:]
						// Remove query params if any
						for j := 0; j < len(dbName); j++ {
							if dbName[j] == '?' {
								dbName = dbName[:j]
								break
							}
						}
					}
					break
				}
			}
		}
	}

	db := client.Database(dbName)

	// Run migrations
	if err := database.RunMigrations(ctx, client, dbName); err != nil {
		client.Disconnect(ctx)
		return nil, nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	cleanup := func() {
		// Drop test database
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return client, db, cleanup, nil
}

// SetupTestRepositories creates test repositories
func SetupTestRepositories(db *mongo.Database) (*repositories.StoreRepository, *repositories.SessionRepository) {
	storeRepo := repositories.NewStoreRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	return storeRepo, sessionRepo
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CleanupTestDatabase drops all collections in the test database
func CleanupTestDatabase(ctx context.Context, db *mongo.Database) error {
	collections := []string{"stores", "sessions"}
	for _, collName := range collections {
		if err := db.Collection(collName).Drop(ctx); err != nil {
			// Ignore namespace not found errors
			if !isNamespaceNotFound(err) {
				return err
			}
		}
	}
	return nil
}

func isNamespaceNotFound(err error) bool {
	// MongoDB returns specific error codes for namespace not found
	// This is a simple check - in production you'd check the error code
	return err != nil && (err.Error() == "namespace not found" || 
		err.Error() == "ns not found" ||
		err.Error() == "collection not found")
}

// WaitForMongoDB waits for MongoDB to be ready (useful for CI/CD)
func WaitForMongoDB(ctx context.Context, uri string, maxAttempts int) error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	for i := 0; i < maxAttempts; i++ {
		if err := client.Ping(ctx, nil); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("MongoDB not ready after %d attempts", maxAttempts)
}

