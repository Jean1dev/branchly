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

type RepoRepository struct {
	coll *mongo.Collection
}

func NewRepoRepository(db *mongo.Database) *RepoRepository {
	return &RepoRepository{coll: db.Collection("repositories")}
}

func (r *RepoRepository) FindByID(ctx context.Context, id string) (*domain.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var out domain.Repository
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("repo repository: find: %w", err)
	}
	return &out, nil
}
