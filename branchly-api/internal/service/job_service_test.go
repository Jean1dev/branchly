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
	activeCount    int64
	created        []*domain.Job
	findByIDForUser *domain.Job
	resetForRetryCalled bool
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
	return m.findByIDForUser, nil
}
func (m *mockJobRepo) ResetForRetry(_ context.Context, _ string) error {
	m.resetForRetryCalled = true
	return nil
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

func (m *mockRepoRepo) Create(_ context.Context, _ *domain.Repository) error { return nil }
func (m *mockRepoRepo) FindByID(_ context.Context, _ string) (*domain.Repository, error) {
	return m.repo, nil
}
func (m *mockRepoRepo) FindByUserID(_ context.Context, _ string) ([]*domain.Repository, error) {
	return nil, nil
}
func (m *mockRepoRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockRepoRepo) FindByUserExternalAndProvider(_ context.Context, _ string, _ string, _ domain.GitProvider) (*domain.Repository, error) {
	return nil, nil
}
func (m *mockRepoRepo) FindByIntegrationID(_ context.Context, _ string) ([]*domain.Repository, error) {
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
) *JobService {
	cfg := &config.Config{MaxActiveJobsPerUser: 3}
	jobs := &mockJobRepo{activeCount: activeJobs}
	repos := &mockRepoRepo{repo: repo}
	svc := &JobService{
		cfg:     cfg,
		jobs:    jobs,
		jobLogs: &mockJobLogRepo{},
		repos:   repos,
		runner:  &mockRunnerClient{},
	}
	return svc
}

func ownedRepo(userID string) *domain.Repository {
	return &domain.Repository{
		ID:            "repo-1",
		UserID:        userID,
		FullName:      "owner/repo",
		IntegrationID: "integ-1",
		Provider:      domain.GitProviderGitHub,
	}
}

// ---- tests: rate limiting ----

func TestCreate_ZeroActiveJobs_Succeeds(t *testing.T) {
	svc := newTestService(0, ownedRepo("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if errors.Is(err, ErrRateLimitExceeded) {
		t.Error("expected no rate limit error with 0 active jobs")
	}
	if errors.Is(err, ErrRepositoryNotFound) {
		t.Error("unexpected ErrRepositoryNotFound")
	}
}

func TestCreate_TwoActiveJobs_Succeeds(t *testing.T) {
	svc := newTestService(2, ownedRepo("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if errors.Is(err, ErrRateLimitExceeded) {
		t.Error("expected no rate limit error with 2 active jobs (limit is 3)")
	}
}

func TestCreate_ThreeActiveJobs_ReturnsErrRateLimitExceeded(t *testing.T) {
	svc := newTestService(3, ownedRepo("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Errorf("expected ErrRateLimitExceeded, got %v", err)
	}
}

func TestCreate_DifferentUsersDoNotShareLimit(t *testing.T) {
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    &mockJobRepo{activeCount: 0},
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
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
	svc := newTestService(2, ownedRepo("user-1"))
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
	svc := newTestService(0, ownedRepo("user-1"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "fix bug",
	})
	if errors.Is(err, ErrRepositoryNotFound) {
		t.Errorf("own repo should not return ErrRepositoryNotFound, got %v", err)
	}
}

func TestCreate_OtherUsersRepoReturnsErrRepositoryNotFound(t *testing.T) {
	svc := newTestService(0, ownedRepo("user-2"))
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "fix bug",
	})
	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Errorf("expected ErrRepositoryNotFound, got %v", err)
	}
}

func TestCreate_NonExistentRepoReturnsErrRepositoryNotFound(t *testing.T) {
	svc := newTestService(0, nil /* repo not found */)
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

// ---- helpers for retry tests ----

func failedJob(userID string) *domain.Job {
	return &domain.Job{
		ID:           "job-1",
		UserID:       userID,
		RepositoryID: "repo-1",
		Status:       domain.JobStatusFailed,
		AgentType:    domain.AgentTypeClaudeCode,
		Prompt:       "fix bug",
	}
}

func retryingJob(userID string) *domain.Job {
	return &domain.Job{
		ID:     "job-1",
		UserID: userID,
		Status: domain.JobStatusRetrying,
	}
}

func completedJob(userID string) *domain.Job {
	return &domain.Job{
		ID:     "job-1",
		UserID: userID,
		Status: domain.JobStatusCompleted,
	}
}

// ---- tests: Retry ----

func TestRetry_FailedJob_ReturnsJob(t *testing.T) {
	jobsMock := &mockJobRepo{findByIDForUser: failedJob("user-1")}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		runner:  &mockRunnerClient{},
	}
	_, err := svc.Retry(context.Background(), "user-1", "job-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !jobsMock.resetForRetryCalled {
		t.Error("expected ResetForRetry to be called")
	}
}

func TestRetry_CompletedJob_ReturnsErrNotRetryable(t *testing.T) {
	jobsMock := &mockJobRepo{findByIDForUser: completedJob("user-1")}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		runner:  &mockRunnerClient{},
	}
	_, err := svc.Retry(context.Background(), "user-1", "job-1")
	if !errors.Is(err, domain.ErrJobNotRetryable) {
		t.Errorf("expected ErrJobNotRetryable, got %v", err)
	}
}

