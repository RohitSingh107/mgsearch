package handlers

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"mgsearch/config"
	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/pkg/security"
	"mgsearch/services"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	repo          *repositories.SessionRepository
	storeRepo     *repositories.StoreRepository
	meiliService  *services.MeilisearchService
	encryptionKey []byte
	cfg           *config.Config
}

func NewSessionHandler(repo *repositories.SessionRepository, storeRepo *repositories.StoreRepository, meiliService *services.MeilisearchService, cfg *config.Config) (*SessionHandler, error) {
	// Decode encryption key from hex
	key, err := security.MustDecodeKey(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}

	return &SessionHandler{
		repo:          repo,
		storeRepo:     storeRepo,
		meiliService:  meiliService,
		encryptionKey: key,
		cfg:           cfg,
	}, nil
}

// encryptAccessToken encrypts the access token before storage
func (h *SessionHandler) encryptAccessToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	encrypted, err := security.EncryptAESGCM(h.encryptionKey, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), nil
}

// decryptAccessToken decrypts the access token after retrieval
func (h *SessionHandler) decryptAccessToken(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	encrypted, err := hex.DecodeString(ciphertext)
	if err != nil {
		// If it's not hex, assume it's already plaintext (for backward compatibility)
		return ciphertext, nil
	}
	decrypted, err := security.DecryptAESGCM(h.encryptionKey, encrypted)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// StoreSession handles POST /api/sessions
// Stores a Shopify session in the backend database (upsert behavior)
func (h *SessionHandler) StoreSession(c *gin.Context) {
	var session models.Session
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Validate required fields
	if session.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required field: id",
			"code":  "VALIDATION_ERROR",
		})
		return
	}
	if session.Shop == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required field: shop",
			"code":  "VALIDATION_ERROR",
		})
		return
	}
	if session.State == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required field: state",
			"code":  "VALIDATION_ERROR",
		})
		return
	}
	if session.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required field: accessToken",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	// Save original plaintext token for store creation (before encryption)
	plaintextToken := session.AccessToken

	// Encrypt access token before storing in session
	encryptedToken, err := h.encryptAccessToken(session.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}
	session.AccessToken = encryptedToken

	// Set timestamps if not provided
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	session.UpdatedAt = time.Now().UTC()

	// Store the session
	if err := h.repo.CreateOrUpdate(c.Request.Context(), &session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Automatically create or update store when session is stored
	// Use the original plaintext token for store creation
	if err := h.createOrUpdateStoreFromSession(c.Request.Context(), session.Shop, plaintextToken); err != nil {
		// Log error but don't fail the session storage
		// Session was stored successfully, store creation is a bonus
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Session stored successfully",
			"warning": "Store creation/update had issues, but session was saved",
		})
		return
	}

	// Return 204 No Content or 200 OK with success message
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Session stored successfully, store created/updated",
	})
}

