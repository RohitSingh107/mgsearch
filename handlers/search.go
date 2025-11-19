package handlers

import (
	"mgsearch/models"
	"mgsearch/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	meilisearchService *services.MeilisearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(meilisearchService *services.MeilisearchService) *SearchHandler {
	return &SearchHandler{
		meilisearchService: meilisearchService,
	}
}

// Search handles search requests
// POST /api/v1/clients/:client_name/:index_name/search
// Body: Any valid Meilisearch search request (can be multi-level nested JSON)
// Examples:
//   - { "q": "query" }
//   - { "q": "query", "filter": "genre = action", "sort": ["release_date:desc"] }
//   - { "q": "query", "filter": ["genre = action", "year > 2020"], "facets": ["genre", "year"] }
func (h *SearchHandler) Search(c *gin.Context) {
	// Get client name and index name from URL parameters
	clientName := c.Param("client_name")
	indexName := c.Param("index_name")

	if clientName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "client name is required",
		})
		return
	}

	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index name is required",
		})
		return
	}

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
	searchResponse, err := h.meilisearchService.Search(indexName, &searchRequest)
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
