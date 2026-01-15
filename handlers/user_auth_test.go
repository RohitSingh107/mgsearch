package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mgsearch/config"
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

func setupUserAuthTest(t *testing.T) (*gin.Engine, *UserAuthHandler, *repositories.UserRepository, *repositories.ClientRepository, *config.Config, func()) {
	ctx := context.Background()
	cfg := testhelpers.TestConfig()

	_, db, cleanup, err := testhelpers.SetupTestDatabase(ctx, cfg)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)

	handler := NewUserAuthHandler(cfg, userRepo, clientRepo)
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSigningKey)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	v1 := router.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register/user", handler.RegisterUser)
			authGroup.POST("/register/client", jwtMiddleware.RequireAuth(), handler.RegisterClient)
			authGroup.POST("/login", handler.Login)
			authGroup.GET("/me", jwtMiddleware.RequireAuth(), handler.GetCurrentUser)
			authGroup.PUT("/user", jwtMiddleware.RequireAuth(), handler.UpdateUser)
		}

		clientsGroup := v1.Group("/clients")
		clientsGroup.Use(jwtMiddleware.RequireAuth())
		{
			clientsGroup.GET("", handler.GetUserClients)
			clientsGroup.GET("/:client_id", handler.GetClientDetails)
			clientsGroup.POST("/:client_id/api-keys", handler.GenerateAPIKey)
			clientsGroup.DELETE("/:client_id/api-keys/:key_id", handler.RevokeAPIKey)
		}
	}

	return router, handler, userRepo, clientRepo, cfg, func() {
		testhelpers.CleanupTestDatabase(ctx, db)
		cleanup()
	}
}

