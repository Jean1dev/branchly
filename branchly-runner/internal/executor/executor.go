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

	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/infra"
	"github.com/branchly/branchly-runner/internal/repository"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v63/github"
)

type RunJobInput struct {
	JobID            string
	UserID           string
	RepositoryName   string
	DefaultBranch    string
	BranchName       string
	Prompt           string
	EncryptedToken   string
}

type Executor struct {
	agent    domain.Agent
	jobs     *repository.JobRepository
	jobLogs  *repository.JobLogRepository
	encKey   []byte
	workDir  string
	appendMu sync.Mutex
}

func NewExecutor(agent domain.Agent, jobs *repository.JobRepository, jobLogs *repository.JobLogRepository, encKey []byte, workDir string) *Executor {
	return &Executor{agent: agent, jobs: jobs, jobLogs: jobLogs, encKey: encKey, workDir: workDir}
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

func splitRepo(full string) (owner, name string, err error) {
	parts := strings.Split(strings.TrimSpace(full), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository name")
	}
	return parts[0], parts[1], nil
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
	vctx, vcancel := persistCtx()
	err := e.jobs.VerifyJobOwner(vctx, in.JobID, in.UserID)
	vcancel()
	if err != nil {
		slog.Error("job verification failed", "job_id", in.JobID, "user_id", in.UserID, "error", err)
		return
	}
	slog.Info("job execution started", "job_id", in.JobID, "repository", in.RepositoryName)
	token, err := infra.Decrypt(in.EncryptedToken, e.encKey)
	if err != nil {
		e.markFailed(in.JobID, in.BranchName, "could not decrypt repository credentials")
		return
	}
	baseBranch := strings.TrimSpace(in.DefaultBranch)
	if baseBranch == "" {
		baseBranch = "main"
	}
	dir, err := jobScratchDir(e.workDir, in.JobID)
	if err != nil {
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("temp dir: %v", err))
		return
	}
	defer func() { _ = os.RemoveAll(dir) }()

	e.appendLog(in.JobID, domain.LogLevelInfo, "Cloning repository…")
	auth := &githttp.BasicAuth{Username: "git", Password: token}
	cloneCtx, cloneCancel := context.WithTimeout(ctx, 60*time.Second)
	defer cloneCancel()
	_, err = git.PlainCloneContext(cloneCtx, dir, false, &git.CloneOptions{
		URL:             fmt.Sprintf("https://github.com/%s.git", in.RepositoryName),
		ReferenceName:   plumbing.NewBranchReferenceName(baseBranch),
		SingleBranch:    true,
		Depth:           1,
		Auth:            auth,
		Progress:        nil,
	})
	if err != nil {
		slog.Warn("git clone failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("git clone failed: %v", err))
		return
	}
	slog.Info("repository cloned", "job_id", in.JobID)
	repo, err := git.PlainOpen(dir)
	if err != nil {
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("open repo: %v", err))
		return
	}
	wt, err := repo.Worktree()
	if err != nil {
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("worktree: %v", err))
		return
	}
	e.appendLog(in.JobID, domain.LogLevelInfo, fmt.Sprintf("Creating branch %s", in.BranchName))
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(in.BranchName),
		Create: true,
	}); err != nil {
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("create branch: %v", err))
		return
	}
	e.appendLog(in.JobID, domain.LogLevelInfo, "Running agent…")
	agentCtx, agentCancel := context.WithTimeout(ctx, 30*time.Minute)
	defer agentCancel()
	summary, err := e.agent.Run(agentCtx, domain.AgentInput{
		WorkDir:    dir,
		Prompt:     in.Prompt,
		RepoName:   in.RepositoryName,
		BranchName: in.BranchName,
		OnLog: func(level domain.LogLevel, message string) {
			e.appendLog(in.JobID, level, message)
		},
	})
	if err != nil {
		slog.Warn("agent failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("agent failed: %v", err))
		return
	}
	slog.Info("agent finished", "job_id", in.JobID)
	st, err := wt.Status()
	if err != nil {
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("git status: %v", err))
		return
	}
	if st.IsClean() {
		e.markFailed(in.JobID, in.BranchName, "no file changes to commit after agent run")
		return
	}
	if _, err := wt.Add("."); err != nil {
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("git add: %v", err))
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
		if err == git.ErrEmptyCommit {
			e.markFailed(in.JobID, in.BranchName, "nothing to commit")
			return
		}
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("git commit: %v", err))
		return
	}
	e.appendLog(in.JobID, domain.LogLevelInfo, "Pushing branch…")
	pushCtx, pushCancel := context.WithTimeout(ctx, 30*time.Second)
	defer pushCancel()
	err = repo.PushContext(pushCtx, &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", in.BranchName, in.BranchName)),
		},
	})
	if err != nil {
		slog.Warn("git push failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("git push failed: %v", err))
		return
	}
	slog.Info("branch pushed", "job_id", in.JobID)
	owner, repoName, err := splitRepo(in.RepositoryName)
	if err != nil {
		e.markFailed(in.JobID, in.BranchName, err.Error())
		return
	}
	e.appendLog(in.JobID, domain.LogLevelInfo, "Opening pull request…")
	prCtx, prCancel := context.WithTimeout(ctx, 15*time.Second)
	defer prCancel()
	gh := github.NewClient(nil).WithAuthToken(token)
	title := "Branchly: " + truncateRunes(in.Prompt, 60)
	body := summary
	head := in.BranchName
	baseRef := baseBranch
	pr, _, err := gh.PullRequests.Create(prCtx, owner, repoName, &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &baseRef,
		Body:  &body,
	})
	if err != nil {
		slog.Warn("open pull request failed", "job_id", in.JobID, "error", err)
		e.markFailed(in.JobID, in.BranchName, fmt.Sprintf("open pull request: %v", err))
		return
	}
	prURL := pr.GetHTMLURL()
	slog.Info("job completed", "job_id", in.JobID, "pr_url", prURL)
	e.markCompleted(in.JobID, in.BranchName, prURL, summary)
}
