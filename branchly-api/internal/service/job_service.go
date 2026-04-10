package service

import (
	"context"
	"fmt"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/infra"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// runnerDispatcher abstracts the HTTP call to the runner, making it injectable in tests.
type runnerDispatcher interface {
	DispatchJob(ctx context.Context, payload infra.DispatchJobPayload) error
}

type JobService struct {
	cfg     *config.Config
	jobs    domain.JobRepository
	jobLogs domain.JobLogRepository
	repos   domain.ConnectedRepositoryRepository
	runner  runnerDispatcher
}

func NewJobService(cfg *config.Config, jobs domain.JobRepository, jobLogs domain.JobLogRepository, repos domain.ConnectedRepositoryRepository, runner *infra.RunnerClient) *JobService {
	return &JobService{cfg: cfg, jobs: jobs, jobLogs: jobLogs, repos: repos, runner: runner}
}

type CreateJobInput struct {
	RepositoryID string
	Prompt       string
	AgentType    domain.AgentType
	ParentJobID  string // optional — if set, this job is a thread continuation
}

func (s *JobService) Create(ctx context.Context, userID string, in CreateJobInput) (*domain.Job, error) {
	// Validate repository ownership — returns ErrRepositoryNotFound for both
	// "not found" and "wrong owner" to avoid leaking existence.
	repo, err := s.repos.FindByID(ctx, in.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("job service: create: repo: %w", err)
	}
	if repo == nil || repo.UserID != userID {
		return nil, ErrRepositoryNotFound
	}

	// Enforce per-user active-job rate limit.
	active, err := s.jobs.CountActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: create: rate limit check: %w", err)
	}
	if active >= int64(s.cfg.MaxActiveJobsPerUser) {
		return nil, ErrRateLimitExceeded
	}

	now := time.Now().UTC()
	job := &domain.Job{
		ID:           uuid.New().String(),
		UserID:       userID,
		RepositoryID: repo.ID,
		Prompt:       in.Prompt,
		AgentType:    in.AgentType,
		KeyProvider:  domain.RequiredKeyProvider(in.AgentType),
		Status:       domain.JobStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Resolve thread context when this is a continuation job.
	var parentBranchName string
	var parentPRUrl string
	if in.ParentJobID != "" {
		parent, err := s.jobs.FindByIDForUser(ctx, in.ParentJobID, userID)
		if err != nil {
			return nil, fmt.Errorf("job service: create: find parent: %w", err)
		}
		if parent == nil {
			return nil, ErrJobNotFound
		}
		// Inherit thread identity from parent (first job sets thread_id = its own ID).
		threadID := parent.ThreadID
		if threadID == "" {
			threadID = parent.ID
		}
		job.ThreadID = threadID
		job.ParentJobID = parent.ID
		job.ThreadPosition = parent.ThreadPosition + 1
		// Continue work on the same branch so commits stack cleanly.
		parentBranchName = parent.BranchName
		// Inherit the existing PR URL so the runner doesn't open a duplicate.
		parentPRUrl = parent.PRUrl
	} else {
		// Root job: thread_id equals its own ID so FindByThreadID works uniformly.
		job.ThreadID = job.ID
	}

	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("job service: create: insert: %w", err)
	}

	dispatchCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	err = s.runner.DispatchJob(dispatchCtx, infra.DispatchJobPayload{
		JobID:          job.ID,
		UserID:         userID,
		RepositoryID:   repo.ID,
		RepositoryName: repo.FullName,
		DefaultBranch:  repo.DefaultBranch,
		Prompt:         job.Prompt,
		IntegrationID:  repo.IntegrationID,
		Provider:       string(repo.Provider),
		AgentType:      string(job.AgentType),
		KeyProvider:    string(job.KeyProvider),
		ThreadID:       job.ThreadID,
		ParentJobID:    job.ParentJobID,
		BranchName:     parentBranchName,
		ParentPRUrl:    parentPRUrl,
	})
	if err != nil {
		_ = s.jobs.UpdateJobFields(ctx, job.ID, domain.JobStatusFailed, "", "", ptrTime(time.Now().UTC()))
		return nil, fmt.Errorf("job service: create: dispatch: %w", err)
	}
	if err := s.jobs.UpdateStatus(ctx, job.ID, domain.JobStatusRunning); err != nil {
		return nil, fmt.Errorf("job service: create: set running: %w", err)
	}
	job.Status = domain.JobStatusRunning
	return job, nil
}

