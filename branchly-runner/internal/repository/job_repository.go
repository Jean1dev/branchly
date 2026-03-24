package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type JobRepository struct {
	coll *mongo.Collection
}

func NewJobRepository(db *mongo.Database) *JobRepository {
	return &JobRepository{coll: db.Collection("jobs")}
}

func (r *JobRepository) UpdateJobFields(ctx context.Context, id string, status domain.JobStatus, prURL string, branchName string, completedAt *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	if err != nil {
		return fmt.Errorf("job repository: update fields: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("job repository: job not found")
	}
	return nil
}

func (r *JobRepository) VerifyJobOwner(ctx context.Context, id, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	n, err := r.coll.CountDocuments(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return fmt.Errorf("job repository: count: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("job not found for user")
	}
	return nil
}
