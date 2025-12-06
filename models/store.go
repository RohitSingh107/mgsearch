package models

import (
	"time"
)

// Store represents a tenant (Shopify merchant) onboarded into the system.
type Store struct {
	ID                   string                 `json:"id" db:"id"`
	ShopDomain           string                 `json:"shop_domain" db:"shop_domain"`
	ShopName             string                 `json:"shop_name" db:"shop_name"`
	EncryptedAccessToken []byte                 `json:"-" db:"encrypted_access_token"`
	APIKeyPublic         string                 `json:"api_key_public" db:"api_key_public"`
	APIKeyPrivate        string                 `json:"-" db:"api_key_private"`
	ProductIndexUID      string                 `json:"product_index_uid" db:"product_index_uid"`
	MeilisearchIndexUID  string                 `json:"meilisearch_index_uid" db:"meilisearch_index_uid"`
	MeilisearchDocType   string                 `json:"meilisearch_document_type" db:"meilisearch_document_type"`
	MeilisearchURL       string                 `json:"meilisearch_url" db:"meilisearch_url"`
	MeilisearchAPIKey    []byte                 `json:"-" db:"meilisearch_api_key"`
	PlanLevel            string                 `json:"plan_level" db:"plan_level"`
	Status               string                 `json:"status" db:"status"`
	WebhookSecret        string                 `json:"-" db:"webhook_secret"`
	InstalledAt          time.Time              `json:"installed_at" db:"installed_at"`
	UninstalledAt        *time.Time             `json:"uninstalled_at,omitempty" db:"uninstalled_at"`
	SyncState            map[string]interface{} `json:"sync_state" db:"sync_state"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
}

// StorePublicView represents the subset of store fields surfaced to authenticated dashboards.
type StorePublicView struct {
	ID              string                 `json:"id"`
	ShopDomain      string                 `json:"shop_domain"`
	ShopName        string                 `json:"shop_name"`
	PlanLevel       string                 `json:"plan_level"`
	Status          string                 `json:"status"`
	ProductIndexUID string                 `json:"product_index_uid"`
	IndexUID        string                 `json:"meilisearch_index_uid"`
	MeilisearchURL  string                 `json:"meilisearch_url"`
	DocumentType    string                 `json:"meilisearch_document_type"`
	APIKeyPublic    string                 `json:"api_key_public,omitempty"` // Storefront key for search API
	SyncState       map[string]interface{} `json:"sync_state"`
	InstalledAt     time.Time              `json:"installed_at"`
}

// ToPublicView converts a Store to its dashboard-friendly representation.
func (s *Store) ToPublicView() StorePublicView {
	return StorePublicView{
		ID:              s.ID,
		ShopDomain:      s.ShopDomain,
		ShopName:        s.ShopName,
		PlanLevel:       s.PlanLevel,
		Status:          s.Status,
		ProductIndexUID: s.ProductIndexUID,
		IndexUID:        s.MeilisearchIndexUID,
		MeilisearchURL:  s.MeilisearchURL,
		DocumentType:    s.MeilisearchDocType,
		APIKeyPublic:    s.APIKeyPublic, // Include storefront key
		SyncState:       s.SyncState,
		InstalledAt:     s.InstalledAt,
	}
}

// IndexUID returns the effective Meilisearch index identifier for the store.
func (s *Store) IndexUID() string {
	if s.MeilisearchIndexUID != "" {
		return s.MeilisearchIndexUID
	}
	return s.ProductIndexUID
}

// DocumentType returns the type label used for Meilisearch documents.
func (s *Store) DocumentType() string {
	if s.MeilisearchDocType != "" {
		return s.MeilisearchDocType
	}
	return "product"
}