// GetThread returns all jobs that belong to the same thread as jobID, ordered
// by thread_position ascending. Legacy jobs without a thread_id are returned
// as a single-element slice.
func (s *JobService) GetThread(ctx context.Context, userID, jobID string) ([]*domain.Job, error) {
	job, err := s.jobs.FindByIDForUser(ctx, jobID, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: get thread: find root: %w", err)
	}
	if job == nil {
		return nil, ErrJobNotFound
	}
	threadID := job.ThreadID
	if threadID == "" {
		// Legacy job created before thread support — return as single-item list.
		return []*domain.Job{job}, nil
	}
	jobs, err := s.jobs.FindByThreadID(ctx, threadID, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: get thread: list: %w", err)
	}
	return jobs, nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func (s *JobService) JobMeta(ctx context.Context, userID, jobID string) (*domain.Job, error) {
	return s.jobs.FindByIDForUser(ctx, jobID, userID)
}

func (s *JobService) Get(ctx context.Context, userID, jobID string) (*domain.Job, error) {
	j, err := s.jobs.FindByIDForUser(ctx, jobID, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: get: %w", err)
	}
	if j == nil {
		return nil, nil
	}
	const tail = 5000
	rows, err := s.jobLogs.ListTailByJobID(ctx, jobID, tail)
	if err != nil {
		return nil, fmt.Errorf("job service: get logs: %w", err)
	}
	if len(rows) > 0 {
		j.Logs = make([]domain.LogEntry, len(rows))
		for i := range rows {
			j.Logs[i] = rows[i].Entry
		}
	}
	return j, nil
}

func (s *JobService) List(ctx context.Context, userID string, status *domain.JobStatus, repositoryID *string) ([]*domain.Job, error) {
	list, err := s.jobs.FindByUserID(ctx, userID, status, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("job service: list: %w", err)
	}
	return list, nil
}

func (s *JobService) UpdateStatusInternal(ctx context.Context, jobID string, status domain.JobStatus, prURL string, branchName string) error {
	var completed *time.Time
	if status == domain.JobStatusCompleted || status == domain.JobStatusFailed {
		t := time.Now().UTC()
		completed = &t
	}
	if err := s.jobs.UpdateJobFields(ctx, jobID, status, prURL, branchName, completed); err != nil {
		return fmt.Errorf("job service: internal status: %w", err)
	}
	return nil
}

func (s *JobService) Retry(ctx context.Context, userID string, jobID string) (*domain.Job, error) {
	job, err := s.jobs.FindByIDForUser(ctx, jobID, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: retry: find: %w", err)
	}
	if job == nil {
		return nil, ErrJobNotFound
	}

	// Only permanently failed jobs support manual retry.
	if job.Status != domain.JobStatusFailed {
		return nil, domain.ErrJobNotRetryable
	}

	// Look up the repository so we can re-dispatch with all required fields.
	repo, err := s.repos.FindByID(ctx, job.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("job service: retry: find repo: %w", err)
	}
	if repo == nil || repo.UserID != userID {
		return nil, ErrRepositoryNotFound
	}

	if err := s.jobs.ResetForRetry(ctx, jobID); err != nil {
		return nil, fmt.Errorf("job service: retry: reset: %w", err)
	}

	dispatchCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	err = s.runner.DispatchJob(dispatchCtx, infra.DispatchJobPayload{
		JobID:          job.ID,
		UserID:         userID,
		RepositoryID:   repo.ID,
		RepositoryName: repo.FullName,
		DefaultBranch:  repo.DefaultBranch,
		Prompt:         job.Prompt,
		IntegrationID:  repo.IntegrationID,
		Provider:       string(repo.Provider),
		AgentType:      string(job.AgentType),
		KeyProvider:    string(job.KeyProvider),
	})
	if err != nil {
		_ = s.jobs.UpdateJobFields(ctx, job.ID, domain.JobStatusFailed, "", "", ptrTime(time.Now().UTC()))
		return nil, fmt.Errorf("job service: retry: dispatch: %w", err)
	}
	if err := s.jobs.UpdateStatus(ctx, job.ID, domain.JobStatusRunning); err != nil {
		return nil, fmt.Errorf("job service: retry: set running: %w", err)
	}

	return s.jobs.FindByIDForUser(ctx, jobID, userID)
}

func (s *JobService) AppendLogInternal(ctx context.Context, jobID string, entry domain.LogEntry) error {
	if err := s.jobLogs.Append(ctx, jobID, entry); err != nil {
		return fmt.Errorf("job service: internal log: %w", err)
	}
	return nil
}

func (s *JobService) ListJobLogsAsc(ctx context.Context, userID, jobID string, limit int) ([]domain.StoredJobLog, error) {
	j, err := s.jobs.FindByIDForUser(ctx, jobID, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: list logs auth: %w", err)
	}
	if j == nil {
		return nil, nil
	}
	return s.jobLogs.ListByJobID(ctx, jobID, limit)
}

func (s *JobService) ListJobLogsAfter(ctx context.Context, userID, jobID string, after primitive.ObjectID, limit int) ([]domain.StoredJobLog, *domain.Job, error) {
	j, err := s.jobs.FindByIDForUser(ctx, jobID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("job service: list logs after auth: %w", err)
	}
	if j == nil {
		return nil, nil, nil
	}
	rows, err := s.jobLogs.ListByJobIDAfter(ctx, jobID, after, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("job service: list logs after: %w", err)
	}
	return rows, j, nil
}
