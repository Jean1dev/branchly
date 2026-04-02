package executor

import (
	"errors"
	"net"
	"strings"

	"github.com/branchly/branchly-runner/internal/domain"
)

// FailureClassifier determines whether a job failure is transient (worth retrying)
// or permanent (should fail immediately without retry).
type FailureClassifier struct{}

// Classify returns the failure type for the given error.
// Unknown errors default to permanent to avoid infinite retries on unexpected conditions.
func (f *FailureClassifier) Classify(err error) domain.FailureType {
	if err == nil {
		return ""
	}

	msg := strings.ToLower(err.Error())

	permanent := []string{
		"authentication",
		"invalid token",
		"repository not found",
		"ownership validation",
		"permission denied",
		"agent exceeded maximum iterations",
		"unknown agent type",
	}
	for _, p := range permanent {
		if strings.Contains(msg, p) {
			return domain.FailureTypePermanent
		}
	}

	transient := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"no such host",
		"temporary failure",
		"service unavailable",
		"rate limit",
		"429",
		"503",
		"502",
	}
	for _, t := range transient {
		if strings.Contains(msg, t) {
			return domain.FailureTypeTransient
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return domain.FailureTypeTransient
	}

	return domain.FailureTypePermanent
}
