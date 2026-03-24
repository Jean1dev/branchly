package domain

import "time"

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
