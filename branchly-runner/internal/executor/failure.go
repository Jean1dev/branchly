package executor

import (
	"github.com/branchly/branchly-runner/internal/classify"
	"github.com/branchly/branchly-runner/internal/domain"
)

// FailureClassifier wraps classify.FailureClassifier for use within the executor.
type FailureClassifier struct {
	inner classify.FailureClassifier
}

func (f *FailureClassifier) Classify(err error) domain.FailureType {
	return f.inner.Classify(err)
}
