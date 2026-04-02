package domain

import (
	"testing"
	"time"
)

func TestBackoffDuration(t *testing.T) {
	tests := []struct {
		attemptNumber int
		expected      time.Duration
	}{
		{1, 1 * time.Minute},
		{2, 5 * time.Minute},
		{3, 25 * time.Minute},
	}

	for _, tt := range tests {
		j := &Job{AttemptNumber: tt.attemptNumber}
		got := j.BackoffDuration()
		if got != tt.expected {
			t.Errorf("BackoffDuration() with attempt %d = %v, want %v",
				tt.attemptNumber, got, tt.expected)
		}
	}
}

func TestCanRetry(t *testing.T) {
	tests := []struct {
		name          string
		attemptNumber int
		maxAttempts   int
		failureType   FailureType
		expected      bool
	}{
		{
			name:          "transient failure with attempts remaining",
			attemptNumber: 1,
			maxAttempts:   3,
			failureType:   FailureTypeTransient,
			expected:      true,
		},
		{
			name:          "transient failure on last attempt",
			attemptNumber: 3,
			maxAttempts:   3,
			failureType:   FailureTypeTransient,
			expected:      false,
		},
		{
			name:          "permanent failure always false",
			attemptNumber: 1,
			maxAttempts:   3,
			failureType:   FailureTypePermanent,
			expected:      false,
		},
		{
			name:          "transient second attempt still retryable",
			attemptNumber: 2,
			maxAttempts:   3,
			failureType:   FailureTypeTransient,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Job{
				AttemptNumber: tt.attemptNumber,
				MaxAttempts:   tt.maxAttempts,
				FailureType:   tt.failureType,
			}
			got := j.CanRetry()
			if got != tt.expected {
				t.Errorf("CanRetry() = %v, want %v", got, tt.expected)
			}
		})
	}
}
