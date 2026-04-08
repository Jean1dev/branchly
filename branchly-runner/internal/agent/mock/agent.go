package mock

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

// Agent is a no-op agent used in dev mode to simulate execution without
// consuming real API tokens. It creates a marker file so the executor's
// git-status check finds actual changes to commit.
type Agent struct{}

func New() *Agent {
	return &Agent{}
}

func (a *Agent) Run(_ context.Context, input domain.AgentInput) (string, error) {
	if input.OnLog != nil {
		input.OnLog(domain.LogLevelInfo, "[mock agent] starting simulated run")
	}

	time.Sleep(500 * time.Millisecond)

	if input.OnLog != nil {
		input.OnLog(domain.LogLevelInfo, fmt.Sprintf("[mock agent] repo=%s branch=%s", input.RepoName, input.BranchName))
	}

	// Create a marker file so the executor finds real file changes to commit.
	markerPath := filepath.Join(input.WorkDir, ".branchly-dev-run")
	content := fmt.Sprintf("mock run — repo=%s branch=%s time=%s\n",
		input.RepoName, input.BranchName, time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(markerPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("[mock agent] could not create marker file: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	if input.OnLog != nil {
		input.OnLog(domain.LogLevelInfo, "[mock agent] simulated work complete")
	}

	return "[mock agent] job completed successfully (dev mode — no real tokens used)", nil
}
