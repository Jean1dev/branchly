package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/executor"
	"github.com/branchly/branchly-runner/internal/pool"
)

// retryJobRepo is the subset of the job repository used by the poller.
type retryJobRepo interface {
	FindDueForRetry(ctx context.Context) ([]*domain.Job, error)
}

// retryRepoRepo is the subset of the repository repository used by the poller.
type retryRepoRepo interface {
	FindByID(ctx context.Context, id string) (*domain.Repository, error)
}

// RetryPoller periodically finds jobs in "retrying" status whose next_retry_at
// has passed and re-submits them for execution.
type RetryPoller struct {
	jobRepo  retryJobRepo
	repoRepo retryRepoRepo
	ex       *executor.Executor
	pool     *pool.Pool
	interval time.Duration
}

func NewRetryPoller(jobRepo retryJobRepo, repoRepo retryRepoRepo, ex *executor.Executor, p *pool.Pool) *RetryPoller {
	return &RetryPoller{
		jobRepo:  jobRepo,
		repoRepo: repoRepo,
		ex:       ex,
		pool:     p,
		interval: 30 * time.Second,
	}
}

func (p *RetryPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *RetryPoller) poll(ctx context.Context) {
	jobs, err := p.jobRepo.FindDueForRetry(ctx)
	if err != nil {
		slog.Error("retry poller: find due jobs", "error", err)
		return
	}
	for _, job := range jobs {
		job := job // capture loop var

		// Look up the connected repository to get FullName, IntegrationID, Provider, DefaultBranch.
		repo, err := p.repoRepo.FindByID(ctx, job.RepositoryID)
		if err != nil || repo == nil {
			slog.Error("retry poller: repository lookup failed",
				"job_id", job.ID,
				"repository_id", job.RepositoryID,
				"error", err,
			)
			continue
		}

		in := executor.RunJobInput{
			JobID:          job.ID,
			UserID:         job.UserID,
			RepositoryID:   job.RepositoryID,
			RepositoryName: repo.FullName,
			DefaultBranch:  repo.DefaultBranch,
			Prompt:         job.Prompt,
			IntegrationID:  repo.IntegrationID,
			Provider:       repo.Provider,
			AgentType:      job.AgentType,
			AttemptNumber:  job.AttemptNumber,
			MaxAttempts:    job.MaxAttempts,
		}

		submitted := p.pool.TryGo(func() {
			p.ex.Run(ctx, in)
		})
		if !submitted {
			slog.Warn("retry poller: pool full, job will retry later", "job_id", job.ID)
		}
	}
}
