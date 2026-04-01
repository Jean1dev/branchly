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

// IntegrationRepository persists git integrations in MongoDB.
type IntegrationRepository interface {
	Upsert(ctx context.Context, ig *domain.GitIntegration) error
	FindByUserAndProvider(ctx context.Context, userID string, provider domain.GitProvider) (*domain.GitIntegration, error)
	FindByID(ctx context.Context, id string) (*domain.GitIntegration, error)
	FindAllByUserID(ctx context.Context, userID string) ([]*domain.GitIntegration, error)
	Delete(ctx context.Context, id string) error
}

type mongoIntegrationRepository struct {
	coll *mongo.Collection
}

func NewIntegrationRepository(db *mongo.Database) IntegrationRepository {
	return &mongoIntegrationRepository{coll: db.Collection("git_integrations")}
}

func (r *mongoIntegrationRepository) Upsert(ctx context.Context, ig *domain.GitIntegration) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{"user_id": ig.UserID, "provider": ig.Provider}
	update := bson.M{
		"$set": bson.M{
			"encrypted_token": ig.EncryptedToken,
			"token_type":      ig.TokenType,
			"scopes":          ig.Scopes,
		},
		"$setOnInsert": bson.M{
			"_id":           ig.ID,
			"user_id":       ig.UserID,
			"provider":      ig.Provider,
			"connected_at":  ig.ConnectedAt,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("integration repository: upsert: %w", err)
	}
	return nil
}

func (r *mongoIntegrationRepository) FindByUserAndProvider(ctx context.Context, userID string, provider domain.GitProvider) (*domain.GitIntegration, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var out domain.GitIntegration
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "provider": provider}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("integration repository: find by user+provider: %w", err)
	}
	return &out, nil
}

func (r *mongoIntegrationRepository) FindByID(ctx context.Context, id string) (*domain.GitIntegration, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var out domain.GitIntegration
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("integration repository: find by id: %w", err)
	}
	return &out, nil
}

func (r *mongoIntegrationRepository) FindAllByUserID(ctx context.Context, userID string) ([]*domain.GitIntegration, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cur, err := r.coll.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("integration repository: find all by user: %w", err)
	}
	defer cur.Close(ctx)
	var list []*domain.GitIntegration
	for cur.Next(ctx) {
		var item domain.GitIntegration
		if err := cur.Decode(&item); err != nil {
			return nil, fmt.Errorf("integration repository: decode: %w", err)
		}
		cp := item
		list = append(list, &cp)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("integration repository: cursor: %w", err)
	}
	return list, nil
}

func (r *mongoIntegrationRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("integration repository: delete: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
