package repositories

import (
	"context"
	"errors"
	"time"

	"mgsearch/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IndexRepository struct {
	collection *mongo.Collection
}

func NewIndexRepository(db *mongo.Database) *IndexRepository {
	return &IndexRepository{
		collection: db.Collection("indexes"),
	}
}

// Create creates a new index record
func (r *IndexRepository) Create(ctx context.Context, index *models.Index) (*models.Index, error) {
	index.CreatedAt = time.Now().UTC()
	index.UpdatedAt = time.Now().UTC()

	result, err := r.collection.InsertOne(ctx, index)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("index already exists")
		}
		return nil, err
	}

	index.ID = result.InsertedID.(primitive.ObjectID)
	return index, nil
}

// FindByClientID finds all indexes for a client
func (r *IndexRepository) FindByClientID(ctx context.Context, clientID primitive.ObjectID) ([]*models.Index, error) {
	filter := bson.M{"client_id": clientID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []*models.Index
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, err
	}

	return indexes, nil
}

// FindByNameAndClientID finds a specific index for a client by name
func (r *IndexRepository) FindByNameAndClientID(ctx context.Context, name string, clientID primitive.ObjectID) (*models.Index, error) {
	var index models.Index
	filter := bson.M{"name": name, "client_id": clientID}
	err := r.collection.FindOne(ctx, filter).Decode(&index)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("index not found")
		}
		return nil, err
	}
	return &index, nil
}

// FindByID finds an index by ID
func (r *IndexRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Index, error) {
	var index models.Index
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&index)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("index not found")
		}
		return nil, err
	}
	return &index, nil
}
