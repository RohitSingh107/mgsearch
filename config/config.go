package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MeilisearchURL      string
	MeilisearchAPIKey   string
	ServerPort          string
	DatabaseURL         string
	DatabaseMaxConns    int32
	ShopifyAPIKey       string
	ShopifyAPISecret    string
	ShopifyAppURL       string
	ShopifyScopes       string
	JWTSigningKey       string
	EncryptionKey       string
	WebhookSharedSecret string
	SessionAPIKey       string // Optional API key for session endpoints
	QdrantURL           string
	QdrantAPIKey        string
}

// LoadConfig loads configuration from .env file and environment variables.
// Environment variables take precedence over .env file values.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables only")
	}

	return &Config{
		MeilisearchURL:      getEnv("MEILISEARCH_URL", ""),
		MeilisearchAPIKey:   getEnv("MEILISEARCH_API_KEY", ""),
		ServerPort:          getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		DatabaseMaxConns:    getEnvAsInt32("DATABASE_MAX_CONNS", 10),
		ShopifyAPIKey:       getEnv("SHOPIFY_API_KEY", ""),
		ShopifyAPISecret:    getEnv("SHOPIFY_API_SECRET", ""),
		ShopifyAppURL:       getEnv("SHOPIFY_APP_URL", ""),
		ShopifyScopes:       getEnv("SHOPIFY_SCOPES", "read_products,write_products,read_product_listings,read_collection_listings,read_inventory,write_webhooks"),
		JWTSigningKey:       getEnv("JWT_SIGNING_KEY", ""),
		EncryptionKey:       getEnv("ENCRYPTION_KEY", ""),
		WebhookSharedSecret: getEnv("SHOPIFY_WEBHOOK_SECRET", ""),
		SessionAPIKey:       getEnv("SESSION_API_KEY", ""), // Optional
		QdrantURL:           getEnv("QDRANT_URL", ""),
		QdrantAPIKey:        getEnv("QDRANT_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt32(key string, defaultValue int32) int32 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return int32(intValue)
		}
	}
	return defaultValue
}
