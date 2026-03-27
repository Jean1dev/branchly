package agent_test

import (
	"context"
	"testing"

	agentpkg "github.com/branchly/branchly-runner/internal/agent"
	"github.com/branchly/branchly-runner/internal/domain"
)

// stubAgent is a no-op domain.Agent used in factory tests.
type stubAgent struct{ name string }

func (s *stubAgent) Run(_ context.Context, _ domain.AgentInput) (string, error) { return s.name, nil }

func TestFactory_ClaudeCode_ReturnsClaudeCodeAgent(t *testing.T) {
	claude := &stubAgent{name: "claude"}
	gemini := &stubAgent{name: "gemini"}
	f := agentpkg.NewFactory(claude, gemini)

	got, err := f.Create(domain.AgentTypeClaudeCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != claude {
		t.Error("expected claude agent to be returned for AgentTypeClaudeCode")
	}
}

func TestFactory_Gemini_ReturnsGeminiAgent(t *testing.T) {
	claude := &stubAgent{name: "claude"}
	gemini := &stubAgent{name: "gemini"}
	f := agentpkg.NewFactory(claude, gemini)

	got, err := f.Create(domain.AgentTypeGemini)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != gemini {
		t.Error("expected gemini agent to be returned for AgentTypeGemini")
	}
}

func TestFactory_UnknownType_ReturnsError(t *testing.T) {
	f := agentpkg.NewFactory(&stubAgent{}, &stubAgent{})

	_, err := f.Create("openai")
	if err == nil {
		t.Error("expected error for unknown agent type, got nil")
	}
}

func TestFactory_Agents_AreSingletons(t *testing.T) {
	claude := &stubAgent{name: "claude"}
	gemini := &stubAgent{name: "gemini"}
	f := agentpkg.NewFactory(claude, gemini)

	a1, _ := f.Create(domain.AgentTypeClaudeCode)
	a2, _ := f.Create(domain.AgentTypeClaudeCode)
	if a1 != a2 {
		t.Error("factory must return the same agent instance (singleton)")
	}
}
