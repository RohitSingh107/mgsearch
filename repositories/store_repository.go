package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mgsearch/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StoreRepository struct {
	collection *mongo.Collection
}

func NewStoreRepository(db *mongo.Database) *StoreRepository {
	return &StoreRepository{collection: db.Collection("stores")}
}

func (r *StoreRepository) CreateOrUpdate(ctx context.Context, store *models.Store) (*models.Store, error) {
	if store.SyncState == nil {
		store.SyncState = map[string]interface{}{}
	}

	now := time.Now()
	if store.CreatedAt.IsZero() {
		store.CreatedAt = now
	}
	store.UpdatedAt = now

	// Set defaults if not provided
	if store.MeilisearchDocType == "" {
		store.MeilisearchDocType = "product"
	}
	if store.PlanLevel == "" {
		store.PlanLevel = "free"
	}
	if store.Status == "" {
		store.Status = "active"
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{"shop_domain": store.ShopDomain}
	update := bson.M{
		"$set": bson.M{
			"shop_name":                store.ShopName,
			"encrypted_access_token":   store.EncryptedAccessToken,
			"api_key_private":          store.APIKeyPrivate,
			"product_index_uid":        store.ProductIndexUID,
			"meilisearch_index_uid":    store.MeilisearchIndexUID,
			"meilisearch_document_type": store.MeilisearchDocType,
			"meilisearch_url":          store.MeilisearchURL,
			"meilisearch_api_key":      store.MeilisearchAPIKey,
			"plan_level":              store.PlanLevel,
			"status":                  "active",
			"webhook_secret":          store.WebhookSecret,
			"installed_at":            store.InstalledAt,
			"sync_state":              store.SyncState,
			"updated_at":              store.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": store.CreatedAt,
		},
	}

	var result models.Store
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// If upsert didn't return a document, insert it manually
			if store.ID.IsZero() {
				store.ID = primitive.NewObjectID()
			}
			_, err = r.collection.InsertOne(ctx, store)
			if err != nil {
				return nil, fmt.Errorf("failed to insert store: %w", err)
			}
			return store, nil
		}
		return nil, fmt.Errorf("failed to create or update store: %w", err)
	}

	return &result, nil
}

func (r *StoreRepository) GetByShopDomain(ctx context.Context, domain string) (*models.Store, error) {
	var store models.Store
	err := r.collection.FindOne(ctx, bson.M{"shop_domain": domain}).Decode(&store)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("store not found")
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepository) GetByID(ctx context.Context, id string) (*models.Store, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid store ID: %w", err)
	}

	var store models.Store
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&store)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("store not found")
		}
		return nil, err
	}
	return &store, nil
}

func (r *StoreRepository) UpdateSyncState(ctx context.Context, storeID string, state map[string]interface{}) error {
	if state == nil {
		state = map[string]interface{}{}
	}

	objectID, err := primitive.ObjectIDFromHex(storeID)
	if err != nil {
		return fmt.Errorf("invalid store ID: %w", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"sync_state": state,
			"updated_at": time.Now().UTC(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}
