package main

import (
	"log"
	"net/http"

	"mgsearch/config"
	"mgsearch/handlers"
	"mgsearch/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate required configuration
	if cfg.MeilisearchURL == "" {
		log.Fatal("MEILISEARCH_URL environment variable is required")
	}
	if cfg.MeilisearchAPIKey == "" {
		log.Fatal("MEILISEARCH_API_KEY environment variable is required")
	}

	// Initialize services
	meilisearchService := services.NewMeilisearchService(cfg)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(meilisearchService)

	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Search endpoint
		// POST /api/v1/clients/:client_name/:index_name/search
		// Example: POST /api/v1/clients/myclient/test_index/search
		// Body: { "q": "search query" }
		v1.POST("/clients/:client_name/:index_name/search", searchHandler.Search)
	}

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
