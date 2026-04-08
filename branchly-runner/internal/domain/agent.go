package domain

import "context"

type Agent interface {
	Run(ctx context.Context, input AgentInput) (summary string, err error)
}

type AgentInput struct {
	WorkDir    string
	Prompt     string
	RepoName   string
	BranchName string
	APIKey     string // resolved at runtime; never logged or persisted
	OnLog      func(level LogLevel, message string)
}
