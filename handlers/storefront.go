package handlers

import (
	"encoding/json"
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
}

func NewStorefrontHandler(stores *repositories.StoreRepository, meili *services.MeilisearchService) *StorefrontHandler {
	return &StorefrontHandler{
		stores: stores,
		meili:  meili,
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
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Storefront-Key, Authorization, ngrok-skip-browser-warning")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "43200")
		c.Status(http.StatusNoContent)
		return
	}

	publicKey := c.GetHeader("X-Storefront-Key")
	if publicKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing storefront key"})
		return
	}

	store, err := h.stores.GetByPublicAPIKey(c.Request.Context(), publicKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid storefront key"})
		return
	}

	indexUID := store.IndexUID()
	if indexUID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store index not configured"})
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

	// Ensure 'q' field exists (required by Meilisearch, can be empty string)
	if _, ok := body["q"]; !ok {
		body["q"] = ""
	}

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
