package handlers

import (
	"net/http"

	"mgsearch/middleware"
	"mgsearch/repositories"

	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	repo *repositories.StoreRepository
}

func NewStoreHandler(repo *repositories.StoreRepository) *StoreHandler {
	return &StoreHandler{repo: repo}
}

func (h *StoreHandler) GetCurrentStore(c *gin.Context) {
	storeID, ok := middleware.GetStoreID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	store, err := h.repo.GetByID(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, store.ToPublicView())
}

func (h *StoreHandler) GetSyncStatus(c *gin.Context) {
	storeID, ok := middleware.GetStoreID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	store, err := h.repo.GetByID(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store not found", "details": err.Error()})
		return
	}

	// Return sync status with minimal store info
	c.JSON(http.StatusOK, gin.H{
		"store_id":     store.ID,
		"shop_domain":  store.ShopDomain,
		"sync_state":   store.SyncState,
		"index_uid":    store.IndexUID(),
		"document_type": store.DocumentType(),
	})
}
