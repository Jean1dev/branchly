package domain

import (
	"context"
	"time"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelSuccess LogLevel = "success"
	LogLevelError   LogLevel = "error"
)

type LogEntry struct {
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Level     LogLevel  `bson:"level" json:"level"`
	Message   string    `bson:"message" json:"message"`
}

type Job struct {
	ID           string     `bson:"_id"`
	UserID       string     `bson:"user_id"`
	RepositoryID string     `bson:"repository_id"`
	Prompt       string     `bson:"prompt"`
	Status       JobStatus  `bson:"status"`
	BranchName   string     `bson:"branch_name"`
	PRUrl        string     `bson:"pr_url,omitempty"`
	Logs         []LogEntry `bson:"logs"`
	CreatedAt    time.Time  `bson:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"`
	CompletedAt  *time.Time `bson:"completed_at,omitempty"`
}

type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	FindByUserID(ctx context.Context, userID string, status *JobStatus, repositoryID *string) ([]*Job, error)
	UpdateStatus(ctx context.Context, id string, status JobStatus) error
	UpdateJobFields(ctx context.Context, id string, status JobStatus, prURL string, branchName string, completedAt *time.Time) error
	AppendLog(ctx context.Context, id string, entry LogEntry) error
	FindByIDForUser(ctx context.Context, id string, userID string) (*Job, error)
}
