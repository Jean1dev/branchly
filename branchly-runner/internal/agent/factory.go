package agent

import (
	"fmt"

	"github.com/branchly/branchly-runner/internal/domain"
)

// Factory selects and returns the singleton agent for a given AgentType.
// Agents are instantiated once in main and reused across all jobs.
type Factory struct {
	claudeCodeAgent domain.Agent
	geminiAgent     domain.Agent
	codexAgent      domain.Agent
}

func NewFactory(claudeCodeAgent domain.Agent, geminiAgent domain.Agent, codexAgent domain.Agent) *Factory {
	return &Factory{
		claudeCodeAgent: claudeCodeAgent,
		geminiAgent:     geminiAgent,
		codexAgent:      codexAgent,
	}
}

func (f *Factory) Create(agentType domain.AgentType) (domain.Agent, error) {
	switch agentType {
	case domain.AgentTypeClaudeCode:
		return f.claudeCodeAgent, nil
	case domain.AgentTypeGemini:
		return f.geminiAgent, nil
	case domain.AgentTypeGPTCodex:
		return f.codexAgent, nil
	default:
		return nil, fmt.Errorf("agent factory: unknown agent type: %s", agentType)
	}
}
