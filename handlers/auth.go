package handlers

import (
	"net/http"
	"strings"
	"time"

	"mgsearch/config"
	"mgsearch/models"
	"mgsearch/pkg/auth"
	"mgsearch/pkg/security"
	"mgsearch/repositories"
	"mgsearch/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg           *config.Config
	shopify       *services.ShopifyService
	stores        *repositories.StoreRepository
	meili         *services.MeilisearchService
	encryptionKey []byte
	sessionTTL    time.Duration
}

type beginAuthRequest struct {
	Shop        string  `json:"shop" binding:"required"`
	RedirectURI *string `json:"redirect_uri,omitempty"`
}

type beginAuthResponse struct {
	AuthURL string `json:"authUrl"`
	State   string `json:"state"`
}

type installStoreRequest struct {
	Shop              string  `json:"shop" binding:"required"`
	AccessToken       string  `json:"access_token" binding:"required"`
	ShopName          *string `json:"shop_name,omitempty"`
	MeilisearchURL    *string `json:"meilisearch_url,omitempty"`
	MeilisearchAPIKey *string `json:"meilisearch_api_key,omitempty"`
}

type exchangeTokenRequest struct {
	Shop string `json:"shop" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type exchangeTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func NewAuthHandler(cfg *config.Config, shopify *services.ShopifyService, stores *repositories.StoreRepository, meili *services.MeilisearchService) (*AuthHandler, error) {
	key, err := security.MustDecodeKey(cfg.EncryptionKey)
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		cfg:           cfg,
		shopify:       shopify,
		stores:        stores,
		meili:         meili,
		encryptionKey: key,
		sessionTTL:    24 * time.Hour,
	}, nil
}

// Begin starts the OAuth flow by returning the Shopify authorization URL.
func (h *AuthHandler) Begin(c *gin.Context) {
	var req beginAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	shop := strings.ToLower(strings.TrimSpace(req.Shop))
	if shop == "" || !strings.HasSuffix(shop, ".myshopify.com") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop domain"})
		return
	}

	state, err := auth.GenerateStateToken(shop, []byte(h.cfg.JWTSigningKey), 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state token"})
		return
	}

	// Use redirect_uri from request if provided, otherwise fallback to default
	redirectURI := strings.TrimRight(h.cfg.ShopifyAppURL, "/") + "/auth/callback"
	if req.RedirectURI != nil && *req.RedirectURI != "" {
		redirectURI = *req.RedirectURI // Use EXACTLY as sent, don't modify
	}

	authURL, err := h.shopify.BuildInstallURL(shop, state, redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build install url"})
		return
	}

	c.JSON(http.StatusOK, beginAuthResponse{
		AuthURL: authURL,
		State:   state,
	})
}

// Callback handles the OAuth redirect from Shopify, exchanges the code, and stores the tenant.
func (h *AuthHandler) Callback(c *gin.Context) {
	values := c.Request.URL.Query()

	if !h.shopify.ValidateHMAC(values) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid hmac"})
		return
	}

	shop := values.Get("shop")
	code := values.Get("code")
	state := values.Get("state")

	if shop == "" || code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	stateShop, err := auth.ParseStateToken(state, []byte(h.cfg.JWTSigningKey))
	if err != nil || stateShop != shop {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid state parameter"})
		return
	}

	accessToken, err := h.shopify.ExchangeAccessToken(c.Request.Context(), shop, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed", "details": err.Error()})
		return
	}

	encryptedToken, err := security.EncryptAESGCM(h.encryptionKey, []byte(accessToken))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token encryption failed"})
		return
	}

	privateKey, err := security.GenerateAPIKey(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate private key"})
		return
	}

	webhookSecret, err := security.GenerateAPIKey(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate webhook secret"})
		return
	}

	meiliURL := strings.TrimSpace(c.GetHeader("X-Meilisearch-Url"))
	if meiliURL == "" {
		meiliURL = h.cfg.MeilisearchURL
	}
	if meiliURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "meilisearch url is required"})
		return
	}

	meiliKey := strings.TrimSpace(c.GetHeader("X-Meilisearch-Api-Key"))
	if meiliKey == "" {
		meiliKey = h.cfg.MeilisearchAPIKey
	}
	if meiliKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "meilisearch api key is required"})
		return
	}

	encryptedMeiliKey, err := security.EncryptAESGCM(h.encryptionKey, []byte(meiliKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure meilisearch api key"})
		return
	}

	indexUID := buildProductIndexUID(shop)
	docType := "product"

	store := &models.Store{
		ShopDomain:           shop,
		ShopName:             shop,
		EncryptedAccessToken: encryptedToken,
		APIKeyPrivate:        privateKey,
		ProductIndexUID:      indexUID,
		MeilisearchIndexUID:  indexUID,
		MeilisearchDocType:   docType,
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

	dbStore, err := h.stores.CreateOrUpdate(c.Request.Context(), store)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist store", "details": err.Error()})
		return
	}

	if err := h.meili.EnsureIndex(dbStore.IndexUID()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure search index", "details": err.Error()})
		return
	}

	sessionToken, err := auth.GenerateSessionToken(dbStore.ID.Hex(), dbStore.ShopDomain, []byte(h.cfg.JWTSigningKey), h.sessionTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"store":   dbStore.ToPublicView(),
		"token":   sessionToken,
		"message": "installation successful",
	})
}

// ExchangeToken handles POST /api/auth/shopify/exchange
// Exchanges OAuth code for access token (optional helper for frontend)
func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	var req exchangeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	shop := strings.ToLower(strings.TrimSpace(req.Shop))
	if shop == "" || !strings.HasSuffix(shop, ".myshopify.com") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop domain"})
		return
	}

	if req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	accessToken, err := h.shopify.ExchangeAccessToken(c.Request.Context(), shop, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed", "details": err.Error()})
		return
	}

	// Get scope from the response if available (we'd need to update ExchangeAccessToken to return it)
	// For now, just return the token
	c.JSON(http.StatusOK, exchangeTokenResponse{
		AccessToken: accessToken,
		Scope:       h.cfg.ShopifyScopes,
	})
}

// InstallStore handles POST /api/auth/shopify/install
// Receives OAuth data from frontend after frontend completes OAuth flow
// Frontend sends access_token directly (already exchanged from code)
func (h *AuthHandler) InstallStore(c *gin.Context) {
	var req installStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	shop := strings.ToLower(strings.TrimSpace(req.Shop))
	if shop == "" || !strings.HasSuffix(shop, ".myshopify.com") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop domain"})
		return
	}

	if req.AccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access_token is required"})
		return
	}

	// Encrypt the access token
	encryptedToken, err := security.EncryptAESGCM(h.encryptionKey, []byte(req.AccessToken))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token encryption failed"})
		return
	}

	privateKey, err := security.GenerateAPIKey(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate private key"})
		return
	}

	webhookSecret, err := security.GenerateAPIKey(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate webhook secret"})
		return
	}

	// Handle Meilisearch configuration
	meiliURL := strings.TrimSpace(c.GetHeader("X-Meilisearch-Url"))
	if meiliURL == "" && req.MeilisearchURL != nil {
		meiliURL = strings.TrimSpace(*req.MeilisearchURL)
	}
	if meiliURL == "" {
		meiliURL = h.cfg.MeilisearchURL
	}
	if meiliURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "meilisearch url is required"})
		return
	}

	meiliKey := strings.TrimSpace(c.GetHeader("X-Meilisearch-Api-Key"))
	if meiliKey == "" && req.MeilisearchAPIKey != nil {
		meiliKey = strings.TrimSpace(*req.MeilisearchAPIKey)
	}
	if meiliKey == "" {
		meiliKey = h.cfg.MeilisearchAPIKey
	}
	if meiliKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "meilisearch api key is required"})
		return
	}

	encryptedMeiliKey, err := security.EncryptAESGCM(h.encryptionKey, []byte(meiliKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure meilisearch api key"})
		return
	}

	// Determine shop name
	shopName := shop
	if req.ShopName != nil && *req.ShopName != "" {
		shopName = *req.ShopName
	}

	indexUID := buildProductIndexUID(shop)
	docType := "product"

	store := &models.Store{
		ShopDomain:           shop,
		ShopName:             shopName,
		EncryptedAccessToken: encryptedToken,
		APIKeyPrivate:        privateKey,
		ProductIndexUID:      indexUID,
		MeilisearchIndexUID:  indexUID,
		MeilisearchDocType:   docType,
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

	dbStore, err := h.stores.CreateOrUpdate(c.Request.Context(), store)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist store", "details": err.Error()})
		return
	}

	// Ensure the Meilisearch index exists
	if err := h.meili.EnsureIndex(dbStore.IndexUID()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ensure search index", "details": err.Error()})
		return
	}

	// Generate session token for frontend
	sessionToken, err := auth.GenerateSessionToken(dbStore.ID.Hex(), dbStore.ShopDomain, []byte(h.cfg.JWTSigningKey), h.sessionTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"store":   dbStore.ToPublicView(),
		"token":   sessionToken,
		"message": "installation successful",
	})
}

func buildProductIndexUID(shop string) string {
	slug := strings.ToLower(strings.ReplaceAll(strings.Split(shop, ".")[0], "-", "_"))
	return slug + "_all_products"
}
