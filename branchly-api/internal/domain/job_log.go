package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobLogRepository interface {
	Append(ctx context.Context, jobID string, entry LogEntry) error
	ListByJobID(ctx context.Context, jobID string, limit int) ([]StoredJobLog, error)
	ListTailByJobID(ctx context.Context, jobID string, n int) ([]StoredJobLog, error)
	ListByJobIDAfter(ctx context.Context, jobID string, after primitive.ObjectID, limit int) ([]StoredJobLog, error)
}

type StoredJobLog struct {
	ID        primitive.ObjectID
	Entry     LogEntry
}
