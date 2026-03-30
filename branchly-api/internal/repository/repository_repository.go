package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoConnectedRepository struct {
	coll *mongo.Collection
}

func NewConnectedRepositoryRepository(db *mongo.Database) domain.ConnectedRepositoryRepository {
	return &mongoConnectedRepository{coll: db.Collection("repositories")}
}

func (r *mongoConnectedRepository) Create(ctx context.Context, repo *domain.Repository) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, repo)
	if err != nil {
		return fmt.Errorf("repository repository: create: %w", err)
	}
	return nil
}

func (r *mongoConnectedRepository) FindByID(ctx context.Context, id string) (*domain.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var out domain.Repository
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository repository: find: %w", err)
	}
	return &out, nil
}

func (r *mongoConnectedRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cur, err := r.coll.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("repository repository: find by user: %w", err)
	}
	defer cur.Close(ctx)
	var list []*domain.Repository
	for cur.Next(ctx) {
		var item domain.Repository
		if err := cur.Decode(&item); err != nil {
			return nil, fmt.Errorf("repository repository: decode: %w", err)
		}
		cp := item
		list = append(list, &cp)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("repository repository: cursor: %w", err)
	}
	return list, nil
}

func (r *mongoConnectedRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("repository repository: delete: %w", err)
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *mongoConnectedRepository) FindByUserExternalAndProvider(ctx context.Context, userID, externalID string, provider domain.GitProvider) (*domain.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var out domain.Repository
	err := r.coll.FindOne(ctx, bson.M{
		"user_id":     userID,
		"external_id": externalID,
		"provider":    provider,
	}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("repository repository: find by user+external+provider: %w", err)
	}
	return &out, nil
}

func (r *mongoConnectedRepository) FindByIntegrationID(ctx context.Context, integrationID string) ([]*domain.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cur, err := r.coll.Find(ctx, bson.M{"integration_id": integrationID})
	if err != nil {
		return nil, fmt.Errorf("repository repository: find by integration: %w", err)
	}
	defer cur.Close(ctx)
	var list []*domain.Repository
	for cur.Next(ctx) {
		var item domain.Repository
		if err := cur.Decode(&item); err != nil {
			return nil, fmt.Errorf("repository repository: decode: %w", err)
		}
		cp := item
		list = append(list, &cp)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("repository repository: cursor: %w", err)
	}
	return list, nil
}
