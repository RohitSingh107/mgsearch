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
)

func setupSessionTest(t *testing.T) (*gin.Engine, *repositories.SessionRepository, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	_, sessionRepo := testhelpers.SetupTestRepositories(db)
	meiliService := services.NewMeilisearchService(cfg)

	// Setup router directly
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSMiddleware())

	storeRepo, _ := testhelpers.SetupTestRepositories(db)
	sessionHandler, err := NewSessionHandler(sessionRepo, storeRepo, meiliService, cfg)
	require.NoError(t, err)

	api := router.Group("/api")
	{
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

	return router, sessionRepo, func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

func TestSessionHandler_StoreSession(t *testing.T) {
	router, _, cleanup := setupSessionTest(t)
	defer cleanup()

	expires := time.Now().Add(24 * time.Hour)
	userID := int64(12345)
	firstName := "John"
	lastName := "Doe"
	email := "john@example.com"
	locale := "en"
	collaborator := true
	emailVerified := true

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "valid session",
			body: map[string]interface{}{
				"id":          "test-session-id",
				"shop":        "test-store.myshopify.com",
				"state":       "test-state",
				"isOnline":    false,
				"scope":       "read_products",
				"expires":     expires.Format(time.RFC3339),
				"accessToken": "test-access-token",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "test-session-id", result["id"])
				assert.Equal(t, "test-store.myshopify.com", result["shop"])
			},
		},
		{
			name: "session with all fields",
			body: map[string]interface{}{
				"id":            "full-session-id",
				"shop":          "test-store.myshopify.com",
				"state":         "test-state",
				"isOnline":      true,
				"scope":         "read_products,write_products",
				"expires":       expires.Format(time.RFC3339),
				"accessToken":   "test-access-token",
				"userId":        userID,
				"firstName":     firstName,
				"lastName":      lastName,
				"email":         email,
				"accountOwner":  true,
				"locale":        locale,
				"collaborator":  collaborator,
				"emailVerified": emailVerified,
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "full-session-id", result["id"])
				assert.Equal(t, float64(userID), result["userId"])
				assert.Equal(t, firstName, result["firstName"])
			},
		},
		{
			name: "missing required fields",
			body: map[string]interface{}{
				"id": "test-session-id",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty body",
			body: map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/sessions", bytes.NewBuffer(bodyBytes))
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

func TestSessionHandler_LoadSession(t *testing.T) {
	router, sessionRepo, cleanup := setupSessionTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test session
	testSession := &models.Session{
		ID:          "test-session-123",
		Shop:        "test-store.myshopify.com",
		State:       "test-state",
		IsOnline:    false,
		Scope:       "read_products",
		AccessToken: "test-access-token",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := sessionRepo.CreateOrUpdate(ctx, testSession)
	require.NoError(t, err)

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "existing session",
			sessionID:      "test-session-123",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "test-session-123", result["id"])
				assert.Equal(t, "test-store.myshopify.com", result["shop"])
			},
		},
		{
			name:           "non-existent session",
			sessionID:      "non-existent-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty session ID",
			sessionID:      "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/sessions/"+tt.sessionID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestSessionHandler_DeleteSession(t *testing.T) {
	router, sessionRepo, cleanup := setupSessionTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test sessions
	session1 := &models.Session{
		ID:          "delete-session-1",
		Shop:        "test-store.myshopify.com",
		State:       "test-state",
		AccessToken: "token-1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := sessionRepo.CreateOrUpdate(ctx, session1)
	require.NoError(t, err)

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
	}{
		{
			name:           "delete existing session",
			sessionID:      "delete-session-1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete non-existent session",
			sessionID:      "non-existent",
			expectedStatus: http.StatusOK, // Delete is idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/sessions/"+tt.sessionID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify session is deleted
			_, err := sessionRepo.GetByID(ctx, tt.sessionID)
			if tt.sessionID == "delete-session-1" {
				assert.Error(t, err)
			}
		})
	}
}

func TestSessionHandler_DeleteMultipleSessions(t *testing.T) {
	router, sessionRepo, cleanup := setupSessionTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test sessions
	sessions := []*models.Session{
		{ID: "batch-delete-1", Shop: "test-store.myshopify.com", State: "state1", AccessToken: "token1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "batch-delete-2", Shop: "test-store.myshopify.com", State: "state2", AccessToken: "token2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "batch-delete-3", Shop: "test-store.myshopify.com", State: "state3", AccessToken: "token3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, session := range sessions {
		err := sessionRepo.CreateOrUpdate(ctx, session)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "delete multiple sessions",
			body: map[string]interface{}{
				"ids": []string{"batch-delete-1", "batch-delete-2"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "delete with empty array",
			body: map[string]interface{}{
				"ids": []string{},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "delete with non-existent IDs",
			body: map[string]interface{}{
				"ids": []string{"non-existent-1", "non-existent-2"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("DELETE", "/api/sessions/batch", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSessionHandler_FindSessionsByShop(t *testing.T) {
	router, sessionRepo, cleanup := setupSessionTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test sessions for different shops
	sessions := []*models.Session{
		{ID: "shop1-session-1", Shop: "shop1.myshopify.com", State: "state1", AccessToken: "token1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "shop1-session-2", Shop: "shop1.myshopify.com", State: "state2", AccessToken: "token2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "shop2-session-1", Shop: "shop2.myshopify.com", State: "state1", AccessToken: "token3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, session := range sessions {
		err := sessionRepo.CreateOrUpdate(ctx, session)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		shop           string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "find sessions for shop1",
			shop:           "shop1.myshopify.com",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result []map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Len(t, result, 2)
				for _, session := range result {
					assert.Equal(t, "shop1.myshopify.com", session["shop"])
				}
			},
		},
		{
			name:           "find sessions for shop2",
			shop:           "shop2.myshopify.com",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result []map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Len(t, result, 1)
			},
		},
		{
			name:           "find sessions for non-existent shop",
			shop:           "nonexistent.myshopify.com",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result []map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Len(t, result, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/sessions/shop/"+tt.shop, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

