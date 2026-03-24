package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type JobLogRepository struct {
	logs *mongo.Collection
	jobs *mongo.Collection
}

func NewJobLogRepository(db *mongo.Database) *JobLogRepository {
	return &JobLogRepository{
		logs: db.Collection("job_logs"),
		jobs: db.Collection("jobs"),
	}
}

func (r *JobLogRepository) Append(ctx context.Context, jobID string, entry domain.LogEntry) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := r.logs.InsertOne(ctx, bson.M{
		"job_id":    jobID,
		"timestamp": entry.Timestamp.UTC(),
		"level":     string(entry.Level),
		"message":   entry.Message,
	})
	if err != nil {
		return fmt.Errorf("job log repository: insert: %w", err)
	}
	_, err = r.jobs.UpdateOne(ctx, bson.M{"_id": jobID}, bson.M{
		"$set": bson.M{"updated_at": time.Now().UTC()},
	})
	if err != nil {
		return fmt.Errorf("job log repository: touch job: %w", err)
	}
	return nil
}
