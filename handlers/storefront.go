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

	collectionName := c.Query("collection")
	if collectionName == "" {
		collectionName = store.QdrantCollection()
	}

	if collectionName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "qdrant collection not configured for store"})
		return
	}

	// Read request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body", "details": err.Error()})
		return
	}

	// Parse body to force with_payload: true
	var bodyMap map[string]interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body", "details": err.Error()})
			return
		}
	} else {
		bodyMap = make(map[string]interface{})
	}

	// Force payload return
	bodyMap["with_payload"] = true

	// Clean up query parameters (remove nulls)
	if query, ok := bodyMap["query"].(map[string]interface{}); ok {
		if recommend, ok := query["recommend"].(map[string]interface{}); ok {
			// Clean positive array
			if positive, ok := recommend["positive"].([]interface{}); ok {
				cleanedPositive := make([]interface{}, 0, len(positive))
				for _, v := range positive {
					if v != nil {
						cleanedPositive = append(cleanedPositive, v)
					}
				}
				if len(cleanedPositive) == 0 {
					// If no positive IDs, return empty result without querying Qdrant
					c.JSON(http.StatusOK, gin.H{"result": []interface{}{}, "status": "ok"})
					return
				}
				recommend["positive"] = cleanedPositive
			}
			// Clean negative array
			if negative, ok := recommend["negative"].([]interface{}); ok {
				cleanedNegative := make([]interface{}, 0, len(negative))
				for _, v := range negative {
					if v != nil {
						cleanedNegative = append(cleanedNegative, v)
					}
				}
				recommend["negative"] = cleanedNegative
			}
		}
	}

	// Re-marshal
	newBody, err := json.Marshal(bodyMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare request", "details": err.Error()})
		return
	}

	resp, err := h.qdrant.ProxyQuery(collectionName, newBody)
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
