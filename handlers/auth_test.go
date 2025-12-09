package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mgsearch/models"
	"mgsearch/middleware"
	"mgsearch/pkg/auth"
	"mgsearch/repositories"
	"mgsearch/services"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthTest(t *testing.T) (*gin.Engine, *repositories.StoreRepository, *services.ShopifyService, *services.MeilisearchService, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	storeRepo, _ := testhelpers.SetupTestRepositories(db)
	meiliService := services.NewMeilisearchService(cfg)
	shopifyService := services.NewShopifyService(cfg)

	// Setup router directly to avoid import cycle
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	authHandler, err := NewAuthHandler(cfg, shopifyService, storeRepo, meiliService)
	require.NoError(t, err)

	api := router.Group("/api")
	{
		shopifyGroup := api.Group("/auth/shopify")
		{
			shopifyGroup.POST("/begin", authHandler.Begin)
			shopifyGroup.GET("/callback", authHandler.Callback)
			shopifyGroup.POST("/exchange", authHandler.ExchangeToken)
			shopifyGroup.POST("/install", authHandler.InstallStore)
		}
	}

	return router, storeRepo, shopifyService, meiliService, func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

func TestAuthHandler_Begin(t *testing.T) {
	router, _, _, _, cleanup := setupAuthTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "valid shop domain",
			body: map[string]interface{}{
				"shop": "test-store.myshopify.com",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "authUrl")
				assert.Contains(t, result, "state")
				assert.NotEmpty(t, result["authUrl"])
				assert.NotEmpty(t, result["state"])
			},
		},
		{
			name: "valid shop with redirect_uri",
			body: map[string]interface{}{
				"shop":         "test-store.myshopify.com",
				"redirect_uri": "https://custom-redirect.example.com/callback",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "authUrl")
				assert.Contains(t, result["authUrl"], "custom-redirect.example.com")
			},
		},
		{
			name: "invalid shop domain - missing myshopify.com",
			body: map[string]interface{}{
				"shop": "test-store.com",
			},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "invalid shop domain", result["error"])
			},
		},
		{
			name: "missing shop field",
			body: map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty shop field",
			body: map[string]interface{}{
				"shop": "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/auth/shopify/begin", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestAuthHandler_InstallStore(t *testing.T) {
	router, storeRepo, _, _, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create a test store first to test update scenario
	testStore := &models.Store{
		ShopDomain:           "existing-store.myshopify.com",
		ShopName:             "Existing Store",
		EncryptedAccessToken: []byte("encrypted-token"),
		APIKeyPublic:         "existing-public-key",
		APIKeyPrivate:        "existing-private-key",
		ProductIndexUID:      "products_existing",
		MeilisearchIndexUID:  "products_existing",
		MeilisearchDocType:   "product",
		MeilisearchURL:       "http://localhost:7700",
		PlanLevel:            "free",
		Status:               "active",
		WebhookSecret:        "webhook-secret",
		InstalledAt:          time.Now(),
		SyncState:            map[string]interface{}{"status": "pending"},
	}
	_, err := storeRepo.CreateOrUpdate(context.Background(), testStore)
	require.NoError(t, err)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "valid installation",
			body: map[string]interface{}{
				"shop":         "new-store.myshopify.com",
				"access_token": "test-access-token",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "store")
				assert.Contains(t, result, "token")
				assert.Contains(t, result, "message")
			},
		},
		{
			name: "update existing store",
			body: map[string]interface{}{
				"shop":         "existing-store.myshopify.com",
				"access_token": "new-access-token",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "store")
			},
		},
		{
			name: "missing access_token",
			body: map[string]interface{}{
				"shop": "test-store.myshopify.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid shop domain",
			body: map[string]interface{}{
				"shop":         "invalid-shop.com",
				"access_token": "token",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/auth/shopify/install", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Meilisearch-Url", "http://localhost:7700")
			req.Header.Set("X-Meilisearch-Api-Key", "test-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestAuthHandler_ExchangeToken(t *testing.T) {
	router, _, _, _, cleanup := setupAuthTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid exchange request",
			body: map[string]interface{}{
				"shop": "test-store.myshopify.com",
				"code": "test-code",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing code",
			body: map[string]interface{}{
				"shop": "test-store.myshopify.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing shop",
			body: map[string]interface{}{
				"code": "test-code",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/auth/shopify/exchange", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Callback(t *testing.T) {
	router, _, _, _, cleanup := setupAuthTest(t)
	defer cleanup()

	cfg := testhelpers.TestConfig()

	// Generate a valid state token
	state, err := auth.GenerateStateToken("test-store.myshopify.com", []byte(cfg.JWTSigningKey), 15*time.Minute)
	require.NoError(t, err)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "valid callback with state",
			queryParams:    "?shop=test-store.myshopify.com&code=test-code&state=" + state,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing shop parameter",
			queryParams:    "?code=test-code&state=" + state,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing code parameter",
			queryParams:    "?shop=test-store.myshopify.com&state=" + state,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing state parameter",
			queryParams:    "?shop=test-store.myshopify.com&code=test-code",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/auth/shopify/callback"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

