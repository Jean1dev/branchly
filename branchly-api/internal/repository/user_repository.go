package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) domain.UserRepository {
	return &mongoUserRepository{coll: db.Collection("users")}
}

func (r *mongoUserRepository) UpsertByProvider(ctx context.Context, u *domain.User) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{
		"provider":    u.Provider,
		"provider_id": u.ProviderID,
	}
	now := time.Now().UTC()
	set := bson.M{
		"email":           u.Email,
		"name":            u.Name,
		"avatar_url":      u.AvatarURL,
		"encrypted_token": u.EncryptedToken,
		"updated_at":      now,
	}
	update := bson.M{
		"$set": set,
		"$setOnInsert": bson.M{
			"_id":         uuid.New().String(),
			"provider":    u.Provider,
			"provider_id": u.ProviderID,
			"created_at":  now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out domain.User
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("user repository: upsert: %w", err)
	}
	return &out, nil
}

func (r *mongoUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
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
