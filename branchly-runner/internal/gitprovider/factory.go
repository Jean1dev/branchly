package gitprovider

import (
	"fmt"

	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/gitprovider/github"
	"github.com/branchly/branchly-runner/internal/gitprovider/gitlab"
)

// Factory creates GitProviderClient instances based on provider type.
type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(provider domain.GitProvider) (domain.GitProviderClient, error) {
	switch provider {
	case domain.GitProviderGitHub:
		return github.NewClient(), nil
	case domain.GitProviderGitLab:
		return gitlab.NewClient(), nil
	default:
		return nil, fmt.Errorf("unsupported git provider: %s", provider)
	}
}
