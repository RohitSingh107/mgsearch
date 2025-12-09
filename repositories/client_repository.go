package repositories

import (
	"context"
	"errors"
	"time"

	"mgsearch/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientRepository struct {
	collection *mongo.Collection
}

func NewClientRepository(db *mongo.Database) *ClientRepository {
	return &ClientRepository{
		collection: db.Collection("clients"),
	}
}

// Create creates a new client
func (r *ClientRepository) Create(ctx context.Context, client *models.Client) (*models.Client, error) {
	client.CreatedAt = time.Now().UTC()
	client.UpdatedAt = time.Now().UTC()

	result, err := r.collection.InsertOne(ctx, client)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("client name already exists")
		}
		return nil, err
	}

	client.ID = result.InsertedID.(primitive.ObjectID)
	return client, nil
}

// FindByID finds a client by ID
func (r *ClientRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Client, error) {
	var client models.Client
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	return &client, nil
}

// FindByName finds a client by name
func (r *ClientRepository) FindByName(ctx context.Context, name string) (*models.Client, error) {
	var client models.Client
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	return &client, nil
}

// FindByAPIKey finds a client by API key hash
func (r *ClientRepository) FindByAPIKey(ctx context.Context, apiKeyHash string) (*models.Client, error) {
	var client models.Client
	filter := bson.M{
		"api_keys": bson.M{
			"$elemMatch": bson.M{
				"key":       apiKeyHash,
				"is_active": true,
			},
		},
		"is_active": true,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&client)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid API key")
		}
		return nil, err
	}
	return &client, nil
}

// Update updates a client
func (r *ClientRepository) Update(ctx context.Context, client *models.Client) error {
	client.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": client.ID}
	update := bson.M{
		"$set": bson.M{
			"name":        client.Name,
			"description": client.Description,
			"is_active":   client.IsActive,
			"user_ids":    client.UserIDs,
			"updated_at":  client.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("client not found")
	}

	return nil
}

// AddAPIKey adds a new API key to the client
func (r *ClientRepository) AddAPIKey(ctx context.Context, clientID primitive.ObjectID, apiKey models.APIKey) error {
	filter := bson.M{"_id": clientID}
	update := bson.M{
		"$push": bson.M{"api_keys": apiKey},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("client not found")
	}

	return nil
}

// UpdateAPIKeyLastUsed updates the last_used_at timestamp for an API key
func (r *ClientRepository) UpdateAPIKeyLastUsed(ctx context.Context, clientID, apiKeyID primitive.ObjectID) error {
	now := time.Now().UTC()
	filter := bson.M{
		"_id":          clientID,
		"api_keys._id": apiKeyID,
	}
	update := bson.M{
		"$set": bson.M{
			"api_keys.$.last_used_at": now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// RevokeAPIKey deactivates an API key
func (r *ClientRepository) RevokeAPIKey(ctx context.Context, clientID, apiKeyID primitive.ObjectID) error {
	filter := bson.M{
		"_id":          clientID,
		"api_keys._id": apiKeyID,
	}
	update := bson.M{
		"$set": bson.M{
			"api_keys.$.is_active": false,
			"updated_at":           time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("client or API key not found")
	}

	return nil
}

// AddUserToClient adds a user ID to client's user_ids array
func (r *ClientRepository) AddUserToClient(ctx context.Context, clientID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": clientID}
	update := bson.M{
		"$addToSet": bson.M{"user_ids": userID},
		"$set":      bson.M{"updated_at": time.Now().UTC()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("client not found")
	}

	return nil
}

// RemoveUserFromClient removes a user ID from client's user_ids array
func (r *ClientRepository) RemoveUserFromClient(ctx context.Context, clientID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": clientID}
	update := bson.M{
		"$pull": bson.M{"user_ids": userID},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("client not found")
	}

	return nil
}

// FindByUserID finds all clients associated with a user
func (r *ClientRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Client, error) {
	filter := bson.M{"user_ids": userID, "is_active": true}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clients []*models.Client
	if err := cursor.All(ctx, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

// List returns a paginated list of clients
func (r *ClientRepository) List(ctx context.Context, skip, limit int64) ([]*models.Client, error) {
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clients []*models.Client
	if err := cursor.All(ctx, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

// Delete deletes a client (soft delete by setting is_active to false)
func (r *ClientRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("client not found")
	}

	return nil
}
