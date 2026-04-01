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

type IntegrationRepository struct {
	coll *mongo.Collection
}

func NewIntegrationRepository(db *mongo.Database) *IntegrationRepository {
	return &IntegrationRepository{coll: db.Collection("git_integrations")}
}

func (r *IntegrationRepository) FindByID(ctx context.Context, id string) (*domain.GitIntegration, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
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
