package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mgsearch/services"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTasksTest(t *testing.T) *gin.Engine {
	cfg := testhelpers.TestConfig()
	meiliService := services.NewMeilisearchService(cfg)

	tasksHandler := NewTasksHandler(meiliService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/clients/:client_name/tasks/:task_id", tasksHandler.GetTask)
	}

	return router
}

func TestTasksHandler_GetTask(t *testing.T) {
	router := setupTasksTest(t)

	tests := []struct {
		name           string
		clientName     string
		taskID         string
		expectedStatus int
	}{
		{
			name:           "valid task ID",
			clientName:     "testclient",
			taskID:         "123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing client name",
			clientName:     "",
			taskID:         "123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing task ID",
			clientName:     "testclient",
			taskID:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid task ID (not a number)",
			clientName:     "testclient",
			taskID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "task ID with letters",
			clientName:     "testclient",
			taskID:         "abc123",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/clients/" + tt.clientName + "/tasks/" + tt.taskID
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

