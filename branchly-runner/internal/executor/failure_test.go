package executor

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/branchly/branchly-runner/internal/domain"
)

func TestFailureClassifier_Classify(t *testing.T) {
	c := &FailureClassifier{}

	tests := []struct {
		name     string
		err      error
		expected domain.FailureType
	}{
		{
			name:     "nil error returns empty",
			err:      nil,
			expected: "",
		},
		{
			name:     "connection timeout is transient",
			err:      errors.New("connection timeout after 30s"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "timeout keyword is transient",
			err:      errors.New("request timeout"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "connection refused is transient",
			err:      errors.New("connection refused by host"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "connection reset is transient",
			err:      errors.New("connection reset by peer"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "no such host is transient",
			err:      errors.New("dial: no such host"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "service unavailable is transient",
			err:      errors.New("service unavailable"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "rate limit is transient",
			err:      errors.New("rate limit exceeded"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "HTTP 429 is transient",
			err:      errors.New("status 429"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "HTTP 503 is transient",
			err:      errors.New("status 503"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "HTTP 502 is transient",
			err:      errors.New("status 502"),
			expected: domain.FailureTypeTransient,
		},
		{
			name:     "authentication failed is permanent",
			err:      errors.New("authentication failed: invalid credentials"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "invalid token is permanent",
			err:      errors.New("invalid token provided"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "repository not found is permanent",
			err:      errors.New("repository not found: lucasmendes/deleted-repo"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "ownership validation is permanent",
			err:      errors.New("ownership validation failed"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "permission denied is permanent",
			err:      errors.New("permission denied"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "agent exceeded maximum iterations is permanent",
			err:      errors.New("agent exceeded maximum iterations"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "unknown agent type is permanent",
			err:      errors.New("unknown agent type: invalid-agent"),
			expected: domain.FailureTypePermanent,
		},
		{
			name:     "unknown error defaults to permanent",
			err:      errors.New("some unexpected internal error"),
			expected: domain.FailureTypePermanent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Classify(tt.err)
			if got != tt.expected {
				t.Errorf("Classify(%v) = %q, want %q", tt.err, got, tt.expected)
			}
		})
	}
}

// mockTimeoutError implements net.Error with Timeout() == true.
type mockTimeoutError struct{}

func (e *mockTimeoutError) Error() string   { return "mock timeout error" }
func (e *mockTimeoutError) Timeout() bool   { return true }
func (e *mockTimeoutError) Temporary() bool { return true }

func TestFailureClassifier_NetTimeoutError(t *testing.T) {
	c := &FailureClassifier{}

	// A net.Error with Timeout() true should be transient even without keyword match.
	var netErr net.Error = &mockTimeoutError{}
	got := c.Classify(netErr)
	if got != domain.FailureTypeTransient {
		t.Errorf("expected transient for net.Error with Timeout()=true, got %q", got)
	}
}

func TestFailureClassifier_WrappedNetTimeoutError(t *testing.T) {
	c := &FailureClassifier{}
	wrapped := fmt.Errorf("outer: %w", &mockTimeoutError{})
	got := c.Classify(wrapped)
	if got != domain.FailureTypeTransient {
		t.Errorf("expected transient for wrapped net.Error, got %q", got)
	}
}
