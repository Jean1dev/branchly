package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoAPIKeyRepository struct {
	coll *mongo.Collection
}

func NewAPIKeyRepository(db *mongo.Database) domain.APIKeyRepository {
	return &mongoAPIKeyRepository{coll: db.Collection("user_api_keys")}
}

func (r *mongoAPIKeyRepository) Upsert(ctx context.Context, key *domain.UserAPIKey) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": key.UserID, "provider": string(key.Provider)}
	update := bson.M{
		"$set": bson.M{
			"_id":           key.ID,
			"user_id":       key.UserID,
			"provider":      string(key.Provider),
			"encrypted_key": key.EncryptedKey,
			"key_hint":      key.KeyHint,
			"updated_at":    key.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"created_at": key.CreatedAt,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("api key repository: upsert: %w", err)
	}
	return nil
}

func (r *mongoAPIKeyRepository) FindByUserAndProvider(ctx context.Context, userID string, provider domain.APIKeyProvider) (*domain.UserAPIKey, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var k domain.UserAPIKey
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "provider": string(provider)}).Decode(&k)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("api key repository: find by user and provider: %w", err)
	}
	return &k, nil
}

func (r *mongoAPIKeyRepository) FindAllByUserID(ctx context.Context, userID string) ([]*domain.UserAPIKey, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cur, err := r.coll.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("api key repository: find all: %w", err)
	}
	defer cur.Close(ctx)

	var list []*domain.UserAPIKey
	for cur.Next(ctx) {
		var k domain.UserAPIKey
		if err := cur.Decode(&k); err != nil {
			return nil, fmt.Errorf("api key repository: decode: %w", err)
		}
		cp := k
		list = append(list, &cp)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("api key repository: cursor: %w", err)
	}
	return list, nil
}

func (r *mongoAPIKeyRepository) Delete(ctx context.Context, userID string, provider domain.APIKeyProvider) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := r.coll.DeleteOne(ctx, bson.M{"user_id": userID, "provider": string(provider)})
	if err != nil {
		return fmt.Errorf("api key repository: delete: %w", err)
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("api key repository: not found")
	}
	return nil
}
