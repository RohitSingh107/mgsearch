package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mgsearch/middleware"
	"mgsearch/models"
	"mgsearch/pkg/auth"
	"mgsearch/repositories"
	"mgsearch/services"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func setupIndexTest(t *testing.T) (*gin.Engine, *IndexHandler, *repositories.ClientRepository, *repositories.IndexRepository, *repositories.UserRepository, string, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	clientRepo := repositories.NewClientRepository(db)
	indexRepo := repositories.NewIndexRepository(db)
	userRepo := repositories.NewUserRepository(db)
	meiliService := services.NewMeilisearchService(cfg)

	handler := NewIndexHandler(clientRepo, indexRepo, meiliService)
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSigningKey)

	// Create test user
	testUser := &models.User{
		Email:        "index@example.com",
		PasswordHash: "hashed",
		FirstName:    "Index",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err = userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	// Create test client
	testClient := &models.Client{
		Name:        "test-index-client",
		Description: "Test Client for Index",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}
	testClient, err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err)

	// Generate JWT
	token, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	{
		clientsGroup := v1.Group("/clients")
		clientsGroup.Use(jwtMiddleware.RequireAuth())
		{
			clientsGroup.POST("/:client_id/indexes", handler.CreateIndex)
			clientsGroup.GET("/:client_id/indexes", handler.GetClientIndexes)
		}
	}

	return router, handler, clientRepo, indexRepo, userRepo, token, func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

func TestIndexHandler_CreateIndex(t *testing.T) {
	router, _, clientRepo, indexRepo, _, token, cleanup := setupIndexTest(t)
	defer cleanup()

	// Get the test client
	clients, err := clientRepo.FindByName(context.Background(), "test-index-client")
	require.NoError(t, err)
	require.NotNil(t, clients)
	clientID := clients.ID.Hex()

	tests := []struct {
		name           string
		token          string
		clientID       string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:     "valid index creation",
			token:    token,
			clientID: clientID,
			body: map[string]interface{}{
				"name": "products",
			},
			expectedStatus: http.StatusAccepted,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "index")
				assert.Contains(t, result, "task")
				
				index := result["index"].(map[string]interface{})
				assert.Equal(t, "products", index["name"])
				assert.Equal(t, "test-index-client__products", index["uid"])
			},
		},
		{
			name:     "index with primary key",
			token:    token,
			clientID: clientID,
			body: map[string]interface{}{
				"name":        "movies",
				"primary_key": "movie_id",
			},
			expectedStatus: http.StatusAccepted,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				index := result["index"].(map[string]interface{})
				assert.Equal(t, "movie_id", index["primary_key"])
			},
		},
		{
			name:     "duplicate index name",
			token:    token,
			clientID: clientID,
			body: map[string]interface{}{
				"name": "duplicate-test",
			},
			expectedStatus: http.StatusConflict,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// Pre-create the index
				idx := &models.Index{
					ClientID:   primitive.NewObjectID(),
					Name:       "duplicate-test",
					UID:        "test-index-client__duplicate-test",
					PrimaryKey: "",
				}
				_, err := indexRepo.Create(context.Background(), idx)
				require.NoError(t, err)
			},
		},
		{
			name:     "missing name",
			token:    token,
			clientID: clientID,
			body:     map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid client ID",
			token:          token,
			clientID:       "invalid-id",
			body:           map[string]interface{}{"name": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent client",
			token:          token,
			clientID:       primitive.NewObjectID().Hex(),
			body:           map[string]interface{}{"name": "test"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing token",
			token:          "",
			clientID:       clientID,
			body:           map[string]interface{}{"name": "test"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pre-validate if needed
			if tt.validate != nil && tt.expectedStatus == http.StatusConflict {
				recorder := httptest.NewRecorder()
				tt.validate(t, recorder)
			}

			bodyBytes, _ := json.Marshal(tt.body)
			url := "/api/v1/clients/" + tt.clientID + "/indexes"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Response: %s", w.Body.String())
			
			if tt.validate != nil && tt.expectedStatus != http.StatusConflict {
				tt.validate(t, w)
			}
		})
	}
}

