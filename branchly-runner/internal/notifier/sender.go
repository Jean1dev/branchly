package notifier

import (
	"context"
	"fmt"
	"log/slog"
)

type EmailSender interface {
	Send(ctx context.Context, event, to string, data JobNotifData) error
}

type StubSender struct{}

func NewStubSender() *StubSender {
	return &StubSender{}
}

func (s *StubSender) Send(_ context.Context, event, to string, data JobNotifData) error {
	slog.Info("[notifier:stub] would send email",
		"event", event,
		"to", to,
		"repo", data.RepoFullName,
		"branch", data.BranchName,
		"agent", data.AgentName,
		"duration_s", fmt.Sprintf("%.0f", data.DurationSeconds),
	)
	return nil
}
