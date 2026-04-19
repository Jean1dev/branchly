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
	users domain.UserRepository
}

func New(users domain.UserRepository) *Notifier {
	return &Notifier{users: users}
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

func logEmail(event, to, subject string, data JobNotifData) {
	slog.Info("[notifier:stub] would send email",
		"event", event,
		"to", to,
		"subject", subject,
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
	logEmail("job_completed", user.Email,
		fmt.Sprintf("Job completed on %s", data.RepoFullName), data)
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
	logEmail("job_failed", user.Email,
		fmt.Sprintf("Job failed on %s", data.RepoFullName), data)
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
	logEmail("pr_opened", user.Email,
		fmt.Sprintf("Pull request opened on %s", data.RepoFullName), data)
}
