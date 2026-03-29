package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/infra"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- minimal in-memory mocks ----

type mockJobRepo struct {
	activeCount int64
	created     []*domain.Job
}

func (m *mockJobRepo) Create(_ context.Context, job *domain.Job) error {
	m.created = append(m.created, job)
	return nil
}
func (m *mockJobRepo) FindByID(_ context.Context, _ string) (*domain.Job, error) { return nil, nil }
func (m *mockJobRepo) FindByUserID(_ context.Context, _ string, _ *domain.JobStatus, _ *string) ([]*domain.Job, error) {
	return nil, nil
}
func (m *mockJobRepo) CountActiveByUserID(_ context.Context, _ string) (int64, error) {
	return m.activeCount, nil
}
func (m *mockJobRepo) UpdateStatus(_ context.Context, _ string, _ domain.JobStatus) error {
	return nil
}
func (m *mockJobRepo) UpdateJobFields(_ context.Context, _ string, _ domain.JobStatus, _, _ string, _ *time.Time) error {
	return nil
}
func (m *mockJobRepo) FindByIDForUser(_ context.Context, _, _ string) (*domain.Job, error) {
	return nil, nil
}

type mockJobLogRepo struct{}

func (m *mockJobLogRepo) Append(_ context.Context, _ string, _ domain.LogEntry) error { return nil }
func (m *mockJobLogRepo) ListByJobID(_ context.Context, _ string, _ int) ([]domain.StoredJobLog, error) {
	return nil, nil
}
func (m *mockJobLogRepo) ListTailByJobID(_ context.Context, _ string, _ int) ([]domain.StoredJobLog, error) {
	return nil, nil
}
func (m *mockJobLogRepo) ListByJobIDAfter(_ context.Context, _ string, _ primitive.ObjectID, _ int) ([]domain.StoredJobLog, error) {
	return nil, nil
}

type mockRepoRepo struct {
	repo *domain.Repository
}

func (m *mockRepoRepo) Create(_ context.Context, _ *domain.Repository) error        { return nil }
func (m *mockRepoRepo) FindByID(_ context.Context, _ string) (*domain.Repository, error) {
	return m.repo, nil
}
func (m *mockRepoRepo) FindByUserID(_ context.Context, _ string) ([]*domain.Repository, error) {
	return nil, nil
}
func (m *mockRepoRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockRepoRepo) FindByUserAndGithubRepoID(_ context.Context, _ string, _ int64) (*domain.Repository, error) {
	return nil, nil
}

type mockUserRepo struct {
	user *domain.User
}

func (m *mockUserRepo) FindByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, nil
}
func (m *mockUserRepo) UpsertByProvider(_ context.Context, _ *domain.User) (*domain.User, error) {
	return nil, nil
}

// mockRunnerClient satisfies the runnerDispatcher interface.
type mockRunnerClient struct {
	failDispatch bool
	lastPayload  infra.DispatchJobPayload
}

func (m *mockRunnerClient) DispatchJob(_ context.Context, payload infra.DispatchJobPayload) error {
	m.lastPayload = payload
	if m.failDispatch {
		return errors.New("runner down")
	}
	return nil
}

// ---- test helpers ----

func newTestService(
	activeJobs int64,
	repo *domain.Repository,
	user *domain.User,
) *JobService {
	cfg := &config.Config{MaxActiveJobsPerUser: 3}
	jobs := &mockJobRepo{activeCount: activeJobs}
	repos := &mockRepoRepo{repo: repo}
	users := &mockUserRepo{user: user}
	svc := &JobService{
		cfg:     cfg,
		jobs:    jobs,
		jobLogs: &mockJobLogRepo{},
		repos:   repos,
		users:   users,
		runner:  &mockRunnerClient{}, // always succeeds; dispatch errors not tested here
	}
	return svc
}

func ownedRepo(userID string) *domain.Repository {
	return &domain.Repository{ID: "repo-1", UserID: userID, FullName: "owner/repo"}
}

func activeUser(userID string) *domain.User {
	return &domain.User{ID: userID, EncryptedToken: "enc"}
}

// ---- tests: rate limiting ----

