package domain

import (
	"context"
	"math"
	"time"
)

type AgentType string

const (
	AgentTypeClaudeCode AgentType = "claude-code"
	AgentTypeGemini     AgentType = "gemini"
)

func (a AgentType) IsValid() bool {
	switch a {
	case AgentTypeClaudeCode, AgentTypeGemini:
		return true
	}
	return false
}

// Repository is the minimal projection of a connected repository needed by
// the runner for ownership validation and cloning.
type Repository struct {
	ID            string      `bson:"_id"`
	UserID        string      `bson:"user_id"`
	IntegrationID string      `bson:"integration_id"`
	Provider      GitProvider `bson:"provider"`
	FullName      string      `bson:"full_name"`
	CloneURL      string      `bson:"clone_url"`
	DefaultBranch string      `bson:"default_branch"`
}

// RepositoryRepository provides read-only access to the repositories collection.
type RepositoryRepository interface {
	FindByID(ctx context.Context, id string) (*Repository, error)
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
)

type FailureType string

const (
	FailureTypeTransient FailureType = "transient"
	FailureTypePermanent FailureType = "permanent"
)

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelSuccess LogLevel = "success"
	LogLevelWarn    LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

type LogEntry struct {
	Timestamp time.Time `bson:"timestamp"`
	Level     LogLevel  `bson:"level"`
	Message   string    `bson:"message"`
}

type JobCost struct {
	AgentType    AgentType `bson:"agent_type"`
	ModelUsed    string    `bson:"model_used"`
	InputTokens  int64     `bson:"input_tokens"`
	OutputTokens int64     `bson:"output_tokens"`
	TotalTokens  int64     `bson:"total_tokens"`
	EstimatedUSD float64   `bson:"estimated_usd"`
	DurationSecs float64   `bson:"duration_secs"`
}

type Job struct {
	ID            string      `bson:"_id"`
	UserID        string      `bson:"user_id"`
	RepositoryID  string      `bson:"repository_id"`
	Prompt        string      `bson:"prompt"`
	Status        JobStatus   `bson:"status"`
	AgentType     AgentType   `bson:"agent_type"`
	BranchName    string      `bson:"branch_name"`
	PRUrl         string      `bson:"pr_url,omitempty"`
	Cost          *JobCost    `bson:"cost,omitempty"`
	AttemptNumber int         `bson:"attempt_number"`
	MaxAttempts   int         `bson:"max_attempts"`
	LastError     string      `bson:"last_error,omitempty"`
	NextRetryAt   *time.Time  `bson:"next_retry_at,omitempty"`
	FailureType   FailureType `bson:"failure_type,omitempty"`
	CreatedAt     time.Time   `bson:"created_at"`
	UpdatedAt     time.Time   `bson:"updated_at"`
	CompletedAt   *time.Time  `bson:"completed_at,omitempty"`
}

// CanRetry reports whether the job is eligible for automatic retry.
func (j *Job) CanRetry() bool {
	return j.AttemptNumber < j.MaxAttempts &&
		j.FailureType == FailureTypeTransient
}

// BackoffDuration returns the exponential backoff delay before the next attempt.
// Sequence: attempt 1 → 1 min, attempt 2 → 5 min, attempt 3 → 25 min.
func (j *Job) BackoffDuration() time.Duration {
	base := time.Minute
	multiplier := math.Pow(5, float64(j.AttemptNumber-1))
	return time.Duration(float64(base) * multiplier)
}
