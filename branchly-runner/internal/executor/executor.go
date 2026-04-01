package executor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	agentpkg "github.com/branchly/branchly-runner/internal/agent"
	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/gitprovider"
	"github.com/branchly/branchly-runner/internal/infra"
	"github.com/branchly/branchly-runner/internal/slug"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

// jobUpdater is satisfied by *repository.JobRepository.
type jobUpdater interface {
	UpdateJobFields(ctx context.Context, id string, status domain.JobStatus, prURL string, branchName string, completedAt *time.Time) error
	VerifyJobOwner(ctx context.Context, id, userID string) error
	SetCost(ctx context.Context, id string, cost *domain.JobCost) error
}

// jobLogger is satisfied by *repository.JobLogRepository.
type jobLogger interface {
	Append(ctx context.Context, jobID string, entry domain.LogEntry) error
}

type RunJobInput struct {
	JobID          string
	UserID         string
	RepositoryID   string
	RepositoryName string
	DefaultBranch  string
	Prompt         string
	IntegrationID  string
	Provider       domain.GitProvider
	AgentType      domain.AgentType
}

type Executor struct {
	factory         *agentpkg.Factory
	providerFactory *gitprovider.Factory
	jobs            jobUpdater
	jobLogs         jobLogger
	repos           domain.RepositoryRepository
	integrations    domain.IntegrationRepository
	encKey          []byte
	workDir         string
	appendMu        sync.Mutex
}

func NewExecutor(
	factory *agentpkg.Factory,
	providerFactory *gitprovider.Factory,
	jobs jobUpdater,
	jobLogs jobLogger,
	repos domain.RepositoryRepository,
	integrations domain.IntegrationRepository,
	encKey []byte,
	workDir string,
) *Executor {
	return &Executor{
		factory:         factory,
		providerFactory: providerFactory,
		jobs:            jobs,
		jobLogs:         jobLogs,
		repos:           repos,
		integrations:    integrations,
		encKey:          encKey,
		workDir:         workDir,
	}
}

func persistCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func (e *Executor) appendLog(jobID string, lvl domain.LogLevel, msg string) {
	e.appendMu.Lock()
	defer e.appendMu.Unlock()
	ctx, cancel := persistCtx()
	defer cancel()
	if err := e.jobLogs.Append(ctx, jobID, domain.LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     lvl,
		Message:   msg,
	}); err != nil {
		slog.Warn("append job log failed", "job_id", jobID, "error", err)
	}
}

func (e *Executor) markFailed(jobID, branchName, msg string) {
	e.appendMu.Lock()
	defer e.appendMu.Unlock()
	ctx, cancel := persistCtx()
	defer cancel()
	if err := e.jobLogs.Append(ctx, jobID, domain.LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     domain.LogLevelError,
		Message:   msg,
	}); err != nil {
		slog.Warn("append job log failed", "job_id", jobID, "error", err)
	}
	t := time.Now().UTC()
	_ = e.jobs.UpdateJobFields(ctx, jobID, domain.JobStatusFailed, "", branchName, &t)
}

func (e *Executor) markCompleted(jobID, branchName, prURL, summary string) {
	e.appendMu.Lock()
	defer e.appendMu.Unlock()
	ctx, cancel := persistCtx()
	defer cancel()
	if strings.TrimSpace(summary) != "" {
		if err := e.jobLogs.Append(ctx, jobID, domain.LogEntry{
			Timestamp: time.Now().UTC(),
			Level:     domain.LogLevelSuccess,
			Message:   summary,
		}); err != nil {
			slog.Warn("append job log failed", "job_id", jobID, "error", err)
		}
	}
	t := time.Now().UTC()
	_ = e.jobs.UpdateJobFields(ctx, jobID, domain.JobStatusCompleted, prURL, branchName, &t)
}

func truncateRunes(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "…"
}