// createOrUpdateStoreFromSession creates or updates a store based on session data
func (h *SessionHandler) createOrUpdateStoreFromSession(ctx context.Context, shopDomain, accessToken string) error {
	// Normalize shop domain
	shopDomain = strings.ToLower(strings.TrimSpace(shopDomain))
	shopDomain = strings.TrimPrefix(shopDomain, "https://")
	shopDomain = strings.TrimPrefix(shopDomain, "http://")
	shopDomain = strings.TrimSuffix(shopDomain, "/")

	if !strings.HasSuffix(shopDomain, ".myshopify.com") {
		return nil // Skip if invalid shop domain
	}

	// Check if store already exists
	existingStore, err := h.storeRepo.GetByShopDomain(ctx, shopDomain)
	if err == nil && existingStore != nil {
		// Store exists, update access token if needed
		encryptedToken, err := security.EncryptAESGCM(h.encryptionKey, []byte(accessToken))
		if err != nil {
			return err
		}

		existingStore.EncryptedAccessToken = encryptedToken
		existingStore.UpdatedAt = time.Now().UTC()
		_, err = h.storeRepo.CreateOrUpdate(ctx, existingStore)
		return err
	}

	// Store doesn't exist, create new one
	// Generate API keys
	privateKey, err := security.GenerateAPIKey(32)
	if err != nil {
		return err
	}

	webhookSecret, err := security.GenerateAPIKey(32)
	if err != nil {
		return err
	}

	// Encrypt access token
	encryptedToken, err := security.EncryptAESGCM(h.encryptionKey, []byte(accessToken))
	if err != nil {
		return err
	}

	// Build index UID (same format as auth handler)
	slug := strings.ToLower(strings.ReplaceAll(strings.Split(shopDomain, ".")[0], "-", "_"))
	indexUID := slug + "_all_products"

	// Get Meilisearch configuration
	meiliURL := h.cfg.MeilisearchURL
	meiliKey := h.cfg.MeilisearchAPIKey

	var encryptedMeiliKey []byte
	if meiliKey != "" {
		encryptedMeiliKey, err = security.EncryptAESGCM(h.encryptionKey, []byte(meiliKey))
		if err != nil {
			return err
		}
	}

	// Determine shop name (use domain as default)
	shopName := shopDomain
	if strings.Contains(shopDomain, ".") {
		parts := strings.Split(shopDomain, ".")
		if len(parts) > 0 {
			shopName = strings.ReplaceAll(parts[0], "-", " ")
		}
	}

	// Create store
	store := &models.Store{
		ShopDomain:           shopDomain,
		ShopName:             shopName,
		EncryptedAccessToken: encryptedToken,
		APIKeyPrivate:        privateKey,
		ProductIndexUID:      indexUID,
		MeilisearchIndexUID:  indexUID,
		MeilisearchDocType:   "product",
		MeilisearchURL:       meiliURL,
		MeilisearchAPIKey:    encryptedMeiliKey,
		PlanLevel:            "free",
		Status:               "active",
		WebhookSecret:        webhookSecret,
		InstalledAt:          time.Now().UTC(),
		SyncState: map[string]interface{}{
			"status": "pending_initial_sync",
		},
	}

	dbStore, err := h.storeRepo.CreateOrUpdate(ctx, store)
	if err != nil {
		return err
	}

	// Ensure Meilisearch index exists
	if h.meiliService != nil && dbStore.IndexUID() != "" {
		if err := h.meiliService.EnsureIndex(dbStore.IndexUID()); err != nil {
			// Log but don't fail - index creation can be retried later
			return nil
		}
	}

	return nil
}

// LoadSession handles GET /api/sessions/:id
// Retrieves a session by ID
func (h *SessionHandler) LoadSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session id is required",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	session, err := h.repo.GetByID(c.Request.Context(), sessionID)
	if err != nil {
		// Check if it's a "not found" error
		if err.Error() == "session not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Decrypt access token before returning
	decryptedToken, err := h.decryptAccessToken(session.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}
	session.AccessToken = decryptedToken

	c.JSON(http.StatusOK, session)
}

// DeleteSession handles DELETE /api/sessions/:id
// Deletes a session by ID (idempotent)
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session id is required",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	err := h.repo.DeleteByID(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}

// DeleteMultipleSessions handles DELETE /api/sessions/batch
// Deletes multiple sessions at once (idempotent)
func (h *SessionHandler) DeleteMultipleSessions(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ids array cannot be empty",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	if err := h.repo.DeleteByIDs(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Return 204 No Content or 200 OK with success message
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"deleted": len(req.IDs),
		"message": "Sessions deleted successfully",
	})
}

// FindSessionsByShop handles GET /api/sessions/shop/:shop
// Retrieves all sessions for a specific shop
func (h *SessionHandler) FindSessionsByShop(c *gin.Context) {
	shop := c.Param("shop")
	if shop == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "shop parameter is required",
			"code":  "VALIDATION_ERROR",
		})
		return
	}

	sessions, err := h.repo.GetByShop(c.Request.Context(), shop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Decrypt access tokens for all sessions
	for _, session := range sessions {
		decryptedToken, err := h.decryptAccessToken(session.AccessToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
			return
		}
		session.AccessToken = decryptedToken
	}

	// Always return an array, even if empty
	if sessions == nil {
		sessions = []*models.Session{}
	}

	c.JSON(http.StatusOK, sessions)
}

