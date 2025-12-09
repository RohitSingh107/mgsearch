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

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("email already exists")
		}
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"is_active":  user.IsActive,
			"client_ids": user.ClientIDs,
			"updated_at": user.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID primitive.ObjectID, passwordHash string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    time.Now().UTC(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// AddClientToUser adds a client ID to user's client_ids array
func (r *UserRepository) AddClientToUser(ctx context.Context, userID, clientID primitive.ObjectID) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$addToSet": bson.M{"client_ids": clientID},
		"$set":      bson.M{"updated_at": time.Now().UTC()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// RemoveClientFromUser removes a client ID from user's client_ids array
func (r *UserRepository) RemoveClientFromUser(ctx context.Context, userID, clientID primitive.ObjectID) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$pull": bson.M{"client_ids": clientID},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// FindByClientID finds all users associated with a client
func (r *UserRepository) FindByClientID(ctx context.Context, clientID primitive.ObjectID) ([]*models.User, error) {
	filter := bson.M{"client_ids": clientID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// List returns a paginated list of users
func (r *UserRepository) List(ctx context.Context, skip, limit int64) ([]*models.User, error) {
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Delete deletes a user (soft delete by setting is_active to false)
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
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
		return errors.New("user not found")
	}

	return nil
}
