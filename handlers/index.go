package handlers

import (
	"fmt"
	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IndexHandler struct {
	clientRepo   *repositories.ClientRepository
	indexRepo    *repositories.IndexRepository
	meiliService *services.MeilisearchService
}

func NewIndexHandler(clientRepo *repositories.ClientRepository, indexRepo *repositories.IndexRepository, meiliService *services.MeilisearchService) *IndexHandler {
	return &IndexHandler{
		clientRepo:   clientRepo,
		indexRepo:    indexRepo,
		meiliService: meiliService,
	}
}

// CreateIndex creates a new index for a client
func (h *IndexHandler) CreateIndex(c *gin.Context) {
	clientIDParam := c.Param("client_id")
	clientID, err := primitive.ObjectIDFromHex(clientIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req models.CreateIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify client exists
	client, err := h.clientRepo.FindByID(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Check if index already exists in DB
	existing, _ := h.indexRepo.FindByNameAndClientID(c.Request.Context(), req.Name, clientID)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Index with this name already exists for this client"})
		return
	}

	// Generate UID
	// Format: client_name__index_name
	uid := fmt.Sprintf("%s__%s", client.Name, req.Name)

	// Create in Meilisearch
	task, err := h.meiliService.CreateIndex(uid, req.PrimaryKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create index in Meilisearch: %v", err)})
		return
	}

	// Save to DB
	index := &models.Index{
		ClientID:   clientID,
		Name:       req.Name,
		UID:        uid,
		PrimaryKey: req.PrimaryKey,
	}

	savedIndex, err := h.indexRepo.Create(c.Request.Context(), index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save index record: %v", err)})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"index": savedIndex,
		"task":  task,
	})
}

// GetClientIndexes returns all indexes for a client
func (h *IndexHandler) GetClientIndexes(c *gin.Context) {
	clientIDParam := c.Param("client_id")
	clientID, err := primitive.ObjectIDFromHex(clientIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	// Verify client exists
	_, err = h.clientRepo.FindByID(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	indexes, err := h.indexRepo.FindByClientID(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, indexes)
}
