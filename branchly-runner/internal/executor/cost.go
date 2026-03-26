package executor

import (
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

// Sonnet 4.6 pricing (USD per 1M tokens).
const (
	modelUsed          = "claude-sonnet-4-6"
	pricePerMInputUSD  = 3.00
	pricePerMOutputUSD = 15.00

	// Approximate token throughput observed for Claude Code executions.
	inputTokensPerSec  = 800.0
	outputTokensPerSec = 200.0
)

// estimateCost produces a cost estimate based solely on execution duration.
// Token counts are derived from observed Claude Code throughput rates.
func estimateCost(duration time.Duration) *domain.JobCost {
	secs := duration.Seconds()
	if secs <= 0 {
		secs = 0
	}

	inputTokens := int64(secs * inputTokensPerSec)
	outputTokens := int64(secs * outputTokensPerSec)
	totalTokens := inputTokens + outputTokens

	inputCost := float64(inputTokens) / 1_000_000 * pricePerMInputUSD
	outputCost := float64(outputTokens) / 1_000_000 * pricePerMOutputUSD
	estimatedUSD := inputCost + outputCost

	return &domain.JobCost{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		EstimatedUSD: estimatedUSD,
		ModelUsed:    modelUsed,
		DurationSecs: secs,
	}
}
