package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MeilisearchURL    string
	MeilisearchAPIKey string
	ServerPort        string
}

// LoadConfig loads configuration from .env file and environment variables
// Environment variables take precedence over .env file values
func LoadConfig() *Config {
	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables only")
	}

	return &Config{
		MeilisearchURL:    getEnv("MEILISEARCH_URL", ""),
		MeilisearchAPIKey: getEnv("MEILISEARCH_API_KEY", ""),
		ServerPort:        getEnv("PORT", "8080"),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
