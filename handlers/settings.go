package handlers

import (
	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SettingsHandler struct {
	meilisearchService *services.MeilisearchService
	clientRepo         *repositories.ClientRepository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(meilisearchService *services.MeilisearchService, clientRepo *repositories.ClientRepository) *SettingsHandler {
	return &SettingsHandler{
		meilisearchService: meilisearchService,
		clientRepo:         clientRepo,
	}
}

// UpdateSettings handles index settings update requests
// PATCH /api/v1/clients/:client_id/indexes/:index_name/settings
// Body: Any valid Meilisearch settings update request (can be multi-level nested JSON)
// Examples include: rankingRules, distinctAttribute, searchableAttributes, displayedAttributes,
// stopWords, sortableAttributes, synonyms, typoTolerance, pagination, faceting, searchCutoffMs
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	// Get client name from context (set by APIKeyMiddleware)
	clientName := c.GetString("client_name")

	// If not found in context (e.g. JWT auth), we need to look it up
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
	indexName := c.Param("index_name")

	if clientName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "client context not found",
		})
		return
	}

	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "index name is required",
		})
		return
	}

	// Construct the actual Meilisearch index UID
	meiliIndexUID := clientName + "__" + indexName

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
	settingsResponse, err := h.meilisearchService.UpdateSettings(meiliIndexUID, &settingsRequest)
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
