package notifier

import (
	"context"
	"fmt"
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
	sender func(event, to string, data JobNotifData)
}

func New(users domain.UserRepository) *Notifier {
	return &Notifier{users: users, sender: logEmail}
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

func logEmail(event, to string, data JobNotifData) {
	slog.Info("[notifier:stub] would send email",
		"event", event,
		"to", to,
		"repo", data.RepoFullName,
		"branch", data.BranchName,
		"agent", data.AgentName,
		"duration_s", fmt.Sprintf("%.0f", data.DurationSeconds),
	)
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
	n.sender("job_completed", user.Email, data)
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
	n.sender("job_failed", user.Email, data)
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
	n.sender("pr_opened", user.Email, data)
}
