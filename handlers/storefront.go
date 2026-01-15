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

func (h *StorefrontHandler) Similar(c *gin.Context) {
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

	var productID interface{}
	var limit int

	// Support both GET (query param 'id') and POST (JSON body {"id": ...})
	if c.Request.Method == "POST" {
		var body struct {
			ID    interface{} `json:"id"`
			Limit int         `json:"limit,omitempty"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
			return
		}
		productID = body.ID
		limit = body.Limit
	} else {
		productIDStr := c.Query("id")
		if productIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing product id"})
			return
		}
		// Try to parse as int, if fails keep as string
		if id, err := strconv.Atoi(productIDStr); err == nil {
			productID = id
		} else {
			productID = productIDStr
		}
		// Parse limit from query params
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
				limit = parsedLimit
			}
		}
	}

	if productID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
		return
	}

	collectionName := store.QdrantCollection()
	if collectionName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "qdrant collection not configured for store"})
		return
	}

	// Use default limit of 10 if not specified
	if limit <= 0 {
		limit = 10
	}

	resp, err := h.qdrant.Recommend(collectionName, []interface{}{productID}, limit)
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

	c.JSON(http.StatusOK, resp)
}
