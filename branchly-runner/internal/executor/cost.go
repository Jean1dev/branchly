package executor

import (
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

// Pricing constants (USD per 1M tokens).
const (
	claudeModel        = "claude-sonnet-4-6"
	geminiModel        = "gemini-2.0-flash"
	codexModel         = "codex-1"
	pricePerMInputUSD  = 3.00
	pricePerMOutputUSD = 15.00

	// Approximate token throughput observed for Claude Code executions.
	inputTokensPerSec  = 800.0
	outputTokensPerSec = 200.0

	// GPT Codex (codex-1) pricing: $30/1M input, $120/1M output.
	codexPricePerMInputUSD  = 30.00
	codexPricePerMOutputUSD = 120.00
)

// estimateCost produces a cost estimate based on execution duration.
// For Claude Code, token counts are derived from observed throughput rates.
// For Gemini, cost is zero (free tier).
func estimateCost(duration time.Duration, agentType domain.AgentType) *domain.JobCost {
	secs := duration.Seconds()
	if secs <= 0 {
		secs = 0
	}

	base := &domain.JobCost{
		AgentType:    agentType,
		DurationSecs: secs,
	}

	switch agentType {
	case domain.AgentTypeClaudeCode:
		inputTokens := int64(secs * inputTokensPerSec)
		outputTokens := int64(secs * outputTokensPerSec)
		base.InputTokens = inputTokens
		base.OutputTokens = outputTokens
		base.TotalTokens = inputTokens + outputTokens
		base.EstimatedUSD = float64(inputTokens)/1_000_000*pricePerMInputUSD + float64(outputTokens)/1_000_000*pricePerMOutputUSD
		base.ModelUsed = claudeModel
	case domain.AgentTypeGemini:
		// Gemini CLI free tier — no token-based cost
		base.ModelUsed = geminiModel
	case domain.AgentTypeGPTCodex:
		inputTokens := int64(secs * inputTokensPerSec)
		outputTokens := int64(secs * outputTokensPerSec)
		base.InputTokens = inputTokens
		base.OutputTokens = outputTokens
		base.TotalTokens = inputTokens + outputTokens
		base.EstimatedUSD = float64(inputTokens)/1_000_000*codexPricePerMInputUSD + float64(outputTokens)/1_000_000*codexPricePerMOutputUSD
		base.ModelUsed = codexModel
	}

	return base
}
