package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIKeyRepository struct {
	coll *mongo.Collection
}

func NewAPIKeyRepository(db *mongo.Database) *APIKeyRepository {
	return &APIKeyRepository{coll: db.Collection("user_api_keys")}
}

func (r *APIKeyRepository) FindByUserAndProvider(ctx context.Context, userID string, provider domain.APIKeyProvider) (*domain.UserAPIKey, error) {
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
