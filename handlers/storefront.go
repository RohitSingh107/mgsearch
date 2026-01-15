package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"

	"github.com/gin-gonic/gin"
)

type StorefrontHandler struct {
	stores *repositories.StoreRepository
	meili  *services.MeilisearchService
	qdrant *services.QdrantService
}

func NewStorefrontHandler(stores *repositories.StoreRepository, meili *services.MeilisearchService, qdrant *services.QdrantService) *StorefrontHandler {
	return &StorefrontHandler{
		stores: stores,
		meili:  meili,
		qdrant: qdrant,
	}
}

func (h *StorefrontHandler) Search(c *gin.Context) {
	// Handle preflight OPTIONS request
	if c.Request.Method == "OPTIONS" {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, ngrok-skip-browser-warning")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200")
		c.Status(http.StatusNoContent)
		return
	}

	var body models.SearchRequest

	// Support both GET (query params) and POST (JSON body)
	if c.Request.Method == "POST" {
		// POST: Parse JSON body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
			return
		}
	} else {
		// GET: Parse query parameters
		query := c.DefaultQuery("q", c.DefaultQuery("query", ""))
		limitStr := c.DefaultQuery("limit", "20")
		offsetStr := c.DefaultQuery("offset", "0")
		sortParam := c.QueryArray("sort")
		filterParam := c.Query("filters")

		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			limit = 20
		}
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			offset = 0
		}

		body = models.SearchRequest{
			"q":      query,
			"limit":  limit,
			"offset": offset,
		}

		if len(sortParam) > 0 {
			body["sort"] = sortParam
		}

		if filterParam != "" {
			var filter interface{}
			if err := json.Unmarshal([]byte(filterParam), &filter); err == nil {
				body["filter"] = filter
			}
		}
	}

	// Identify Store
	var store *models.Store
	var err error

	// Require shop domain
	shopDomain := c.Query("shop")
	if shopDomain == "" {
		// Try finding shop in body (for POST requests)
		if s, ok := body["shop"].(string); ok {
			shopDomain = s
		}
	}

	if shopDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop domain"})
		return
	}
	store, err = h.stores.GetByShopDomain(c.Request.Context(), shopDomain)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	indexUID := store.IndexUID()
	if indexUID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store index not configured"})
		return
	}

	// Ensure 'q' field exists (required by Meilisearch, can be empty string)
	if _, ok := body["q"]; !ok {
		body["q"] = ""
	}

	// Remove internal fields
	delete(body, "shop")

	resp, err := h.meili.Search(indexUID, &body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed", "details": err.Error()})
		return
	}

	// Add CORS headers explicitly
	origin := c.GetHeader("Origin")
	if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	c.JSON(http.StatusOK, resp)
}

func (h *StorefrontHandler) Similar(c *gin.Context) {
	// Handle preflight OPTIONS request
	if c.Request.Method == "OPTIONS" {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, ngrok-skip-browser-warning")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200")
		c.Status(http.StatusNoContent)
		return
	}

	shopDomain := c.Query("shop")
	if shopDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing shop parameter"})
		return
	}

	store, err := h.stores.GetByShopDomain(c.Request.Context(), shopDomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	collectionName := store.QdrantCollection()
	if collectionName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "qdrant collection not configured for store"})
		return
	}

	// Read raw request body to proxy it
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body", "details": err.Error()})
		return
	}

	resp, err := h.qdrant.ProxyQuery(collectionName, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch similar products", "details": err.Error()})
		return
	}

	// Add CORS headers explicitly
	origin := c.GetHeader("Origin")
	if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	c.Data(http.StatusOK, "application/json", resp)
}
