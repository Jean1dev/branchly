package gitprovider_test

import (
	"testing"

	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/gitprovider"
)

func TestFactory_Create_GitHub_ReturnsClient(t *testing.T) {
	f := gitprovider.NewFactory()
	client, err := f.Create(domain.GitProviderGitHub)
	if err != nil {
		t.Fatalf("expected no error for github, got %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client for github")
	}
}

func TestFactory_Create_GitLab_ReturnsClient(t *testing.T) {
	f := gitprovider.NewFactory()
	client, err := f.Create(domain.GitProviderGitLab)
	if err != nil {
		t.Fatalf("expected no error for gitlab, got %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client for gitlab")
	}
}

func TestFactory_Create_UnknownProvider_ReturnsError(t *testing.T) {
	f := gitprovider.NewFactory()
	client, err := f.Create("bitbucket")
	if err == nil {
		t.Error("expected error for unsupported provider, got nil")
	}
	if client != nil {
		t.Error("expected nil client for unsupported provider")
	}
}

func TestFactory_Create_EmptyProvider_ReturnsError(t *testing.T) {
	f := gitprovider.NewFactory()
	_, err := f.Create("")
	if err == nil {
		t.Error("expected error for empty provider string, got nil")
	}
}

func TestFactory_Create_GitHubAndGitLab_ReturnDistinctClients(t *testing.T) {
	f := gitprovider.NewFactory()
	ghClient, err := f.Create(domain.GitProviderGitHub)
	if err != nil {
		t.Fatalf("github: %v", err)
	}
	glClient, err := f.Create(domain.GitProviderGitLab)
	if err != nil {
		t.Fatalf("gitlab: %v", err)
	}
	// Both should be non-nil and distinct instances.
	if ghClient == nil || glClient == nil {
		t.Error("both clients must be non-nil")
	}
	if ghClient == glClient {
		t.Error("github and gitlab clients should be distinct instances")
	}
}
