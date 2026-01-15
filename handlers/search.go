package handlers

import (
	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchHandler struct {
	meilisearchService *services.MeilisearchService
	clientRepo         *repositories.ClientRepository
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(meilisearchService *services.MeilisearchService, clientRepo *repositories.ClientRepository) *SearchHandler {
	return &SearchHandler{
		meilisearchService: meilisearchService,
		clientRepo:         clientRepo,
	}
}

// Search handles search requests
// POST /api/v1/clients/:client_id/indexes/:index_name/search
// Body: Any valid Meilisearch search request (can be multi-level nested JSON)
// Examples:
//   - { "q": "query" }
//   - { "q": "query", "filter": "genre = action", "sort": ["release_date:desc"] }
//   - { "q": "query", "filter": ["genre = action", "year > 2020"], "facets": ["genre", "year"] }
func (h *SearchHandler) Search(c *gin.Context) {
	// Get client name from context (set by APIKeyMiddleware)
	clientName := c.GetString("client_name")

	// If client_name is not in context (e.g. JWT auth), fetch it from DB using client_id param
	if clientName == "" {
		clientIDParam := c.Param("client_id")
		if clientIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client ID is required"})
			return
		}

		clientID, err := primitive.ObjectIDFromHex(clientIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
			return
		}

		client, err := h.clientRepo.FindByID(c.Request.Context(), clientID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}

		// Verify that the user has access to this client (if using JWT)
		if userID, ok := c.Get("user_id"); ok {
			userIDStr, _ := userID.(string)
			userIDObj, err := primitive.ObjectIDFromHex(userIDStr)
			if err == nil {
				hasAccess := false
				for _, uid := range client.UserIDs {
					if uid == userIDObj {
						hasAccess = true
						break
					}
				}
				if !hasAccess {
					c.JSON(http.StatusForbidden, gin.H{"error": "User does not have access to this client"})
					return
				}
			}
		}

		clientName = client.Name
	}

	// Get index name from URL parameters
	indexName := strings.TrimSpace(c.Param("index_name"))

	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index name is required",
		})
		return
	}

	// Construct the actual Meilisearch index UID
	// Format: client_name__index_name
	meiliIndexUID := clientName + "__" + indexName

	// Parse request body as flexible JSON structure (supports nested JSON)
	var searchRequest models.SearchRequest
	if err := c.ShouldBindJSON(&searchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Perform search (pass through any request body structure to Meilisearch)
	searchResponse, err := h.meilisearchService.Search(meiliIndexUID, &searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to perform search",
			"details": err.Error(),
		})
		return
	}

	// Return response from Meilisearch
	c.JSON(http.StatusOK, searchResponse)
}

// IndexDocument handles document indexing requests
// POST /api/v1/clients/:client_id/indexes/:index_name/documents
// Body: A single document object that will be sent to Meilisearch
func (h *SearchHandler) IndexDocument(c *gin.Context) {
	// Get client name from context (set by APIKeyMiddleware)
	clientName := c.GetString("client_name")

	// If client_name is not in context (e.g. JWT auth), fetch it from DB using client_id param
	if clientName == "" {
		clientIDParam := c.Param("client_id")
		if clientIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "client ID is required"})
			return
		}

		clientID, err := primitive.ObjectIDFromHex(clientIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
			return
		}

		client, err := h.clientRepo.FindByID(c.Request.Context(), clientID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}

		// Verify that the user has access to this client (if using JWT)
		if userID, ok := c.Get("user_id"); ok {
			userIDStr, _ := userID.(string)
			userIDObj, err := primitive.ObjectIDFromHex(userIDStr)
			if err == nil {
				hasAccess := false
				for _, uid := range client.UserIDs {
					if uid == userIDObj {
						hasAccess = true
						break
					}
				}
				if !hasAccess {
					c.JSON(http.StatusForbidden, gin.H{"error": "User does not have access to this client"})
					return
				}
			}
		}

		clientName = client.Name
	}

	indexName := strings.TrimSpace(c.Param("index_name"))

	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index name is required",
		})
		return
	}

	// Construct the actual Meilisearch index UID
	meiliIndexUID := clientName + "__" + indexName

	var document models.Document
	if err := c.ShouldBindJSON(&document); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid document body",
			"details": err.Error(),
		})
		return
	}

	if len(document) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "document cannot be empty",
		})
		return
	}

	indexResponse, err := h.meilisearchService.IndexDocument(meiliIndexUID, document)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to index document",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, indexResponse)
}