func TestUserAuthHandler_RegisterUser(t *testing.T) {
	router, _, userRepo, _, _, cleanup := setupUserAuthTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "valid user registration",
			body: map[string]interface{}{
				"email":      "test@example.com",
				"password":   "SecurePass123!",
				"first_name": "John",
				"last_name":  "Doe",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "user")
				assert.Contains(t, result, "token")
				assert.NotEmpty(t, result["token"])
				
				user := result["user"].(map[string]interface{})
				assert.Equal(t, "test@example.com", user["email"])
				assert.Equal(t, "John", user["first_name"])
				assert.Equal(t, "Doe", user["last_name"])
				assert.Equal(t, true, user["is_active"])
			},
		},
		{
			name: "duplicate email registration",
			body: map[string]interface{}{
				"email":      "duplicate@example.com",
				"password":   "SecurePass123!",
				"first_name": "Jane",
				"last_name":  "Smith",
			},
			expectedStatus: http.StatusConflict,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				// First register the user
				user := &models.User{
					Email:        "duplicate@example.com",
					PasswordHash: "hashed",
					FirstName:    "Existing",
					LastName:     "User",
					ClientIDs:    []primitive.ObjectID{},
					IsActive:     true,
				}
				_, err := userRepo.Create(context.Background(), user)
				require.NoError(t, err)
			},
		},
		{
			name: "missing email",
			body: map[string]interface{}{
				"password":   "SecurePass123!",
				"first_name": "John",
				"last_name":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			body: map[string]interface{}{
				"email":      "not-an-email",
				"password":   "SecurePass123!",
				"first_name": "John",
				"last_name":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			body: map[string]interface{}{
				"email":      "test2@example.com",
				"password":   "short",
				"first_name": "John",
				"last_name":  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing first name",
			body: map[string]interface{}{
				"email":     "test3@example.com",
				"password":  "SecurePass123!",
				"last_name": "Doe",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing last name",
			body: map[string]interface{}{
				"email":      "test4@example.com",
				"password":   "SecurePass123!",
				"first_name": "John",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "email with spaces should be trimmed",
			body: map[string]interface{}{
				"email":      "  spaced@example.com  ",
				"password":   "SecurePass123!",
				"first_name": "John",
				"last_name":  "Doe",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				user := result["user"].(map[string]interface{})
				assert.Equal(t, "spaced@example.com", user["email"])
			},
		},
		{
			name: "email should be lowercase",
			body: map[string]interface{}{
				"email":      "UPPERCASE@EXAMPLE.COM",
				"password":   "SecurePass123!",
				"first_name": "John",
				"last_name":  "Doe",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				user := result["user"].(map[string]interface{})
				assert.Equal(t, "uppercase@example.com", user["email"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run validation setup if provided
			if tt.validate != nil && tt.expectedStatus == http.StatusConflict {
				recorder := httptest.NewRecorder()
				tt.validate(t, recorder)
			}

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/register/user", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Response body: %s", w.Body.String())
			
			if tt.validate != nil && tt.expectedStatus != http.StatusConflict {
				tt.validate(t, w)
			}
		})
	}
}

func TestUserAuthHandler_Login(t *testing.T) {
	router, _, userRepo, _, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create a test user
	passwordHash, _ := auth.HashPassword("SecurePass123!")
	testUser := &models.User{
		Email:        "login@example.com",
		PasswordHash: passwordHash,
		FirstName:    "Test",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	_, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Create an inactive user
	inactiveUser := &models.User{
		Email:        "inactive@example.com",
		PasswordHash: passwordHash,
		FirstName:    "Inactive",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     false,
	}
	_, err = userRepo.Create(context.Background(), inactiveUser)
	require.NoError(t, err)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			body: map[string]interface{}{
				"email":    "login@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "user")
				assert.Contains(t, result, "token")
				assert.NotEmpty(t, result["token"])
				
				// Verify token is valid
				token := result["token"].(string)
				claims, err := auth.ParseJWT(token, []byte(cfg.JWTSigningKey))
				require.NoError(t, err)
				assert.Equal(t, "login@example.com", claims.Email)
			},
		},
		{
			name: "wrong password",
			body: map[string]interface{}{
				"email":    "login@example.com",
				"password": "WrongPassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "non-existent email",
			body: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "inactive user",
			body: map[string]interface{}{
				"email":    "inactive@example.com",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "account is inactive", result["error"])
			},
		},
		{
			name: "missing email",
			body: map[string]interface{}{
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			body: map[string]interface{}{
				"email": "login@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "case insensitive email",
			body: map[string]interface{}{
				"email":    "LOGIN@EXAMPLE.COM",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "email with spaces",
			body: map[string]interface{}{
				"email":    "  login@example.com  ",
				"password": "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Response body: %s", w.Body.String())
			
			if tt.validate != nil {
				tt.validate(t, w)
			}
		})
	}
}

func TestUserAuthHandler_GetCurrentUser(t *testing.T) {
	router, _, userRepo, _, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create a test user
	testUser := &models.User{
		Email:        "current@example.com",
		PasswordHash: "hashed",
		FirstName:    "Current",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Generate valid JWT
	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid token",
			token:          validToken,
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "user")
				user := result["user"].(map[string]interface{})
				assert.Equal(t, "current@example.com", user["email"])
			},
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDk0NTkyMDB9.xxx",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
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

func TestUserAuthHandler_UpdateUser(t *testing.T) {
	router, _, userRepo, _, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create a test user
	testUser := &models.User{
		Email:        "update@example.com",
		PasswordHash: "hashed",
		FirstName:    "Original",
		LastName:     "Name",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:  "update first name",
			token: validToken,
			body: map[string]interface{}{
				"first_name": "Updated",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				user := result["user"].(map[string]interface{})
				assert.Equal(t, "Updated", user["first_name"])
				assert.Equal(t, "Name", user["last_name"]) // Unchanged
			},
		},
		{
			name:  "update last name",
			token: validToken,
			body: map[string]interface{}{
				"last_name": "NewLastName",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "update both names",
			token: validToken,
			body: map[string]interface{}{
				"first_name": "New",
				"last_name":  "Names",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "empty update (no fields changed)",
			token: validToken,
			body:  map[string]interface{}{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			body:           map[string]interface{}{"first_name": "Test"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PUT", "/api/v1/auth/user", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
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

func TestUserAuthHandler_RegisterClient(t *testing.T) {
	router, _, userRepo, clientRepo, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create a test user
	testUser := &models.User{
		Email:        "client@example.com",
		PasswordHash: "hashed",
		FirstName:    "Client",
		LastName:     "Owner",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	// Create existing client for duplicate test
	existingClient := &models.Client{
		Name:        "existing-client",
		Description: "Existing",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}
	_, err = clientRepo.Create(context.Background(), existingClient)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:  "valid client creation",
			token: validToken,
			body: map[string]interface{}{
				"name":        "my-new-client",
				"description": "My Application",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "client")
				client := result["client"].(map[string]interface{})
				assert.Equal(t, "my-new-client", client["name"])
				assert.Equal(t, "My Application", client["description"])
			},
		},
		{
			name:  "duplicate client name",
			token: validToken,
			body: map[string]interface{}{
				"name": "existing-client",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:  "missing name",
			token: validToken,
			body: map[string]interface{}{
				"description": "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "client without description",
			token: validToken,
			body: map[string]interface{}{
				"name": "no-description-client",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing token",
			token:          "",
			body:           map[string]interface{}{"name": "test"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/auth/register/client", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
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

func TestUserAuthHandler_GenerateAPIKey(t *testing.T) {
	router, _, userRepo, clientRepo, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create user and client
	testUser := &models.User{
		Email:        "apikey@example.com",
		PasswordHash: "hashed",
		FirstName:    "API",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	testClient := &models.Client{
		Name:        "test-client",
		Description: "Test",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}
	testClient, err = clientRepo.Create(context.Background(), testClient)
	require.NoError(t, err)

	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	// Create another user without access
	otherUser := &models.User{
		Email:        "other@example.com",
		PasswordHash: "hashed",
		FirstName:    "Other",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	otherUser, err = userRepo.Create(context.Background(), otherUser)
	require.NoError(t, err)

	otherToken, err := auth.GenerateJWT(otherUser.ID.Hex(), otherUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		clientID       string
		body           map[string]interface{}
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:     "valid API key generation",
			token:    validToken,
			clientID: testClient.ID.Hex(),
			body: map[string]interface{}{
				"name": "Production Key",
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "api_key")
				assert.Contains(t, result, "key_id")
				assert.Contains(t, result, "prefix")
				assert.Contains(t, result, "warning")
				assert.NotEmpty(t, result["api_key"])
				assert.Len(t, result["api_key"].(string), 64) // 32 bytes = 64 hex chars
			},
		},
		{
			name:     "API key with permissions",
			token:    validToken,
			clientID: testClient.ID.Hex(),
			body: map[string]interface{}{
				"name":        "Read-Only Key",
				"permissions": []string{"search", "read"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "API key with expiration",
			token:    validToken,
			clientID: testClient.ID.Hex(),
			body: map[string]interface{}{
				"name":       "Temporary Key",
				"expires_at": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "missing name",
			token:    validToken,
			clientID: testClient.ID.Hex(),
			body:     map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid client ID",
			token:          validToken,
			clientID:       "invalid-id",
			body:           map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent client",
			token:          validToken,
			clientID:       primitive.NewObjectID().Hex(),
			body:           map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "access denied - user not linked to client",
			token:          otherToken,
			clientID:       testClient.ID.Hex(),
			body:           map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing token",
			token:          "",
			clientID:       testClient.ID.Hex(),
			body:           map[string]interface{}{"name": "Test"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "invalid expiration format",
			token:    validToken,
			clientID: testClient.ID.Hex(),
			body: map[string]interface{}{
				"name":       "Test",
				"expires_at": "invalid-date",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			url := "/api/v1/clients/" + tt.clientID + "/api-keys"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
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

func TestUserAuthHandler_RevokeAPIKey(t *testing.T) {
	router, _, userRepo, clientRepo, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create user and client with API key
	testUser := &models.User{
		Email:        "revoke@example.com",
		PasswordHash: "hashed",
		FirstName:    "Revoke",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	apiKeyID := primitive.NewObjectID()
	testClient := &models.Client{
		Name:        "revoke-test-client",
		Description: "Test",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys: []models.APIKey{
			{
				ID:          apiKeyID,
				Key:         "hashed-key",
				Name:        "Test Key",
				KeyPrefix:   "test_",
				Permissions: []string{},
				IsActive:    true,
				CreatedAt:   time.Now(),
			},
		},
		IsActive: true,
	}
	testClient, err = clientRepo.Create(context.Background(), testClient)
	require.NoError(t, err)

	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		clientID       string
		keyID          string
		expectedStatus int
	}{
		{
			name:           "valid revocation",
			token:          validToken,
			clientID:       testClient.ID.Hex(),
			keyID:          apiKeyID.Hex(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid key ID",
			token:          validToken,
			clientID:       testClient.ID.Hex(),
			keyID:          "invalid-id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent key",
			token:          validToken,
			clientID:       testClient.ID.Hex(),
			keyID:          primitive.NewObjectID().Hex(),
			expectedStatus: http.StatusInternalServerError, // Key not found
		},
		{
			name:           "missing token",
			token:          "",
			clientID:       testClient.ID.Hex(),
			keyID:          apiKeyID.Hex(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/clients/" + tt.clientID + "/api-keys/" + tt.keyID
			req := httptest.NewRequest("DELETE", url, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserAuthHandler_GetUserClients(t *testing.T) {
	router, _, userRepo, clientRepo, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	// Create user with multiple clients
	testUser := &models.User{
		Email:        "clients@example.com",
		PasswordHash: "hashed",
		FirstName:    "Multi",
		LastName:     "Client",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	// Create clients
	client1 := &models.Client{
		Name:        "client-one",
		Description: "First Client",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}
	_, err = clientRepo.Create(context.Background(), client1)
	require.NoError(t, err)

	client2 := &models.Client{
		Name:        "client-two",
		Description: "Second Client",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}
	_, err = clientRepo.Create(context.Background(), client2)
	require.NoError(t, err)

	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "get user clients",
			token:          validToken,
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "clients")
				clients := result["clients"].([]interface{})
				assert.Len(t, clients, 2)
			},
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/clients", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
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

func TestUserAuthHandler_GetClientDetails(t *testing.T) {
	router, _, userRepo, clientRepo, cfg, cleanup := setupUserAuthTest(t)
	defer cleanup()

	testUser := &models.User{
		Email:        "details@example.com",
		PasswordHash: "hashed",
		FirstName:    "Details",
		LastName:     "User",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	testUser, err := userRepo.Create(context.Background(), testUser)
	require.NoError(t, err)

	testClient := &models.Client{
		Name:        "details-client",
		Description: "Client for details test",
		UserIDs:     []primitive.ObjectID{testUser.ID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}
	testClient, err = clientRepo.Create(context.Background(), testClient)
	require.NoError(t, err)

	validToken, err := auth.GenerateJWT(testUser.ID.Hex(), testUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	// Create user without access
	otherUser := &models.User{
		Email:        "noaccess@example.com",
		PasswordHash: "hashed",
		FirstName:    "No",
		LastName:     "Access",
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}
	otherUser, err = userRepo.Create(context.Background(), otherUser)
	require.NoError(t, err)

	noAccessToken, err := auth.GenerateJWT(otherUser.ID.Hex(), otherUser.Email, []byte(cfg.JWTSigningKey), 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		clientID       string
		expectedStatus int
		validate       func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:           "valid client details",
			token:          validToken,
			clientID:       testClient.ID.Hex(),
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *httptest.ResponseRecorder) {
				var result map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err)
				assert.Contains(t, result, "client")
				client := result["client"].(map[string]interface{})
				assert.Equal(t, "details-client", client["name"])
			},
		},
		{
			name:           "access denied",
			token:          noAccessToken,
			clientID:       testClient.ID.Hex(),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid client ID",
			token:          validToken,
			clientID:       "invalid-id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "non-existent client",
			token:          validToken,
			clientID:       primitive.NewObjectID().Hex(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "missing token",
			token:          "",
			clientID:       testClient.ID.Hex(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/clients/" + tt.clientID
			req := httptest.NewRequest("GET", url, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
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
