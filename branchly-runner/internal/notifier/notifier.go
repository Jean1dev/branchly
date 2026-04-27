package notifier

import (
	"context"
	"log/slog"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

type JobNotifData struct {
	UserID           string
	RepoFullName     string
	BranchName       string
	AgentName        string
	Prompt           string
	DurationSeconds  float64
	EstimatedCostUSD *float64
	PRUrl            string
	JobLogsUrl       string
	ErrorMessage     string
	FinishedAt       time.Time
}

type Notifier struct {
	users  domain.UserRepository
	sender EmailSender
}

func New(users domain.UserRepository, sender EmailSender) *Notifier {
	if sender == nil {
		sender = NewStubSender()
	}
	return &Notifier{users: users, sender: sender}
}

func (n *Notifier) fetchUser(ctx context.Context, userID string) (*domain.User, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user, err := n.users.FindByID(ctx, userID)
	if err != nil {
		slog.Warn("notifier: failed to fetch user", "user_id", userID, "error", err)
		return nil, false
	}
	if user == nil {
		slog.Warn("notifier: user not found", "user_id", userID)
		return nil, false
	}
	return user, true
}

func (n *Notifier) NotifyJobCompleted(ctx context.Context, data JobNotifData) {
	user, ok := n.fetchUser(ctx, data.UserID)
	if !ok {
		return
	}
	prefs := user.NotificationPreferences.Email
	if !prefs.Enabled || !prefs.OnJobCompleted {
		return
	}
	if err := n.sender.Send(ctx, "job_completed", user.Email, data); err != nil {
		slog.Warn("notifier: send failed", "event", "job_completed", "user_id", data.UserID, "error", err)
	}
}

func (n *Notifier) NotifyJobFailed(ctx context.Context, data JobNotifData) {
	user, ok := n.fetchUser(ctx, data.UserID)
	if !ok {
		return
	}
	prefs := user.NotificationPreferences.Email
	if !prefs.Enabled || !prefs.OnJobFailed {
		return
	}
	if err := n.sender.Send(ctx, "job_failed", user.Email, data); err != nil {
		slog.Warn("notifier: send failed", "event", "job_failed", "user_id", data.UserID, "error", err)
	}
}

func (n *Notifier) NotifyPROpened(ctx context.Context, data JobNotifData) {
	user, ok := n.fetchUser(ctx, data.UserID)
	if !ok {
		return
	}
	prefs := user.NotificationPreferences.Email
	if !prefs.Enabled || !prefs.OnPROpened {
		return
	}
	if err := n.sender.Send(ctx, "pr_opened", user.Email, data); err != nil {
		slog.Warn("notifier: send failed", "event", "pr_opened", "user_id", data.UserID, "error", err)
	}
}
