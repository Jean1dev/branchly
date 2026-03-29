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

// Repository is the minimal projection of a connected repository needed by
// the runner for ownership validation. The full document lives in the API.
type Repository struct {
	ID       string `bson:"_id"`
	UserID   string `bson:"user_id"`
	FullName string `bson:"full_name"`
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
)

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelSuccess LogLevel = "success"
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
