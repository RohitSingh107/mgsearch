package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mgsearch/middleware"
	"mgsearch/models"
	"mgsearch/pkg/auth"
	"mgsearch/repositories"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupStoreTest(t *testing.T) (*gin.Engine, *repositories.StoreRepository, string, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	storeRepo, _ := testhelpers.SetupTestRepositories(db)

	// Create a test store
	testStore := &models.Store{
		ID:                   primitive.NewObjectID(),
		ShopDomain:           "test-store.myshopify.com",
		ShopName:             "Test Store",
		EncryptedAccessToken: []byte("encrypted-token"),
		APIKeyPublic:         "test-public-key-123",
		APIKeyPrivate:        "test-private-key",
		ProductIndexUID:      "products_test_store",
		MeilisearchIndexUID:  "products_test_store",
		MeilisearchDocType:   "product",
		MeilisearchURL:       "http://localhost:7700",
		PlanLevel:            "free",
		Status:               "active",
		WebhookSecret:        "webhook-secret",
		InstalledAt:          time.Now(),
		SyncState:            map[string]interface{}{"status": "synced", "last_sync": "2024-01-01"},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	createdStore, err := storeRepo.CreateOrUpdate(ctx, testStore)
	require.NoError(t, err)

	// Generate JWT token for the store
	token, err := auth.GenerateSessionToken(createdStore.ID.Hex(), createdStore.ShopDomain, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	// Setup router directly
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	storeHandler := NewStoreHandler(storeRepo)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSigningKey)

	api := router.Group("/api")
	{
		storeGroup := api.Group("/stores")
		storeGroup.Use(authMiddleware.RequireStoreSession())
		{
			storeGroup.GET("/current", storeHandler.GetCurrentStore)
			storeGroup.GET("/sync-status", storeHandler.GetSyncStatus)
		}
	}

	return router, storeRepo, token, func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

func TestStoreHandler_GetCurrentStore(t *testing.T) {
	router, _, token, cleanup := setupStoreTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "test-store.myshopify.com", result["shop_domain"])
				assert.Equal(t, "Test Store", result["shop_name"])
				assert.Equal(t, "free", result["plan_level"])
				assert.Equal(t, "active", result["status"])
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed authorization header",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/stores/current", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestStoreHandler_GetSyncStatus(t *testing.T) {
	router, _, token, cleanup := setupStoreTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "store_id")
				assert.Contains(t, result, "shop_domain")
				assert.Contains(t, result, "sync_state")
				assert.Contains(t, result, "index_uid")
				assert.Equal(t, "test-store.myshopify.com", result["shop_domain"])
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/stores/sync-status", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

