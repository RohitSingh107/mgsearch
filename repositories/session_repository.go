package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mgsearch/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionRepository struct {
	collection *mongo.Collection
}

func NewSessionRepository(db *mongo.Database) *SessionRepository {
	return &SessionRepository{collection: db.Collection("sessions")}
}

func (r *SessionRepository) CreateOrUpdate(ctx context.Context, session *models.Session) error {
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": session.ID}
	update := bson.M{
		"$set": bson.M{
			"shop":           session.Shop,
			"state":          session.State,
			"is_online":      session.IsOnline,
			"scope":          session.Scope,
			"expires":        session.Expires,
			"access_token":   session.AccessToken,
			"user_id":        session.UserID,
			"first_name":     session.FirstName,
			"last_name":      session.LastName,
			"email":          session.Email,
			"account_owner":  session.AccountOwner,
			"locale":         session.Locale,
			"collaborator":   session.Collaborator,
			"email_verified": session.EmailVerified,
			"updated_at":     session.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": session.CreatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	var session models.Session
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *SessionRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	filter := bson.M{"expires": bson.M{"$lt": now}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *SessionRepository) GetByShop(ctx context.Context, shop string) ([]*models.Session, error) {
	filter := bson.M{"shop": shop}
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var sessions []*models.Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, fmt.Errorf("failed to decode sessions: %w", err)
	}

	return sessions, nil
}
