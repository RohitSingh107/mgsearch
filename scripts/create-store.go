package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"time"

	"mgsearch/config"
	"mgsearch/models"
	"mgsearch/pkg/database"
	"mgsearch/pkg/security"
	"mgsearch/repositories"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.DatabaseURL == "" {
		fmt.Fprintf(os.Stderr, "Error: DATABASE_URL not set\n")
		os.Exit(1)
	}

	shopDomain := "mg-store-207095.myshopify.com"
	if len(os.Args) > 1 {
		shopDomain = os.Args[1]
	}

	shopName := "Mg Store"
	if len(os.Args) > 2 {
		shopName = os.Args[2]
	}

	// Normalize shop domain
	shopDomain = strings.ToLower(strings.TrimSpace(shopDomain))
	shopDomain = strings.TrimPrefix(shopDomain, "https://")
	shopDomain = strings.TrimPrefix(shopDomain, "http://")
	shopDomain = strings.TrimSuffix(shopDomain, "/")

	if !strings.HasSuffix(shopDomain, ".myshopify.com") {
		fmt.Fprintf(os.Stderr, "Error: Shop domain must end with .myshopify.com\n")
		os.Exit(1)
	}

	ctx := context.Background()

	// Connect to database
	client, err := database.NewClient(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	db := database.GetDatabase(client, "mgsearch")

	// Generate keys
	publicKey, err := security.GenerateAPIKey(16)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating public key: %v\n", err)
		os.Exit(1)
	}

	privateKey, err := security.GenerateAPIKey(32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating private key: %v\n", err)
		os.Exit(1)
	}

	webhookSecret, err := security.GenerateAPIKey(32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating webhook secret: %v\n", err)
		os.Exit(1)
	}

	// Generate dummy encrypted token (32 random bytes)
	dummyToken := make([]byte, 32)
	if _, err := rand.Read(dummyToken); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating dummy token: %v\n", err)
		os.Exit(1)
	}
	
	// Decode encryption key from hex
	encryptionKey, err := security.MustDecodeKey(cfg.EncryptionKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding encryption key: %v\n", err)
		os.Exit(1)
	}
	
	encryptedToken, err := security.EncryptAESGCM(encryptionKey, dummyToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encrypting token: %v\n", err)
		os.Exit(1)
	}

	// Build index UID
	indexUID := "products_" + strings.ReplaceAll(shopDomain, ".", "_")

	// Get Meilisearch URL (use from config)
	meiliURL := cfg.MeilisearchURL
	if meiliURL == "" {
		fmt.Fprintf(os.Stderr, "Warning: MEILISEARCH_URL not set, store will be created without it\n")
	}

	// Create store
	storeRepo := repositories.NewStoreRepository(db)
	storeModel := &models.Store{
		ShopDomain:           shopDomain,
		ShopName:             shopName,
		EncryptedAccessToken: encryptedToken,
		APIKeyPublic:         publicKey,
		APIKeyPrivate:        privateKey,
		ProductIndexUID:      indexUID,
		MeilisearchIndexUID:  indexUID,
		MeilisearchDocType:   "product",
		MeilisearchURL:       meiliURL,
		PlanLevel:            "free",
		Status:               "active",
		WebhookSecret:        webhookSecret,
		InstalledAt:          time.Now().UTC(),
		SyncState: map[string]interface{}{
			"status": "pending_initial_sync",
		},
	}

	dbStore, err := storeRepo.CreateOrUpdate(ctx, storeModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating store: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Store created successfully!")
	fmt.Println("")
	fmt.Println("Store ID (UUID):", dbStore.ID)
	fmt.Println("Shop Domain:", dbStore.ShopDomain)
	fmt.Println("Storefront Key:", dbStore.APIKeyPublic)
	fmt.Println("")
	fmt.Println("Use this storefront key for search requests:")
	fmt.Printf("  X-Storefront-Key: %s\n", dbStore.APIKeyPublic)
	fmt.Println("")
	fmt.Println("Generate a JWT token with:")
	fmt.Printf("  go run scripts/generate-token.go %s %s\n", dbStore.ID, dbStore.ShopDomain)
}
