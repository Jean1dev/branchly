package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

// ---- mock user repository ----

type mockUserRepo struct {
	user *domain.User
	err  error
}

func (m *mockUserRepo) FindByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

// ---- helpers ----

type sentEmail struct {
	event string
	to    string
	data  JobNotifData
}

type mockSender struct {
	calls *[]sentEmail
	err   error
}

func (m *mockSender) Send(_ context.Context, event, to string, data JobNotifData) error {
	*m.calls = append(*m.calls, sentEmail{event: event, to: to, data: data})
	return m.err
}

func allEnabledUser() *domain.User {
	return &domain.User{
		ID:    "user-1",
		Email: "user@example.com",
		Name:  "Test User",
		NotificationPreferences: domain.NotificationPreferences{
			Email: domain.EmailNotificationPreferences{
				Enabled:        true,
				OnJobCompleted: true,
				OnJobFailed:    true,
				OnPROpened:     true,
			},
		},
	}
}

func testData() JobNotifData {
	return JobNotifData{
		UserID:          "user-1",
		RepoFullName:    "owner/repo",
		BranchName:      "branchly/add-feature",
		AgentName:       "Claude Code",
		Prompt:          "add a test",
		DurationSeconds: 42,
		FinishedAt:      time.Now(),
	}
}

func newNotifier(user *domain.User, userErr error, calls *[]sentEmail) *Notifier {
	return &Notifier{
		users:  &mockUserRepo{user: user, err: userErr},
		sender: &mockSender{calls: calls},
	}
}

// ---- tests: NotifyJobCompleted ----

func TestNotifyJobCompleted_AllEnabled_SendsEmail(t *testing.T) {
	var calls []sentEmail
	n := newNotifier(allEnabledUser(), nil, &calls)
	n.NotifyJobCompleted(context.Background(), testData())

	if len(calls) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(calls))
	}
	if calls[0].event != "job_completed" {
		t.Errorf("expected event job_completed, got %s", calls[0].event)
	}
	if calls[0].to != "user@example.com" {
		t.Errorf("expected to user@example.com, got %s", calls[0].to)
	}
}

func TestNotifyJobCompleted_MasterDisabled_NoEmail(t *testing.T) {
	u := allEnabledUser()
	u.NotificationPreferences.Email.Enabled = false
	var calls []sentEmail
	n := newNotifier(u, nil, &calls)
	n.NotifyJobCompleted(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email when master disabled, got %d", len(calls))
	}
}

func TestNotifyJobCompleted_SpecificPrefDisabled_NoEmail(t *testing.T) {
	u := allEnabledUser()
	u.NotificationPreferences.Email.OnJobCompleted = false
	var calls []sentEmail
	n := newNotifier(u, nil, &calls)
	n.NotifyJobCompleted(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email when on_job_completed disabled, got %d", len(calls))
	}
}

func TestNotifyJobCompleted_UserNotFound_NoEmail(t *testing.T) {
	var calls []sentEmail
	n := newNotifier(nil, nil, &calls)
	n.NotifyJobCompleted(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email when user not found, got %d", len(calls))
	}
}

func TestNotifyJobCompleted_RepoError_NoEmail(t *testing.T) {
	var calls []sentEmail
	n := newNotifier(nil, errors.New("db down"), &calls)
	n.NotifyJobCompleted(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email on repo error, got %d", len(calls))
	}
}

// ---- tests: NotifyJobFailed ----

func TestNotifyJobFailed_AllEnabled_SendsEmail(t *testing.T) {
	var calls []sentEmail
	n := newNotifier(allEnabledUser(), nil, &calls)
	n.NotifyJobFailed(context.Background(), testData())

	if len(calls) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(calls))
	}
	if calls[0].event != "job_failed" {
		t.Errorf("expected event job_failed, got %s", calls[0].event)
	}
}

func TestNotifyJobFailed_SpecificPrefDisabled_NoEmail(t *testing.T) {
	u := allEnabledUser()
	u.NotificationPreferences.Email.OnJobFailed = false
	var calls []sentEmail
	n := newNotifier(u, nil, &calls)
	n.NotifyJobFailed(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email when on_job_failed disabled, got %d", len(calls))
	}
}

func TestNotifyJobFailed_MasterDisabled_NoEmail(t *testing.T) {
	u := allEnabledUser()
	u.NotificationPreferences.Email.Enabled = false
	var calls []sentEmail
	n := newNotifier(u, nil, &calls)
	n.NotifyJobFailed(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email when master disabled, got %d", len(calls))
	}
}

// ---- tests: NotifyPROpened ----

func TestNotifyPROpened_AllEnabled_SendsEmail(t *testing.T) {
	var calls []sentEmail
	n := newNotifier(allEnabledUser(), nil, &calls)
	n.NotifyPROpened(context.Background(), testData())

	if len(calls) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(calls))
	}
	if calls[0].event != "pr_opened" {
		t.Errorf("expected event pr_opened, got %s", calls[0].event)
	}
}

func TestNotifyPROpened_SpecificPrefDisabled_NoEmail(t *testing.T) {
	u := allEnabledUser()
	u.NotificationPreferences.Email.OnPROpened = false
	var calls []sentEmail
	n := newNotifier(u, nil, &calls)
	n.NotifyPROpened(context.Background(), testData())

	if len(calls) != 0 {
		t.Errorf("expected no email when on_pr_opened disabled, got %d", len(calls))
	}
}

func TestNotifyPROpened_DataPassedThrough(t *testing.T) {
	var calls []sentEmail
	d := testData()
	d.PRUrl = "https://github.com/owner/repo/pull/1"
	n := newNotifier(allEnabledUser(), nil, &calls)
	n.NotifyPROpened(context.Background(), d)

	if len(calls) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(calls))
	}
	if calls[0].data.PRUrl != d.PRUrl {
		t.Errorf("expected pr_url %s, got %s", d.PRUrl, calls[0].data.PRUrl)
	}
	if calls[0].data.RepoFullName != d.RepoFullName {
		t.Errorf("expected repo %s, got %s", d.RepoFullName, calls[0].data.RepoFullName)
	}
}
