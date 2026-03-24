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

type mongoJobRepository struct {
	coll *mongo.Collection
}

func NewJobRepository(db *mongo.Database) domain.JobRepository {
	return &mongoJobRepository{coll: db.Collection("jobs")}
}

func (r *mongoJobRepository) Create(ctx context.Context, job *domain.Job) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, job)
	if err != nil {
		return fmt.Errorf("job repository: create: %w", err)
	}
	return nil
}

func (r *mongoJobRepository) FindByID(ctx context.Context, id string) (*domain.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var j domain.Job
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&j)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("job repository: find: %w", err)
	}
	return &j, nil
}

func (r *mongoJobRepository) FindByIDForUser(ctx context.Context, id string, userID string) (*domain.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var j domain.Job
	err := r.coll.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&j)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("job repository: find for user: %w", err)
	}
	return &j, nil
}

func (r *mongoJobRepository) FindByUserID(ctx context.Context, userID string, status *domain.JobStatus, repositoryID *string) ([]*domain.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{"user_id": userID}
	if status != nil && *status != "" {
		filter["status"] = string(*status)
	}
	if repositoryID != nil && *repositoryID != "" {
		filter["repository_id"] = *repositoryID
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("job repository: list: %w", err)
	}
	defer cur.Close(ctx)
	var list []*domain.Job
	for cur.Next(ctx) {
		var item domain.Job
		if err := cur.Decode(&item); err != nil {
			return nil, fmt.Errorf("job repository: decode: %w", err)
		}
		cp := item
		list = append(list, &cp)
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("job repository: cursor: %w", err)
	}
	return list, nil
}

func (r *mongoJobRepository) UpdateStatus(ctx context.Context, id string, status domain.JobStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"status":     string(status),
			"updated_at": time.Now().UTC(),
		},
	})
	if err != nil {
		return fmt.Errorf("job repository: update status: %w", err)
	}
	return nil
}

func (r *mongoJobRepository) UpdateJobFields(ctx context.Context, id string, status domain.JobStatus, prURL string, branchName string, completedAt *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	set := bson.M{
		"status":     string(status),
		"updated_at": time.Now().UTC(),
	}
	if prURL != "" {
		set["pr_url"] = prURL
	}
	if branchName != "" {
		set["branch_name"] = branchName
	}
	if completedAt != nil {
		set["completed_at"] = completedAt
	}
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	if err != nil {
		return fmt.Errorf("job repository: update fields: %w", err)
	}
	return nil
}

