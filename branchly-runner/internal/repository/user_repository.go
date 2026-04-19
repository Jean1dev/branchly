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

type UserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var u domain.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("user repository: find: %w", err)
	}
	return &u, nil
}
