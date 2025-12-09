package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"mgsearch/config"
	"mgsearch/middleware"
	"mgsearch/models"
	"mgsearch/pkg/auth"
	"mgsearch/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAuthHandler struct {
	cfg        *config.Config
	userRepo   *repositories.UserRepository
	clientRepo *repositories.ClientRepository
}

func NewUserAuthHandler(cfg *config.Config, userRepo *repositories.UserRepository, clientRepo *repositories.ClientRepository) *UserAuthHandler {
	return &UserAuthHandler{
		cfg:        cfg,
		userRepo:   userRepo,
		clientRepo: clientRepo,
	}
}

// verifyClientAccess checks if the user has access to the specified client
// Returns the client if access is granted, nil if no access, error if client not found
func (h *UserAuthHandler) verifyClientAccess(c *gin.Context, clientID, userID primitive.ObjectID) (*models.Client, error) {
	// Get client
	client, err := h.clientRepo.FindByID(c.Request.Context(), clientID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this client
	for _, uid := range client.UserIDs {
		if uid == userID {
			return client, nil
		}
	}

	// User doesn't have access
	return nil, nil
}

// RegisterUserRequest represents the user registration request
type RegisterUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// RegisterUser handles POST /api/v1/auth/register/user
func (h *UserAuthHandler) RegisterUser(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	existingUser, _ := h.userRepo.FindByEmail(c.Request.Context(), email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		ClientIDs:    []primitive.ObjectID{},
		IsActive:     true,
	}

	user, err = h.userRepo.Create(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID.Hex(), user.Email, []byte(h.cfg.JWTSigningKey), 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user":    user.ToPublicView(),
		"token":   token,
	})
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /api/v1/auth/login
func (h *UserAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalize email
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Find user
	user, err := h.userRepo.FindByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account is inactive"})
		return
	}

	// Verify password
	if err := auth.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID.Hex(), user.Email, []byte(h.cfg.JWTSigningKey), 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user":    user.ToPublicView(),
		"token":   token,
	})
}

// GetCurrentUser handles GET /api/v1/auth/me
func (h *UserAuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToPublicView(),
	})
}

// UpdateUserRequest represents the update user request
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// UpdateUser handles PUT /api/v1/auth/user
func (h *UserAuthHandler) UpdateUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Update fields if provided
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated successfully",
		"user":    user.ToPublicView(),
	})
}

// RegisterClientRequest represents the client registration request
type RegisterClientRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
}

// RegisterClient handles POST /api/v1/auth/register/client
func (h *UserAuthHandler) RegisterClient(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req RegisterClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Check if client name already exists
	existingClient, _ := h.clientRepo.FindByName(c.Request.Context(), req.Name)
	if existingClient != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "client name already exists"})
		return
	}

	// Create client
	client := &models.Client{
		Name:        req.Name,
		Description: req.Description,
		UserIDs:     []primitive.ObjectID{userObjID},
		APIKeys:     []models.APIKey{},
		IsActive:    true,
	}

	client, err = h.clientRepo.Create(c.Request.Context(), client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create client", "details": err.Error()})
		return
	}

	// Add client to user's client_ids
	if err := h.userRepo.AddClientToUser(c.Request.Context(), userObjID, client.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to associate client with user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "client registered successfully",
		"client":  client.ToPublicView(),
	})
}

// GenerateAPIKeyRequest represents the API key generation request
type GenerateAPIKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Permissions []string `json:"permissions,omitempty"`
	ExpiresAt   *string  `json:"expires_at,omitempty"`
}

// GenerateAPIKey handles POST /api/v1/auth/clients/:client_id/api-keys
func (h *UserAuthHandler) GenerateAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	clientIDStr := c.Param("client_id")
	clientID, err := primitive.ObjectIDFromHex(clientIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	// Verify user has access to this client
	client, err := h.verifyClientAccess(c, clientID, userObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}
	if client == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this client"})
		return
	}

	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Generate API key
	rawAPIKey, err := generateSecureAPIKey(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate API key"})
		return
	}

	// Hash the API key for storage
	apiKeyHash := hashAPIKey(rawAPIKey)

	// Get key prefix (first 8 characters)
	keyPrefix := rawAPIKey[:8]

	// Parse expiration if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format, use RFC3339"})
			return
		}
		expiresAt = &parsedTime
	}

	// Create API key entry
	apiKey := models.APIKey{
		ID:          primitive.NewObjectID(),
		Key:         apiKeyHash,
		Name:        req.Name,
		KeyPrefix:   keyPrefix,
		Permissions: req.Permissions,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
	}

	// Add API key to client
	if err := h.clientRepo.AddAPIKey(c.Request.Context(), clientID, apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add API key", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "API key generated successfully",
		"api_key": rawAPIKey, // Return the raw key only once
		"key_id":  apiKey.ID.Hex(),
		"prefix":  keyPrefix,
		"warning": "Save this API key now. You won't be able to see it again.",
	})
}

// RevokeAPIKey handles DELETE /api/v1/auth/clients/:client_id/api-keys/:key_id
func (h *UserAuthHandler) RevokeAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	clientIDStr := c.Param("client_id")
	clientID, err := primitive.ObjectIDFromHex(clientIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	keyIDStr := c.Param("key_id")
	keyID, err := primitive.ObjectIDFromHex(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key ID"})
		return
	}

	// Verify user has access to this client
	client, err := h.verifyClientAccess(c, clientID, userObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}
	if client == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this client"})
		return
	}

	// Revoke API key
	if err := h.clientRepo.RevokeAPIKey(c.Request.Context(), clientID, keyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke API key", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
	})
}

// GetClientDetails handles GET /api/v1/auth/clients/:client_id
func (h *UserAuthHandler) GetClientDetails(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	clientIDStr := c.Param("client_id")
	clientID, err := primitive.ObjectIDFromHex(clientIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client ID"})
		return
	}

	// Verify user has access to this client
	client, err := h.verifyClientAccess(c, clientID, userObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}
	if client == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this client"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client": client.ToPublicView(),
	})
}

// GetUserClients handles GET /api/v1/auth/clients
func (h *UserAuthHandler) GetUserClients(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	clients, err := h.clientRepo.FindByUserID(c.Request.Context(), userObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch clients", "details": err.Error()})
		return
	}

	clientViews := make([]map[string]interface{}, len(clients))
	for i, client := range clients {
		clientViews[i] = client.ToPublicView()
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clientViews,
	})
}

// Helper functions

func generateSecureAPIKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
