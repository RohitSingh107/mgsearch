package handlers

import (
	"mgsearch/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TasksHandler struct {
	meilisearchService *services.MeilisearchService
}

// NewTasksHandler creates a new tasks handler
func NewTasksHandler(meilisearchService *services.MeilisearchService) *TasksHandler {
	return &TasksHandler{
		meilisearchService: meilisearchService,
	}
}

// GetTask handles task details requests
// GET /api/v1/clients/:client_id/tasks/:task_id
// Returns task details from Meilisearch
func (h *TasksHandler) GetTask(c *gin.Context) {
	// Get client ID and task ID from URL parameters
	clientID := c.Param("client_id")
	taskID := c.Param("task_id")

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "client ID is required",
		})
		return
	}

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "task ID is required",
		})
		return
	}

	// Validate task ID is a valid number
	if _, err := strconv.Atoi(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "task ID must be a valid number",
		})
		return
	}

	// Get task details from Meilisearch
	taskResponse, err := h.meilisearchService.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get task details",
			"details": err.Error(),
		})
		return
	}

	// Return response from Meilisearch
	c.JSON(http.StatusOK, taskResponse)
}