func jobScratchDir(workDir, jobID string) (string, error) {
	cfg := filepath.Clean(workDir)
	fallback := filepath.Join(os.TempDir(), "branchly-jobs")
	roots := []string{cfg}
	if fallback != cfg {
		roots = append(roots, fallback)
	}
	var lastErr error
	for _, root := range roots {
		if err := os.MkdirAll(root, 0o755); err != nil {
			lastErr = err
			continue
		}
		d, err := os.MkdirTemp(root, jobID+"-*")
		if err == nil {
			if root != cfg {
				slog.Warn("work directory not usable, using fallback",
					"configured", workDir, "used_root", root, "job_id", jobID)
			}
			return d, nil
		}
		lastErr = err
	}
	d, err := os.MkdirTemp("", "branchly-job-*")
	if err != nil {
		return "", fmt.Errorf("%w (last configured attempt: %v)", err, lastErr)
	}
	slog.Warn("using system temp for job scratch", "dir", d, "job_id", jobID, "configured", workDir)
	return d, nil
}

func (e *Executor) Run(ctx context.Context, in RunJobInput) {
	// Step 1: verify the job belongs to the stated user.
	vctx, vcancel := persistCtx()
	err := e.jobs.VerifyJobOwner(vctx, in.JobID, in.UserID)
	vcancel()
	if err != nil {
		slog.Error("job verification failed", "job_id", in.JobID, "user_id", in.UserID, "error", err)
		return
	}

	// Step 2: derive branch name from the prompt (no external I/O).
	branchName := slug.GenerateSlug(in.Prompt)

	// Step 3: validate repository ownership in the runner's own database.
	rctx, rcancel := persistCtx()
	repo, repoErr := e.repos.FindByID(rctx, in.RepositoryID)
	rcancel()
	if repoErr != nil || repo == nil || repo.UserID != in.UserID {
		slog.Error("runner: repository ownership validation failed",
			"job_id", in.JobID,
			"repository_id", in.RepositoryID,
			"user_id", in.UserID,
		)
		e.markFailed(in.JobID, branchName, "repository ownership validation failed")
		return
	}
	if repo.FullName != in.RepositoryName {
		slog.Error("runner: repository name mismatch",
			"job_id", in.JobID,
			"payload_name", in.RepositoryName,
			"db_name", repo.FullName,
		)
		e.markFailed(in.JobID, branchName, "repository name mismatch")
		return
	}

	// Step 4: resolve the agent early — fail fast before any expensive I/O.
	selectedAgent, err := e.factory.Create(in.AgentType)
	if err != nil {
		slog.Error("unknown agent type", "job_id", in.JobID, "agent_type", in.AgentType)
		e.markFailed(in.JobID, branchName, "unknown agent type: "+string(in.AgentType))
		return
	}

	// Step 5: fetch the integration and validate ownership.
	ictx, icancel := persistCtx()
	integration, integErr := e.integrations.FindByID(ictx, in.IntegrationID)
	icancel()
	if integErr != nil || integration == nil {
		slog.Error("runner: integration not found", "job_id", in.JobID, "integration_id", in.IntegrationID)
		e.markFailed(in.JobID, branchName, "integration not found")
		return
	}
	if integration.UserID != in.UserID {
		slog.Error("runner: integration ownership mismatch",
			"job_id", in.JobID,
			"integration_user_id", integration.UserID,
			"job_user_id", in.UserID,
		)
		e.markFailed(in.JobID, branchName, "integration ownership validation failed")
		return
	}

	// Step 6: decrypt token.
	token, err := infra.Decrypt(integration.EncryptedToken, e.encKey)
	integration.EncryptedToken = "" // clear from memory
	if err != nil {
		e.markFailed(in.JobID, branchName, "could not decrypt repository credentials")
		return
	}
	zeroToken := func() { token = "" }

	// Step 7: resolve provider client.
	providerClient, err := e.providerFactory.Create(in.Provider)
	if err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("unsupported provider: %s", in.Provider))
		return
	}

	slog.Info("job execution started", "job_id", in.JobID, "repository", in.RepositoryName, "provider", in.Provider)
	startedAt := time.Now()

	baseBranch := strings.TrimSpace(in.DefaultBranch)
	if baseBranch == "" {
		baseBranch = "main"
	}

	dir, err := jobScratchDir(e.workDir, in.JobID)
	if err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("temp dir: %v", err))
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	// Use clone_url from the repository document; fall back to constructing it.
	cloneURL := repo.CloneURL
	if cloneURL == "" {
		cloneURL = fmt.Sprintf("https://github.com/%s.git", in.RepositoryName)
	}

	e.appendLog(in.JobID, domain.LogLevelInfo, "Cloning repository…")
	// Both GitHub OAuth and GitLab PAT use the same BasicAuth format.
	auth := &githttp.BasicAuth{Username: "oauth2", Password: token}
	cloneCtx, cloneCancel := context.WithTimeout(ctx, 60*time.Second)
	defer cloneCancel()
	_, err = git.PlainCloneContext(cloneCtx, dir, false, &git.CloneOptions{
		URL:           cloneURL,
		ReferenceName: plumbing.NewBranchReferenceName(baseBranch),
		SingleBranch:  true,
		Depth:         1,
		Auth:          auth,
		Progress:      nil,
	})
	if err != nil {
		zeroToken()
		slog.Warn("git clone failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, branchName, fmt.Sprintf("git clone failed: %v", err))
		return
	}
	slog.Info("repository cloned", "job_id", in.JobID)

	gitRepo, err := git.PlainOpen(dir)
	if err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("open repo: %v", err))
		return
	}
	wt, err := gitRepo.Worktree()
	if err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("worktree: %v", err))
		return
	}

	e.appendLog(in.JobID, domain.LogLevelInfo, fmt.Sprintf("Creating branch %s", branchName))
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	}); err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("create branch: %v", err))
		return
	}

	e.appendLog(in.JobID, domain.LogLevelInfo, "Running agent…")
	agentCtx, agentCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer agentCancel()
	summary, err := selectedAgent.Run(agentCtx, domain.AgentInput{
		WorkDir:    dir,
		Prompt:     in.Prompt,
		RepoName:   in.RepositoryName,
		BranchName: branchName,
		OnLog: func(level domain.LogLevel, message string) {
			e.appendLog(in.JobID, level, message)
		},
	})
	if err != nil {
		zeroToken()
		slog.Warn("agent failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, branchName, fmt.Sprintf("agent failed: %v", err))
		return
	}
	slog.Info("agent finished", "job_id", in.JobID)

	st, err := wt.Status()
	if err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("git status: %v", err))
		return
	}
	if st.IsClean() {
		zeroToken()
		e.markFailed(in.JobID, branchName, "no file changes to commit after agent run")
		return
	}

	if _, err := wt.Add("."); err != nil {
		zeroToken()
		e.markFailed(in.JobID, branchName, fmt.Sprintf("git add: %v", err))
		return
	}

	e.appendLog(in.JobID, domain.LogLevelInfo, "Committing changes…")
	_, err = wt.Commit("branchly: automated changes", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Branchly",
			Email: "branchly@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		zeroToken()
		if err == git.ErrEmptyCommit {
			e.markFailed(in.JobID, branchName, "nothing to commit")
			return
		}
		e.markFailed(in.JobID, branchName, fmt.Sprintf("git commit: %v", err))
		return
	}

	e.appendLog(in.JobID, domain.LogLevelInfo, "Pushing branch…")
	pushCtx, pushCancel := context.WithTimeout(ctx, 30*time.Second)
	defer pushCancel()
	err = gitRepo.PushContext(pushCtx, &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName)),
		},
	})
	if err != nil {
		zeroToken()
		slog.Warn("git push failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, branchName, fmt.Sprintf("git push failed: %v", err))
		return
	}
	slog.Info("branch pushed", "job_id", in.JobID)

	e.appendLog(in.JobID, domain.LogLevelInfo, "Opening pull request…")
	prCtx, prCancel := context.WithTimeout(ctx, 15*time.Second)
	defer prCancel()

	prURL, err := providerClient.OpenPR(prCtx, token, domain.PROptions{
		RepoFullName: in.RepositoryName,
		Title:        "Branchly: " + truncateRunes(in.Prompt, 60),
		Body:         summary,
		Head:         branchName,
		Base:         baseBranch,
	})
	// Token is no longer needed — zero it before handling the result.
	zeroToken()
	if err != nil {
		slog.Warn("open pull request failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, branchName, fmt.Sprintf("open pull request: %v", err))
		return
	}

	slog.Info("job completed", "job_id", in.JobID, "pr_url", prURL)
	e.markCompleted(in.JobID, branchName, prURL, summary)

	cost := estimateCost(time.Since(startedAt), in.AgentType)
	cctx, ccancel := persistCtx()
	defer ccancel()
	if err := e.jobs.SetCost(cctx, in.JobID, cost); err != nil {
		slog.Warn("set job cost failed", "job_id", in.JobID, "error", err)
	}
}
