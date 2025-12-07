package handlers

import (
	"mgsearch/models"
	"mgsearch/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	meilisearchService *services.MeilisearchService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(meilisearchService *services.MeilisearchService) *SettingsHandler {
	return &SettingsHandler{
		meilisearchService: meilisearchService,
	}
}

// UpdateSettings handles index settings update requests
// PATCH /api/v1/clients/:client_name/:index_name/settings
// Body: Any valid Meilisearch settings update request (can be multi-level nested JSON)
// Examples include: rankingRules, distinctAttribute, searchableAttributes, displayedAttributes,
// stopWords, sortableAttributes, synonyms, typoTolerance, pagination, faceting, searchCutoffMs
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
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
	var settingsRequest models.SettingsRequest
	if err := c.ShouldBindJSON(&settingsRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update settings (pass through any request body structure to Meilisearch)
	settingsResponse, err := h.meilisearchService.UpdateSettings(indexName, &settingsRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update settings",
			"details": err.Error(),
		})
		return
	}

	// Return response from Meilisearch
	c.JSON(http.StatusOK, settingsResponse)
}
