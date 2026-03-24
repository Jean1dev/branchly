package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoJobLogRepository struct {
	logs *mongo.Collection
	jobs *mongo.Collection
}

func NewJobLogRepository(db *mongo.Database) domain.JobLogRepository {
	return &mongoJobLogRepository{
		logs: db.Collection("job_logs"),
		jobs: db.Collection("jobs"),
	}
}

func (r *mongoJobLogRepository) Append(ctx context.Context, jobID string, entry domain.LogEntry) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

type jobLogBSON struct {
	ID        primitive.ObjectID `bson:"_id"`
	JobID     string             `bson:"job_id"`
	Timestamp time.Time          `bson:"timestamp"`
	Level     string             `bson:"level"`
	Message   string             `bson:"message"`
}

func (r *mongoJobLogRepository) ListByJobID(ctx context.Context, jobID string, limit int) ([]domain.StoredJobLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if limit < 1 {
		limit = 1
	}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
	cur, err := r.logs.Find(ctx, bson.M{"job_id": jobID}, opts)
	if err != nil {
		return nil, fmt.Errorf("job log repository: find: %w", err)
	}
	defer cur.Close(ctx)
	return decodeJobLogCursor(ctx, cur)
}

func (r *mongoJobLogRepository) ListTailByJobID(ctx context.Context, jobID string, n int) ([]domain.StoredJobLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if n < 1 {
		n = 1
	}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetLimit(int64(n))
	cur, err := r.logs.Find(ctx, bson.M{"job_id": jobID}, opts)
	if err != nil {
		return nil, fmt.Errorf("job log repository: find tail: %w", err)
	}
	defer cur.Close(ctx)
	rows, err := decodeJobLogCursor(ctx, cur)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, nil
}

func (r *mongoJobLogRepository) ListByJobIDAfter(ctx context.Context, jobID string, after primitive.ObjectID, limit int) ([]domain.StoredJobLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if limit < 1 {
		limit = 1
	}
	filter := bson.M{"job_id": jobID, "_id": bson.M{"$gt": after}}
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(int64(limit))
	cur, err := r.logs.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("job log repository: find after: %w", err)
	}
	defer cur.Close(ctx)
	return decodeJobLogCursor(ctx, cur)
}

func decodeJobLogCursor(ctx context.Context, cur *mongo.Cursor) ([]domain.StoredJobLog, error) {
	var out []domain.StoredJobLog
	for cur.Next(ctx) {
		var row jobLogBSON
		if err := cur.Decode(&row); err != nil {
			return nil, fmt.Errorf("job log repository: decode: %w", err)
		}
		out = append(out, domain.StoredJobLog{
			ID: row.ID,
			Entry: domain.LogEntry{
				Timestamp: row.Timestamp.UTC(),
				Level:     domain.LogLevel(row.Level),
				Message:   row.Message,
			},
		})
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("job log repository: cursor: %w", err)
	}
	return out, nil
}
