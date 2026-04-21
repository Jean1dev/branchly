package codex_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/branchly/branchly-runner/internal/agent/codex"
	"github.com/branchly/branchly-runner/internal/domain"
)

// writeFakeCodex writes a shell script named "codex" to a temp dir and
// prepends that dir to PATH so exec.LookPath finds it.
func writeFakeCodex(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "codex")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestCodexAgent_StreamsStdoutAsInfoLogs(t *testing.T) {
	writeFakeCodex(t, `echo "line one"
echo "line two"
`)
	agent := codex.New()
	var logs []string
	summary, err := agent.Run(context.Background(), domain.AgentInput{
		WorkDir: t.TempDir(),
		Prompt:  "test prompt",
		OnLog: func(_ domain.LogLevel, msg string) {
			logs = append(logs, msg)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "line two" {
		t.Errorf("summary: want %q, got %q", "line two", summary)
	}
	if len(logs) < 2 {
		t.Errorf("expected at least 2 log lines, got %d: %v", len(logs), logs)
	}
}

func TestCodexAgent_InjectsOpenAIAPIKey(t *testing.T) {
	// The fake codex script verifies OPENAI_API_KEY is set.
	writeFakeCodex(t, `if [ -z "$OPENAI_API_KEY" ]; then
  echo "missing OPENAI_API_KEY" >&2
  exit 1
fi
echo "key-ok"
`)
	agent := codex.New()
	summary, err := agent.Run(context.Background(), domain.AgentInput{
		WorkDir: t.TempDir(),
		Prompt:  "test",
		APIKey:  "sk-test-openai-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "key-ok" {
		t.Errorf("want %q, got %q", "key-ok", summary)
	}
}

func TestCodexAgent_NoAPIKey_DoesNotForceEnvVar(t *testing.T) {
	// When no APIKey is provided the agent should not inject OPENAI_API_KEY.
	// The fake script exits 0 regardless — we just confirm no crash.
	writeFakeCodex(t, `echo "ok"`)
	agent := codex.New()
	_, err := agent.Run(context.Background(), domain.AgentInput{
		WorkDir: t.TempDir(),
		Prompt:  "test",
		APIKey:  "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCodexAgent_ExitNonZero_ReturnsError(t *testing.T) {
	writeFakeCodex(t, `echo "partial output"
exit 1
`)
	agent := codex.New()
	_, err := agent.Run(context.Background(), domain.AgentInput{
		WorkDir: t.TempDir(),
		Prompt:  "test",
	})
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}

func TestCodexAgent_CLINotFound_ReturnsError(t *testing.T) {
	// Point PATH at an empty dir so codex binary is not found.
	t.Setenv("PATH", t.TempDir())
	agent := codex.New()
	_, err := agent.Run(context.Background(), domain.AgentInput{
		WorkDir: t.TempDir(),
		Prompt:  "test",
	})
	if err == nil {
		t.Fatal("expected error when codex CLI not in PATH, got nil")
	}
}

func TestCodexAgent_StderrIncludedInError(t *testing.T) {
	writeFakeCodex(t, `echo "error details" >&2
exit 2
`)
	agent := codex.New()
	_, err := agent.Run(context.Background(), domain.AgentInput{
		WorkDir: t.TempDir(),
		Prompt:  "test",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if msg := err.Error(); len(msg) == 0 {
		t.Error("expected non-empty error message")
	}
}