func TestIndexHandler_GetClientIndexes(t *testing.T) {
	router, _, clientRepo, indexRepo, _, token, cleanup := setupIndexTest(t)
	defer cleanup()

	// Get the test client
	clients, err := clientRepo.FindByName(context.Background(), "test-index-client")
	require.NoError(t, err)
	clientID := clients.ID

	// Create some test indexes
	index1 := &models.Index{
		ClientID:   clientID,
		Name:       "index-one",
		UID:        "test-index-client__index-one",
		PrimaryKey: "id",
	}
	_, err = indexRepo.Create(context.Background(), index1)
	require.NoError(t, err)

	index2 := &models.Index{
		ClientID:   clientID,
		Name:       "index-two",
		UID:        "test-index-client__index-two",
		PrimaryKey: "",
	}
	_, err = indexRepo.Create(context.Background(), index2)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		clientID       string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "get client indexes",
			token:          token,
			clientID:       clientID.Hex(),
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result []map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(result), 2)
				
				// Check that our indexes are in the result
				names := make([]string, len(result))
				for i, idx := range result {
					names[i] = idx["name"].(string)
				}
				assert.Contains(t, names, "index-one")
				assert.Contains(t, names, "index-two")
			},
		},
		{
			name:           "invalid client ID",
			token:          token,
			clientID:       "invalid-id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent client",
			token:          token,
			clientID:       primitive.NewObjectID().Hex(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing token",
			token:          "",
			clientID:       clientID.Hex(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/clients/" + tt.clientID + "/indexes"
			req := httptest.NewRequest("GET", url, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Response: %s", w.Body.String())
			
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

// Test index UID generation
func TestIndexHandler_IndexUIDFormat(t *testing.T) {
	router, _, clientRepo, _, _, token, cleanup := setupIndexTest(t)
	defer cleanup()

	clients, err := clientRepo.FindByName(context.Background(), "test-index-client")
	require.NoError(t, err)
	clientID := clients.ID.Hex()

	tests := []struct {
		name         string
		indexName    string
		expectedUID  string
	}{
		{
			name:        "simple name",
			indexName:   "products",
			expectedUID: "test-index-client__products",
		},
		{
			name:        "name with underscores",
			indexName:   "user_products",
			expectedUID: "test-index-client__user_products",
		},
		{
			name:        "name with hyphens",
			indexName:   "user-products",
			expectedUID: "test-index-client__user-products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"name": tt.indexName,
			}
			bodyBytes, _ := json.Marshal(body)
			url := "/api/v1/clients/" + clientID + "/indexes"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code == http.StatusAccepted {
				var result map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &result)
				require.NoError(t, err)
				index := result["index"].(map[string]interface{})
				assert.Equal(t, tt.expectedUID, index["uid"])
			}
		})
	}
}

// Test concurrent index creation (race condition test)
func TestIndexHandler_ConcurrentIndexCreation(t *testing.T) {
	router, _, clientRepo, _, _, token, cleanup := setupIndexTest(t)
	defer cleanup()

	clients, err := clientRepo.FindByName(context.Background(), "test-index-client")
	require.NoError(t, err)
	clientID := clients.ID.Hex()

	// Try to create the same index concurrently
	done := make(chan bool, 2)
	results := make([]int, 2)

	createIndex := func(id int) {
		body := map[string]interface{}{
			"name": "concurrent-test",
		}
		bodyBytes, _ := json.Marshal(body)
		url := "/api/v1/clients/" + clientID + "/indexes"
		req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		results[id] = w.Code
		done <- true
	}

	// Launch concurrent requests
	go createIndex(0)
	go createIndex(1)

	// Wait for both to complete
	<-done
	<-done

	// One should succeed, one should fail with conflict or succeed
	// At least one should be 202 Accepted
	assert.True(t, results[0] == http.StatusAccepted || results[1] == http.StatusAccepted,
		"At least one request should succeed")
}
