package executor

import (
	"testing"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

func TestEstimateCost_ZeroDuration(t *testing.T) {
	c := estimateCost(0, domain.AgentTypeClaudeCode)
	if c.InputTokens != 0 || c.OutputTokens != 0 || c.TotalTokens != 0 {
		t.Errorf("expected zero tokens for zero duration, got %+v", c)
	}
	if c.EstimatedUSD != 0 {
		t.Errorf("expected zero cost for zero duration, got %v", c.EstimatedUSD)
	}
	if c.ModelUsed != claudeModel {
		t.Errorf("expected model %q, got %q", claudeModel, c.ModelUsed)
	}
	if c.AgentType != domain.AgentTypeClaudeCode {
		t.Errorf("expected agent type %q, got %q", domain.AgentTypeClaudeCode, c.AgentType)
	}
}

func TestEstimateCost_OneMinute_ClaudeCode(t *testing.T) {
	c := estimateCost(60*time.Second, domain.AgentTypeClaudeCode)
	wantInput := int64(60 * inputTokensPerSec)
	wantOutput := int64(60 * outputTokensPerSec)
	if c.InputTokens != wantInput {
		t.Errorf("input tokens: want %d, got %d", wantInput, c.InputTokens)
	}
	if c.OutputTokens != wantOutput {
		t.Errorf("output tokens: want %d, got %d", wantOutput, c.OutputTokens)
	}
	if c.TotalTokens != wantInput+wantOutput {
		t.Errorf("total tokens: want %d, got %d", wantInput+wantOutput, c.TotalTokens)
	}
	if c.DurationSecs != 60 {
		t.Errorf("duration_secs: want 60, got %v", c.DurationSecs)
	}
}

func TestEstimateCost_CostCalculation(t *testing.T) {
	secsForMInputTokens := float64(1_000_000) / inputTokensPerSec
	c := estimateCost(time.Duration(secsForMInputTokens)*time.Second, domain.AgentTypeClaudeCode)
	wantInputCost := 3.00
	gotInputCost := float64(c.InputTokens) / 1_000_000 * pricePerMInputUSD
	if diff := gotInputCost - wantInputCost; diff > 0.01 || diff < -0.01 {
		t.Errorf("input cost: want ~$%.2f, got $%.4f", wantInputCost, gotInputCost)
	}
}

func TestEstimateCost_NegativeDurationTreatedAsZero(t *testing.T) {
	c := estimateCost(-5*time.Second, domain.AgentTypeClaudeCode)
	if c.InputTokens != 0 || c.OutputTokens != 0 {
		t.Errorf("negative duration should yield zero tokens, got %+v", c)
	}
}

func TestEstimateCost_Gemini_ZeroCost(t *testing.T) {
	c := estimateCost(10*time.Minute, domain.AgentTypeGemini)
	if c.EstimatedUSD != 0 {
		t.Errorf("gemini free tier should have zero cost, got %v", c.EstimatedUSD)
	}
	if c.ModelUsed != geminiModel {
		t.Errorf("expected model %q, got %q", geminiModel, c.ModelUsed)
	}
	if c.AgentType != domain.AgentTypeGemini {
		t.Errorf("expected agent type %q, got %q", domain.AgentTypeGemini, c.AgentType)
	}
	if c.DurationSecs != 600 {
		t.Errorf("duration_secs: want 600, got %v", c.DurationSecs)
	}
}
