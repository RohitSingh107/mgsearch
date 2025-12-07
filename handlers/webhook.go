package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"mgsearch/models"
	"mgsearch/repositories"
	"mgsearch/services"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	shopify *services.ShopifyService
	stores  *repositories.StoreRepository
	meili   *services.MeilisearchService
}

func NewWebhookHandler(shopify *services.ShopifyService, stores *repositories.StoreRepository, meili *services.MeilisearchService) *WebhookHandler {
	return &WebhookHandler{
		shopify: shopify,
		stores:  stores,
		meili:   meili,
	}
}

func (h *WebhookHandler) HandleShopifyWebhook(c *gin.Context) {
	topic := c.Param("topic")
	subtopic := c.Param("subtopic")
	event := fmt.Sprintf("%s/%s", topic, subtopic)

	signature := c.GetHeader("X-Shopify-Hmac-Sha256")
	shopDomain := c.GetHeader("X-Shopify-Shop-Domain")

	if signature == "" || shopDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required headers"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read body"})
		return
	}

	if !h.shopify.VerifyWebhookSignature(signature, body) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook signature"})
		return
	}

	store, err := h.stores.GetByShopDomain(c.Request.Context(), shopDomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not registered"})
		return
	}

	indexUID := store.IndexUID()
	if indexUID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store index not configured"})
		return
	}

	switch event {
	case "products/create", "products/update":
		if err := h.handleProductUpsert(store, indexUID, body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update index", "details": err.Error()})
			return
		}
	case "products/delete":
		if err := h.handleProductDelete(store, indexUID, body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document", "details": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusOK, gin.H{"message": "event ignored"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed"})
}

func (h *WebhookHandler) handleProductUpsert(store *models.Store, indexUID string, payload []byte) error {
	var product map[string]interface{}
	if err := json.Unmarshal(payload, &product); err != nil {
		return err
	}

	if product["id"] == nil {
		return fmt.Errorf("product id missing")
	}

	document := models.Document(product)
	document["shop_domain"] = store.ShopDomain
	document["store_id"] = store.ID.Hex()
	document["document_type"] = store.DocumentType()

	_, err := h.meili.IndexDocument(indexUID, document)
	return err
}

func (h *WebhookHandler) handleProductDelete(store *models.Store, indexUID string, payload []byte) error {
	var product struct {
		ID interface{} `json:"id"`
	}
	if err := json.Unmarshal(payload, &product); err != nil {
		return err
	}

	if product.ID == nil {
		return fmt.Errorf("product id missing")
	}

	idStr := fmt.Sprintf("%v", product.ID)
	return h.meili.DeleteDocument(indexUID, idStr)
}
