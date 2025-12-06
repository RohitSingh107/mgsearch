package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mgsearch/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoreRepository struct {
	pool *pgxpool.Pool
}

func NewStoreRepository(pool *pgxpool.Pool) *StoreRepository {
	return &StoreRepository{pool: pool}
}

func (r *StoreRepository) scanStore(row pgx.Row) (*models.Store, error) {
	var syncStateRaw []byte
	store := &models.Store{}

	err := row.Scan(
		&store.ID,
		&store.ShopDomain,
		&store.ShopName,
		&store.EncryptedAccessToken,
		&store.APIKeyPublic,
		&store.APIKeyPrivate,
		&store.ProductIndexUID,
		&store.MeilisearchIndexUID,
		&store.MeilisearchDocType,
		&store.MeilisearchURL,
		&store.MeilisearchAPIKey,
		&store.PlanLevel,
		&store.Status,
		&store.WebhookSecret,
		&store.InstalledAt,
		&store.UninstalledAt,
		&syncStateRaw,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(syncStateRaw) > 0 {
		var state map[string]interface{}
		if err := json.Unmarshal(syncStateRaw, &state); err == nil {
			store.SyncState = state
		}
	} else {
		store.SyncState = map[string]interface{}{}
	}

	return store, nil
}

func (r *StoreRepository) CreateOrUpdate(ctx context.Context, store *models.Store) (*models.Store, error) {
	if store.SyncState == nil {
		store.SyncState = map[string]interface{}{}
	}
	syncStateJSON, err := json.Marshal(store.SyncState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync state: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO stores (
			shop_domain,
			shop_name,
			encrypted_access_token,
			api_key_public,
			api_key_private,
			product_index_uid,
			meilisearch_index_uid,
			meilisearch_document_type,
			meilisearch_url,
			meilisearch_api_key,
			plan_level,
			status,
			webhook_secret,
			installed_at,
			sync_state
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (shop_domain) DO UPDATE SET
			shop_name = EXCLUDED.shop_name,
			encrypted_access_token = EXCLUDED.encrypted_access_token,
			api_key_public = EXCLUDED.api_key_public,
			api_key_private = EXCLUDED.api_key_private,
			product_index_uid = EXCLUDED.product_index_uid,
			meilisearch_index_uid = EXCLUDED.meilisearch_index_uid,
			meilisearch_document_type = EXCLUDED.meilisearch_document_type,
			meilisearch_url = EXCLUDED.meilisearch_url,
			meilisearch_api_key = EXCLUDED.meilisearch_api_key,
			plan_level = EXCLUDED.plan_level,
			status = 'active',
			webhook_secret = EXCLUDED.webhook_secret,
			installed_at = EXCLUDED.installed_at,
			sync_state = EXCLUDED.sync_state,
			updated_at = NOW()
		RETURNING
			id, shop_domain, shop_name, encrypted_access_token, api_key_public,
			api_key_private, product_index_uid, meilisearch_index_uid, meilisearch_document_type,
			meilisearch_url, meilisearch_api_key,
			plan_level, status, webhook_secret,
			installed_at, uninstalled_at, sync_state, created_at, updated_at
	`, store.ShopDomain, store.ShopName, store.EncryptedAccessToken, store.APIKeyPublic,
		store.APIKeyPrivate, store.ProductIndexUID, store.MeilisearchIndexUID, store.MeilisearchDocType,
		store.MeilisearchURL, store.MeilisearchAPIKey,
		store.PlanLevel, store.Status,
		store.WebhookSecret, store.InstalledAt, syncStateJSON,
	)

	return r.scanStore(row)
}

func (r *StoreRepository) GetByShopDomain(ctx context.Context, domain string) (*models.Store, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, shop_domain, shop_name, encrypted_access_token, api_key_public,
		       api_key_private, product_index_uid, meilisearch_index_uid, meilisearch_document_type,
		       meilisearch_url, meilisearch_api_key,
		       plan_level, status, webhook_secret,
		       installed_at, uninstalled_at, sync_state, created_at, updated_at
		FROM stores WHERE shop_domain = $1
	`, domain)

	return r.scanStore(row)
}

func (r *StoreRepository) GetByPublicAPIKey(ctx context.Context, key string) (*models.Store, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, shop_domain, shop_name, encrypted_access_token, api_key_public,
		       api_key_private, product_index_uid, meilisearch_index_uid, meilisearch_document_type,
		       meilisearch_url, meilisearch_api_key,
		       plan_level, status, webhook_secret,
		       installed_at, uninstalled_at, sync_state, created_at, updated_at
		FROM stores WHERE api_key_public = $1
	`, key)

	return r.scanStore(row)
}

func (r *StoreRepository) GetByID(ctx context.Context, id string) (*models.Store, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, shop_domain, shop_name, encrypted_access_token, api_key_public,
		       api_key_private, product_index_uid, meilisearch_index_uid, meilisearch_document_type,
		       meilisearch_url, meilisearch_api_key,
		       plan_level, status, webhook_secret,
		       installed_at, uninstalled_at, sync_state, created_at, updated_at
		FROM stores WHERE id = $1
	`, id)

	return r.scanStore(row)
}

func (r *StoreRepository) UpdateSyncState(ctx context.Context, storeID string, state map[string]interface{}) error {
	if state == nil {
		state = map[string]interface{}{}
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal sync state: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE stores SET sync_state = $1, updated_at = $2 WHERE id = $3
	`, payload, time.Now().UTC(), storeID)
	return err
}
