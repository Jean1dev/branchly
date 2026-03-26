package executor

import (
	"testing"
	"time"
)

func TestEstimateCost_ZeroDuration(t *testing.T) {
	c := estimateCost(0)
	if c.InputTokens != 0 || c.OutputTokens != 0 || c.TotalTokens != 0 {
		t.Errorf("expected zero tokens for zero duration, got %+v", c)
	}
	if c.EstimatedUSD != 0 {
		t.Errorf("expected zero cost for zero duration, got %v", c.EstimatedUSD)
	}
	if c.ModelUsed != modelUsed {
		t.Errorf("expected model %q, got %q", modelUsed, c.ModelUsed)
	}
}

func TestEstimateCost_OneMinute(t *testing.T) {
	c := estimateCost(60 * time.Second)
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
	// 1 000 000 input tokens => $3.00, 1 000 000 output tokens => $15.00
	secsForMInputTokens := float64(1_000_000) / inputTokensPerSec
	c := estimateCost(time.Duration(secsForMInputTokens) * time.Second)
	// At exactly 1M input tokens the input cost should be $3.00.
	wantInputCost := 3.00
	gotInputCost := float64(c.InputTokens) / 1_000_000 * pricePerMInputUSD
	if diff := gotInputCost - wantInputCost; diff > 0.01 || diff < -0.01 {
		t.Errorf("input cost: want ~$%.2f, got $%.4f", wantInputCost, gotInputCost)
	}
}

func TestEstimateCost_NegativeDurationTreatedAsZero(t *testing.T) {
	c := estimateCost(-5 * time.Second)
	if c.InputTokens != 0 || c.OutputTokens != 0 {
		t.Errorf("negative duration should yield zero tokens, got %+v", c)
	}
}

func TestEstimateCost_TenMinutes_ModelName(t *testing.T) {
	c := estimateCost(10 * time.Minute)
	if c.ModelUsed != "claude-sonnet-4-6" {
		t.Errorf("wrong model: %q", c.ModelUsed)
	}
	if c.EstimatedUSD <= 0 {
		t.Errorf("expected positive cost for 10-minute job, got %v", c.EstimatedUSD)
	}
}
