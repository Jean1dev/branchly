package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *JobRepository) SetCost(ctx context.Context, id string, cost *domain.JobCost) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"cost":       cost,
		"updated_at": time.Now().UTC(),
	}})
	if err != nil {
		return fmt.Errorf("job repository: set cost: %w", err)
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

func (r *JobRepository) SetRetrying(ctx context.Context, id string, lastError string, failureType domain.FailureType, nextRetryAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"status":        string(domain.JobStatusRetrying),
			"last_error":    lastError,
			"failure_type":  string(failureType),
			"next_retry_at": nextRetryAt,
			"updated_at":    time.Now().UTC(),
		},
	})
	if err != nil {
		return fmt.Errorf("job repository: set retrying: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("job repository: job not found")
	}
	return nil
}

func (r *JobRepository) SetFailed(ctx context.Context, id string, lastError string, failureType domain.FailureType) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	now := time.Now().UTC()
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"status":       string(domain.JobStatusFailed),
			"last_error":   lastError,
			"failure_type": string(failureType),
			"completed_at": now,
			"updated_at":   now,
		},
	})
	if err != nil {
		return fmt.Errorf("job repository: set failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("job repository: job not found")
	}
	return nil
}

// FindDueForRetry atomically finds jobs in "retrying" status whose next_retry_at
// has passed, increments attempt_number, and transitions them to "running".
// Using FindOneAndUpdate per document is atomic — avoids double-pickup races.
// Returns domain.Job slices; callers must look up repository details separately.
func (r *JobRepository) FindDueForRetry(ctx context.Context) ([]*domain.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	now := time.Now().UTC()
	filter := bson.M{
		"status":        string(domain.JobStatusRetrying),
		"next_retry_at": bson.M{"$lte": now},
	}

	after := options.After
	updateOpts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	var results []*domain.Job
	for {
		var job domain.Job
		err := r.coll.FindOneAndUpdate(ctx, filter, bson.M{
			"$set": bson.M{
				"status":     string(domain.JobStatusRunning),
				"updated_at": time.Now().UTC(),
			},
			"$inc": bson.M{"attempt_number": 1},
		}, &updateOpts).Decode(&job)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				break
			}
			return nil, fmt.Errorf("job repository: find due for retry: %w", err)
		}
		cp := job
		results = append(results, &cp)
	}

	return results, nil
}
