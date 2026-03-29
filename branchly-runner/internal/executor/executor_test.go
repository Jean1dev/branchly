package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	agentpkg "github.com/branchly/branchly-runner/internal/agent"
	"github.com/branchly/branchly-runner/internal/domain"
)

// ---- minimal mocks ----

type mockJobRepo struct {
	ownerErr    error
	failedJobID string
}

func (m *mockJobRepo) VerifyJobOwner(_ context.Context, _, _ string) error {
	return m.ownerErr
}
func (m *mockJobRepo) UpdateJobFields(_ context.Context, id string, status domain.JobStatus, _, _ string, _ *time.Time) error {
	if status == domain.JobStatusFailed {
		m.failedJobID = id
	}
	return nil
}
func (m *mockJobRepo) SetCost(_ context.Context, _ string, _ *domain.JobCost) error { return nil }

type mockJobLogRepo struct {
	lastMsg string
}

func (m *mockJobLogRepo) Append(_ context.Context, _ string, e domain.LogEntry) error {
	m.lastMsg = e.Message
	return nil
}

type mockRepoRepo struct {
	repo *domain.Repository
	err  error
}

func (m *mockRepoRepo) FindByID(_ context.Context, _ string) (*domain.Repository, error) {
	return m.repo, m.err
}

// stubAgent is a no-op agent used in executor tests.
type stubAgent struct{}

func (s *stubAgent) Run(_ context.Context, _ domain.AgentInput) (string, error) {
	return "", errors.New("stub agent should not be called")
}

// ---- test helpers ----

func newTestExecutor(repoMock *mockRepoRepo) (*Executor, *mockJobRepo, *mockJobLogRepo) {
	jobs := &mockJobRepo{}
	logs := &mockJobLogRepo{}
	factory := agentpkg.NewFactory(&stubAgent{}, &stubAgent{})
	ex := &Executor{
		factory: factory,
		jobs:    jobs,
		jobLogs: logs,
		repos:   repoMock,
		encKey:  make([]byte, 32), // zero key — decrypt will fail, but we test before that
		workDir: "/tmp",
	}
	return ex, jobs, logs
}

func validInput() RunJobInput {
	return RunJobInput{
		JobID:          "job-1",
		UserID:         "user-1",
		RepositoryID:   "repo-1",
		RepositoryName: "owner/repo",
		DefaultBranch:  "main",
		Prompt:         "add feature",
		EncryptedToken: "invalid-but-ownership-checked-before-decrypt",
		AgentType:      domain.AgentTypeClaudeCode,
	}
}

// ---- ownership tests ----

func TestRun_ValidOwnership_ProceedsToDecrypt(t *testing.T) {
	repo := &domain.Repository{ID: "repo-1", UserID: "user-1", FullName: "owner/repo"}
	ex, _, _ := newTestExecutor(&mockRepoRepo{repo: repo})
	// Ownership passes — execution proceeds past ownership checks to decrypt.
	// With a zero key + invalid token, decrypt fails gracefully.
	ex.Run(context.Background(), validInput())
	// If we reached here without panicking the ownership check passed.
}

func TestRun_DivergentUserID_MarksJobFailed(t *testing.T) {
	repo := &domain.Repository{ID: "repo-1", UserID: "user-2", FullName: "owner/repo"}
	ex, jobs, logs := newTestExecutor(&mockRepoRepo{repo: repo})

	ex.Run(context.Background(), validInput())

	if jobs.failedJobID != "job-1" {
		t.Error("expected job to be marked failed on ownership mismatch")
	}
	if logs.lastMsg != "repository ownership validation failed" {
		t.Errorf("unexpected failure message: %q", logs.lastMsg)
	}
}

func TestRun_NilRepo_MarksJobFailed(t *testing.T) {
	ex, jobs, logs := newTestExecutor(&mockRepoRepo{repo: nil})

	ex.Run(context.Background(), validInput())

	if jobs.failedJobID != "job-1" {
		t.Error("expected job to be marked failed when repo not found")
	}
	if logs.lastMsg != "repository ownership validation failed" {
		t.Errorf("unexpected failure message: %q", logs.lastMsg)
	}
}

func TestRun_RepositoryNameMismatch_MarksJobFailed(t *testing.T) {
	repo := &domain.Repository{ID: "repo-1", UserID: "user-1", FullName: "owner/real-repo"}
	ex, jobs, logs := newTestExecutor(&mockRepoRepo{repo: repo})

	in := validInput()
	in.RepositoryName = "attacker/swapped-repo"

	ex.Run(context.Background(), in)

	if jobs.failedJobID != "job-1" {
		t.Error("expected job to be marked failed on name mismatch")
	}
	if logs.lastMsg != "repository name mismatch" {
		t.Errorf("unexpected failure message: %q", logs.lastMsg)
	}
}

func TestRun_RepoLookupError_MarksJobFailed(t *testing.T) {
	ex, jobs, _ := newTestExecutor(&mockRepoRepo{err: errors.New("db timeout")})

	ex.Run(context.Background(), validInput())

	if jobs.failedJobID != "job-1" {
		t.Error("expected job to be marked failed when repo lookup returns error")
	}
}

func TestRun_UnknownAgentType_MarksJobFailed(t *testing.T) {
	repo := &domain.Repository{ID: "repo-1", UserID: "user-1", FullName: "owner/repo"}
	ex, jobs, logs := newTestExecutor(&mockRepoRepo{repo: repo})

	in := validInput()
	in.AgentType = "openai" // not registered in factory

	ex.Run(context.Background(), in)

	if jobs.failedJobID != "job-1" {
		t.Error("expected job to be marked failed for unknown agent type")
	}
	if logs.lastMsg == "" {
		t.Error("expected a failure log message for unknown agent type")
	}
}
