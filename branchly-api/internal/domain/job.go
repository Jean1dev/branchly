package domain

import (
	"context"
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

type JobCost struct {
	AgentType    AgentType `bson:"agent_type"    json:"agent_type"`
	ModelUsed    string    `bson:"model_used"    json:"model_used"`
	InputTokens  int64     `bson:"input_tokens"  json:"input_tokens"`
	OutputTokens int64     `bson:"output_tokens" json:"output_tokens"`
	TotalTokens  int64     `bson:"total_tokens"  json:"total_tokens"`
	EstimatedUSD float64   `bson:"estimated_usd" json:"estimated_usd"`
	DurationSecs float64   `bson:"duration_secs" json:"duration_secs"`
}

type Job struct {
	ID           string     `bson:"_id"`
	UserID       string     `bson:"user_id"`
	RepositoryID string     `bson:"repository_id"`
	Prompt       string     `bson:"prompt"`
	Status       JobStatus  `bson:"status"`
	AgentType    AgentType  `bson:"agent_type"`
	BranchName   string     `bson:"branch_name"`
	PRUrl        string     `bson:"pr_url,omitempty"`
	Logs         []LogEntry `bson:"logs,omitempty"`
	Cost         *JobCost   `bson:"cost,omitempty"`
	CreatedAt    time.Time  `bson:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"`
	CompletedAt  *time.Time `bson:"completed_at,omitempty"`
}

type JobRepository interface {
	Create(ctx context.Context, job *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	FindByUserID(ctx context.Context, userID string, status *JobStatus, repositoryID *string) ([]*Job, error)
	CountActiveByUserID(ctx context.Context, userID string) (int64, error)
	UpdateStatus(ctx context.Context, id string, status JobStatus) error
	UpdateJobFields(ctx context.Context, id string, status JobStatus, prURL string, branchName string, completedAt *time.Time) error
	FindByIDForUser(ctx context.Context, id string, userID string) (*Job, error)
}
