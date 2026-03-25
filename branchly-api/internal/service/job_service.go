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

type JobService struct {
	cfg     *config.Config
	jobs    domain.JobRepository
	jobLogs domain.JobLogRepository
	repos   domain.ConnectedRepositoryRepository
	users   domain.UserRepository
	runner  *infra.RunnerClient
}

func NewJobService(cfg *config.Config, jobs domain.JobRepository, jobLogs domain.JobLogRepository, repos domain.ConnectedRepositoryRepository, users domain.UserRepository, runner *infra.RunnerClient) *JobService {
	return &JobService{cfg: cfg, jobs: jobs, jobLogs: jobLogs, repos: repos, users: users, runner: runner}
}


type CreateJobInput struct {
	RepositoryID string
	Prompt       string
}

func (s *JobService) Create(ctx context.Context, userID string, in CreateJobInput) (*domain.Job, error) {
	repo, err := s.repos.FindByID(ctx, in.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("job service: create: repo: %w", err)
	}
	if repo == nil || repo.UserID != userID {
		return nil, ErrNotFound
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("job service: create: user: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	now := time.Now().UTC()
	job := &domain.Job{
		ID:           uuid.New().String(),
		UserID:       userID,
		RepositoryID: repo.ID,
		Prompt:       in.Prompt,
		Status:       domain.JobStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("job service: create: insert: %w", err)
	}
	dispatchCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	err = s.runner.DispatchJob(dispatchCtx, infra.DispatchJobPayload{
		JobID:          job.ID,
		UserID:         userID,
		RepositoryName: repo.FullName,
		DefaultBranch:  repo.DefaultBranch,
		Prompt:         job.Prompt,
		EncryptedToken: u.EncryptedToken,
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
