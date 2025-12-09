package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupWebhookTest(t *testing.T) (*gin.Engine, *repositories.StoreRepository, string, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	storeRepo, _ := testhelpers.SetupTestRepositories(db)
	meiliService := services.NewMeilisearchService(cfg)
	shopifyService := services.NewShopifyService(cfg)

	// Create a test store
	testStore := &models.Store{
		ID:                   primitive.NewObjectID(),
		ShopDomain:           "webhook-test.myshopify.com",
		ShopName:             "Webhook Test Store",
		EncryptedAccessToken: []byte("encrypted-token"),
		APIKeyPublic:         "webhook-public-key",
		APIKeyPrivate:        "private-key",
		ProductIndexUID:      "products_webhook_test",
		MeilisearchIndexUID:  "products_webhook_test",
		MeilisearchDocType:   "product",
		MeilisearchURL:       "http://localhost:7700",
		PlanLevel:            "free",
		Status:               "active",
		WebhookSecret:        cfg.WebhookSharedSecret,
		InstalledAt:          time.Now(),
		SyncState:            map[string]interface{}{},
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	_, err = storeRepo.CreateOrUpdate(ctx, testStore)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	webhookHandler := NewWebhookHandler(shopifyService, storeRepo, meiliService)

	router.POST("/webhooks/shopify/:topic/:subtopic", webhookHandler.HandleShopifyWebhook)

	return router, storeRepo, cfg.WebhookSharedSecret, func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

// calculateHMAC calculates the HMAC-SHA256 signature for webhook verification
func calculateHMAC(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestWebhookHandler_HandleShopifyWebhook(t *testing.T) {
	router, _, secret, cleanup := setupWebhookTest(t)
	defer cleanup()

	productPayload := map[string]interface{}{
		"id":    12345,
		"title": "Test Product",
		"price": 99.99,
	}
	bodyBytes, _ := json.Marshal(productPayload)
	bodyStr := string(bodyBytes)
	signature := calculateHMAC(secret, bodyStr)

	tests := []struct {
		name           string
		topic          string
		subtopic       string
		body           []byte
		signature      string
		shopDomain     string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "products/create event",
			topic:          "products",
			subtopic:       "create",
			body:           bodyBytes,
			signature:      signature,
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "processed", result["status"])
			},
		},
		{
			name:           "products/update event",
			topic:          "products",
			subtopic:       "update",
			body:           bodyBytes,
			signature:      signature,
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "products/delete event",
			topic:          "products",
			subtopic:       "delete",
			body:           bodyBytes,
			signature:      signature,
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unknown event (ignored)",
			topic:          "orders",
			subtopic:       "create",
			body:           bodyBytes,
			signature:      signature,
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "event ignored", result["message"])
			},
		},
		{
			name:           "missing signature header",
			topic:          "products",
			subtopic:       "create",
			body:           bodyBytes,
			signature:      "",
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing shop domain header",
			topic:          "products",
			subtopic:       "create",
			body:           bodyBytes,
			signature:      signature,
			shopDomain:     "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid signature",
			topic:          "products",
			subtopic:       "create",
			body:           bodyBytes,
			signature:      "invalid-signature",
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "store not registered",
			topic:          "products",
			subtopic:       "create",
			body:           bodyBytes,
			signature:      signature,
			shopDomain:     "nonexistent.myshopify.com",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "product without ID",
			topic:          "products",
			subtopic:       "create",
			body:           []byte(`{"title": "Product without ID"}`),
			signature:      calculateHMAC(secret, `{"title": "Product without ID"}`),
			shopDomain:     "webhook-test.myshopify.com",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/webhooks/shopify/" + tt.topic + "/" + tt.subtopic
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.signature != "" {
				req.Header.Set("X-Shopify-Hmac-Sha256", tt.signature)
			}
			if tt.shopDomain != "" {
				req.Header.Set("X-Shopify-Shop-Domain", tt.shopDomain)
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

