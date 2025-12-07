package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"mgsearch/config"
	"mgsearch/handlers"
	"mgsearch/middleware"
	"mgsearch/pkg/database"
	"mgsearch/repositories"
	"mgsearch/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	validateConfig(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := database.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("failed to disconnect from database: %v", err)
		}
	}()

	if err := database.Ping(ctx, client); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}

	// Extract database name from connection string or use default
	dbName := "mgsearch"
	if cfg.DatabaseURL != "" {
		// Try to extract database name from MongoDB URI
		// Format: mongodb://host:port/dbname
		if idx := strings.LastIndex(cfg.DatabaseURL, "/"); idx != -1 && idx < len(cfg.DatabaseURL)-1 {
			if queryIdx := strings.Index(cfg.DatabaseURL[idx+1:], "?"); queryIdx != -1 {
				dbName = cfg.DatabaseURL[idx+1 : idx+1+queryIdx]
			} else {
				dbName = cfg.DatabaseURL[idx+1:]
			}
		}
	}

	if err := database.RunMigrations(ctx, client, dbName); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	db := database.GetDatabase(client, dbName)
	storeRepo := repositories.NewStoreRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	meiliService := services.NewMeilisearchService(cfg)
	shopifyService := services.NewShopifyService(cfg)

	authHandler, err := handlers.NewAuthHandler(cfg, shopifyService, storeRepo, meiliService)
	if err != nil {
		log.Fatalf("failed to initialize auth handler: %v", err)
	}
	storeHandler := handlers.NewStoreHandler(storeRepo)
	sessionHandler, err := handlers.NewSessionHandler(sessionRepo, storeRepo, meiliService, cfg)
	if err != nil {
		log.Fatalf("failed to initialize session handler: %v", err)
	}
	webhookHandler := handlers.NewWebhookHandler(shopifyService, storeRepo, meiliService)
	storefrontHandler := handlers.NewStorefrontHandler(storeRepo, meiliService)
	searchHandler := handlers.NewSearchHandler(meiliService)
	settingsHandler := handlers.NewSettingsHandler(meiliService)
	tasksHandler := handlers.NewTasksHandler(meiliService)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSigningKey)

	router := gin.Default()

	// Add CORS middleware for storefront requests
	router.Use(middleware.CORSMiddleware())

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	{
		shopifyGroup := api.Group("/auth/shopify")
		{
			shopifyGroup.POST("/begin", authHandler.Begin)
			shopifyGroup.GET("/callback", authHandler.Callback)
			shopifyGroup.POST("/exchange", authHandler.ExchangeToken)
			shopifyGroup.POST("/install", authHandler.InstallStore)
		}

		storeGroup := api.Group("/stores")
		storeGroup.Use(authMiddleware.RequireStoreSession())
		{
			storeGroup.GET("/current", storeHandler.GetCurrentStore)
			storeGroup.GET("/sync-status", storeHandler.GetSyncStatus)
		}

		sessionGroup := api.Group("/sessions")
		sessionGroup.Use(middleware.OptionalAPIKeyMiddleware(cfg.SessionAPIKey))
		{
			sessionGroup.POST("", sessionHandler.StoreSession)
			sessionGroup.GET("/:id", sessionHandler.LoadSession)
			sessionGroup.DELETE("/:id", sessionHandler.DeleteSession)
			sessionGroup.DELETE("/batch", sessionHandler.DeleteMultipleSessions)
			sessionGroup.GET("/shop/:shop", sessionHandler.FindSessionsByShop)
		}
	}

	router.POST("/webhooks/shopify/:topic/:subtopic", webhookHandler.HandleShopifyWebhook)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/search", storefrontHandler.Search)
		v1.POST("/search", storefrontHandler.Search) // Support POST for JSON body with filters
		v1.POST("/clients/:client_name/:index_name/search", searchHandler.Search)
		v1.POST("/clients/:client_name/:index_name/documents", searchHandler.IndexDocument)

		// Settings endpoint
		// PATCH /api/v1/clients/:client_name/:index_name/settings
		// Example: PATCH /api/v1/clients/myclient/movies/settings
		// Body: { "rankingRules": [...], "searchableAttributes": [...], ... }
		v1.PATCH("/clients/:client_name/:index_name/settings", settingsHandler.UpdateSettings)

		// Tasks endpoint
		// GET /api/v1/clients/:client_name/tasks/:task_id
		// Example: GET /api/v1/clients/myclient/tasks/15
		v1.GET("/clients/:client_name/tasks/:task_id", tasksHandler.GetTask)
	}

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func validateConfig(cfg *config.Config) {
	if cfg.MeilisearchURL == "" {
		log.Fatal("MEILISEARCH_URL is required")
	}
	if cfg.MeilisearchAPIKey == "" {
		log.Fatal("MEILISEARCH_API_KEY is required")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.ShopifyAPIKey == "" || cfg.ShopifyAPISecret == "" || cfg.ShopifyAppURL == "" {
		log.Fatal("SHOPIFY_API_KEY, SHOPIFY_API_SECRET, and SHOPIFY_APP_URL are required")
	}
	if cfg.JWTSigningKey == "" {
		log.Fatal("JWT_SIGNING_KEY is required")
	}
	if cfg.EncryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY is required (32-byte hex string)")
	}
}
