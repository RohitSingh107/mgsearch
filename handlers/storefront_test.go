package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mgsearch/middleware"
	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupStorefrontTest(t *testing.T) (*gin.Engine, *repositories.StoreRepository, string, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	storeRepo, _ := testhelpers.SetupTestRepositories(db)
	meiliService := services.NewMeilisearchService(cfg)

	// Create a test store with public API key
	testStore := &models.Store{
		ID:                   primitive.NewObjectID(),
		ShopDomain:           "storefront-test.myshopify.com",
		ShopName:             "Storefront Test Store",
		EncryptedAccessToken: []byte("encrypted-token"),
		APIKeyPublic:         "storefront-public-key-123",
		APIKeyPrivate:        "private-key",
		ProductIndexUID:      "products_storefront_test",
		MeilisearchIndexUID:  "products_storefront_test",
		MeilisearchDocType:   "product",
		MeilisearchURL:       "http://localhost:7700",
		PlanLevel:            "free",
		Status:               "active",
		WebhookSecret:        "webhook-secret",
		InstalledAt:          time.Now(),
		SyncState:            map[string]interface{}{},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	_, err = storeRepo.CreateOrUpdate(ctx, testStore)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	storefrontHandler := NewStorefrontHandler(storeRepo, meiliService)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/search", storefrontHandler.Search)
		v1.POST("/search", storefrontHandler.Search)
	}

	return router, storeRepo, "storefront-public-key-123", func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

func TestStorefrontHandler_Search_GET(t *testing.T) {
	router, _, publicKey, cleanup := setupStorefrontTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		storefrontKey  string
		queryParams    string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid GET request with query",
			storefrontKey:  publicKey,
			queryParams:    "?q=shoes&limit=10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid GET request with filters",
			storefrontKey:  publicKey,
			queryParams:    "?q=shoes&filters=[\"price > 100\"]",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing storefront key",
			storefrontKey:  "",
			queryParams:    "?q=shoes",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid storefront key",
			storefrontKey:  "invalid-key",
			queryParams:    "?q=shoes",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "OPTIONS preflight request",
			storefrontKey:  publicKey,
			queryParams:    "",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := "GET"
			if tt.name == "OPTIONS preflight request" {
				method = "OPTIONS"
			}

			req := httptest.NewRequest(method, "/api/v1/search"+tt.queryParams, nil)
			if tt.storefrontKey != "" {
				req.Header.Set("X-Storefront-Key", tt.storefrontKey)
			}
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestStorefrontHandler_Search_POST(t *testing.T) {
	router, _, publicKey, cleanup := setupStorefrontTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		storefrontKey  string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:          "valid POST request",
			storefrontKey: publicKey,
			body: map[string]interface{}{
				"q":      "shoes",
				"limit":  10,
				"offset": 0,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "POST with filters and sorting",
			storefrontKey: publicKey,
			body: map[string]interface{}{
				"q":      "shoes",
				"filter": "price > 100",
				"sort":   []string{"price:asc"},
				"limit":  20,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing storefront key",
			storefrontKey:  "",
			body:           map[string]interface{}{"q": "shoes"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid storefront key",
			storefrontKey:  "invalid-key",
			body:           map[string]interface{}{"q": "shoes"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:          "invalid request body",
			storefrontKey: publicKey,
			body:          nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			} else {
				bodyBytes = []byte("invalid json")
			}

			req := httptest.NewRequest("POST", "/api/v1/search", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.storefrontKey != "" {
				req.Header.Set("X-Storefront-Key", tt.storefrontKey)
			}
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