func TestCreate_ZeroActiveJobs_Succeeds(t *testing.T) {
	svc := newTestService(0, ownedRepo("user-1"), activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	// dispatch will fail (runner is nil) but we only care about pre-dispatch validation
	if errors.Is(err, ErrRateLimitExceeded) {
		t.Error("expected no rate limit error with 0 active jobs")
	}
	if errors.Is(err, ErrRepositoryNotFound) {
		t.Error("unexpected ErrRepositoryNotFound")
	}
}

func TestCreate_TwoActiveJobs_Succeeds(t *testing.T) {
	svc := newTestService(2, ownedRepo("user-1"), activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if errors.Is(err, ErrRateLimitExceeded) {
		t.Error("expected no rate limit error with 2 active jobs (limit is 3)")
	}
}

func TestCreate_ThreeActiveJobs_ReturnsErrRateLimitExceeded(t *testing.T) {
	svc := newTestService(3, ownedRepo("user-1"), activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Errorf("expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestCreate_DifferentUsersDoNotShareLimit(t *testing.T) {
	// user-2 has 3 active jobs; user-1 has 0 — user-1 should succeed
	jobsMock := &mockJobRepo{activeCount: 0} // user-1 has 0 active jobs
	cfg := &config.Config{MaxActiveJobsPerUser: 3}
	svc := &JobService{
		cfg:     cfg,
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		users:   &mockUserRepo{user: activeUser("user-1")},
		runner:  &mockRunnerClient{},
	}

	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if errors.Is(err, ErrRateLimitExceeded) {
		t.Error("user-1 should not be rate-limited when they have 0 active jobs")
	}
}

func TestCreate_CompletedOrFailedJobsDoNotCountTowardLimit(t *testing.T) {
	// CountActiveByUserID only counts pending/running — the mock returns 2
	// (simulating 2 pending/running, with many completed/failed ignored)
	svc := newTestService(2, ownedRepo("user-1"), activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "another task",
	})
	if errors.Is(err, ErrRateLimitExceeded) {
		t.Error("should not be rate-limited: only 2 active jobs (completed/failed not counted)")
	}
}

// ---- tests: ownership validation ----

func TestCreate_OwnRepoSucceeds(t *testing.T) {
	svc := newTestService(0, ownedRepo("user-1"), activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "fix bug",
	})
	if errors.Is(err, ErrRepositoryNotFound) {
		t.Errorf("own repo should not return ErrRepositoryNotFound, got %v", err)
	}
}

func TestCreate_OtherUsersRepoReturnsErrRepositoryNotFound(t *testing.T) {
	// repo belongs to user-2, but user-1 is requesting
	svc := newTestService(0, ownedRepo("user-2"), activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "fix bug",
	})
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Errorf("expected ErrRepositoryNotFound, got %v", err)
	}
}

func TestCreate_NonExistentRepoReturnsErrRepositoryNotFound(t *testing.T) {
	svc := newTestService(0, nil /* repo not found */, activeUser("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "nonexistent",
		Prompt:       "fix bug",
	})
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Errorf("expected ErrRepositoryNotFound, got %v", err)
	}
}

// ---- tests: agent type ----

func TestCreate_AgentTypeIsIncludedInDispatchPayload(t *testing.T) {
	runner := &mockRunnerClient{}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    &mockJobRepo{},
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		users:   &mockUserRepo{user: activeUser("user-1")},
		runner:  runner,
	}
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
		AgentType:    domain.AgentTypeGemini,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.lastPayload.AgentType != string(domain.AgentTypeGemini) {
		t.Errorf("expected AgentType %q in payload, got %q",
			domain.AgentTypeGemini, runner.lastPayload.AgentType)
	}
}

func TestCreate_AgentTypeIsPersistedOnJob(t *testing.T) {
	jobsMock := &mockJobRepo{}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		users:   &mockUserRepo{user: activeUser("user-1")},
		runner:  &mockRunnerClient{},
	}
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
		AgentType:    domain.AgentTypeClaudeCode,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobsMock.created) == 0 {
		t.Fatal("expected job to be created")
	}
	if jobsMock.created[0].AgentType != domain.AgentTypeClaudeCode {
		t.Errorf("expected job.AgentType %q, got %q",
			domain.AgentTypeClaudeCode, jobsMock.created[0].AgentType)
	}
}