func TestRetry_RetryingJob_ReturnsErrNotRetryable(t *testing.T) {
	jobsMock := &mockJobRepo{findByIDForUser: retryingJob("user-1")}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		runner:  &mockRunnerClient{},
	}
	_, err := svc.Retry(context.Background(), "user-1", "job-1")
	if !errors.Is(err, domain.ErrJobNotRetryable) {
		t.Errorf("expected ErrJobNotRetryable, got %v", err)
	}
}

func TestRetry_OtherUserJob_ReturnsErrJobNotFound(t *testing.T) {
	// FindByIDForUser returns nil when the job doesn't belong to the user
	jobsMock := &mockJobRepo{findByIDForUser: nil}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-2")},
		runner:  &mockRunnerClient{},
	}
	_, err := svc.Retry(context.Background(), "user-1", "job-1")
	if !errors.Is(err, ErrJobNotFound) {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

func TestRetry_ResetsAttemptNumber(t *testing.T) {
	jobsMock := &mockJobRepo{findByIDForUser: failedJob("user-1")}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
		runner:  &mockRunnerClient{},
	}
	_, err := svc.Retry(context.Background(), "user-1", "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !jobsMock.resetForRetryCalled {
		t.Error("expected ResetForRetry to be called to reset attempt_number")
	}
}

func TestCreate_IntegrationIDInDispatchPayload(t *testing.T) {
	runner := &mockRunnerClient{}
	repo := ownedRepo("user-1")
	repo.IntegrationID = "integ-abc"
	repo.Provider = domain.GitProviderGitHub
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    &mockJobRepo{},
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: repo},
		runner:  runner,
	}
	_, err := svc.Create(context.Background(), "user-1", CreateJobInput{
		RepositoryID: "repo-1",
		Prompt:       "add feature",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.lastPayload.IntegrationID != "integ-abc" {
		t.Errorf("expected IntegrationID %q, got %q", "integ-abc", runner.lastPayload.IntegrationID)
	}
	if runner.lastPayload.Provider != string(domain.GitProviderGitHub) {
		t.Errorf("expected Provider %q, got %q", domain.GitProviderGitHub, runner.lastPayload.Provider)
	}
}

func TestCreate_AgentTypeIsPersistedOnJob(t *testing.T) {
	jobsMock := &mockJobRepo{}
	svc := &JobService{
		cfg:     &config.Config{MaxActiveJobsPerUser: 3},
		jobs:    jobsMock,
		jobLogs: &mockJobLogRepo{},
		repos:   &mockRepoRepo{repo: ownedRepo("user-1")},
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
