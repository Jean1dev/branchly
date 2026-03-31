package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/infra"
	"github.com/branchly/branchly-api/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrInvalidToken           = errors.New("invalid or expired token")
	ErrAlreadyConnected       = errors.New("provider already connected")
	ErrIntegrationInUse       = errors.New("integration has connected repositories")
	ErrCannotDisconnectGitHub = errors.New("github integration cannot be disconnected")
)

type IntegrationService struct {
	cfg          *config.Config
	integrations repository.IntegrationRepository
	repos        domain.ConnectedRepositoryRepository
	httpClient   *http.Client
	gitlabBase   string
	azureBase    string
}

func NewIntegrationService(
	cfg *config.Config,
	integrations repository.IntegrationRepository,
	repos domain.ConnectedRepositoryRepository,
	httpClient *http.Client,
) *IntegrationService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 8 * time.Second}
	}
	return &IntegrationService{
		cfg:          cfg,
		integrations: integrations,
		repos:        repos,
		httpClient:   httpClient,
		gitlabBase:   "https://gitlab.com",
		azureBase:    "https://dev.azure.com",
	}
}

// ConnectGitHub creates or updates a GitHub OAuth integration for the user.
// Called automatically during the auth flow on every sign-in to refresh the token.
func (s *IntegrationService) ConnectGitHub(ctx context.Context, userID, oauthToken string) (*domain.GitIntegration, error) {
	enc, err := infra.Encrypt(oauthToken, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("integration service: encrypt github token: %w", err)
	}
	ig := &domain.GitIntegration{
		ID:             uuid.New().String(),
		UserID:         userID,
		Provider:       domain.GitProviderGitHub,
		EncryptedToken: enc,
		TokenType:      domain.TokenTypeOAuth,
		Scopes:         []string{"repo"},
		ConnectedAt:    time.Now().UTC(),
	}
	if err := s.integrations.Upsert(ctx, ig); err != nil {
		return nil, fmt.Errorf("integration service: connect github upsert: %w", err)
	}
	out, err := s.integrations.FindByUserAndProvider(ctx, userID, domain.GitProviderGitHub)
	if err != nil {
		return nil, fmt.Errorf("integration service: connect github find: %w", err)
	}
	return out, nil
}

// ConnectGitLab validates a PAT against the GitLab API and saves the integration.
func (s *IntegrationService) ConnectGitLab(ctx context.Context, userID, pat string) (*domain.GitIntegration, error) {
	existing, err := s.integrations.FindByUserAndProvider(ctx, userID, domain.GitProviderGitLab)
	if err != nil {
		return nil, fmt.Errorf("integration service: connect gitlab check: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyConnected
	}

	if err := s.validateGitLabPAT(ctx, pat); err != nil {
		return nil, err
	}

	enc, err := infra.Encrypt(pat, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("integration service: encrypt gitlab token: %w", err)
	}
	ig := &domain.GitIntegration{
		ID:             uuid.New().String(),
		UserID:         userID,
		Provider:       domain.GitProviderGitLab,
		EncryptedToken: enc,
		TokenType:      domain.TokenTypePAT,
		Scopes:         []string{"read_user", "read_api", "read_repository", "write_repository"},
		ConnectedAt:    time.Now().UTC(),
	}
	if err := s.integrations.Upsert(ctx, ig); err != nil {
		return nil, fmt.Errorf("integration service: connect gitlab upsert: %w", err)
	}
	out, err := s.integrations.FindByUserAndProvider(ctx, userID, domain.GitProviderGitLab)
	if err != nil {
		return nil, fmt.Errorf("integration service: connect gitlab find: %w", err)
	}
	return out, nil
}

func (s *IntegrationService) validateGitLabPAT(ctx context.Context, pat string) error {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, s.gitlabBase+"/api/v4/user", nil)
	if err != nil {
		return fmt.Errorf("integration service: gitlab validate request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", pat)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("integration service: gitlab validate call: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrInvalidToken
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("integration service: gitlab validate status %d", resp.StatusCode)
	}
	return nil
}

// ListByUserID returns all integrations for a user, GitHub always first.
func (s *IntegrationService) ListByUserID(ctx context.Context, userID string) ([]*domain.GitIntegration, error) {
	list, err := s.integrations.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("integration service: list: %w", err)
	}
	slices.SortFunc(list, func(a, b *domain.GitIntegration) int {
		if a.Provider == domain.GitProviderGitHub {
			return -1
		}
		if b.Provider == domain.GitProviderGitHub {
			return 1
		}
		return strings.Compare(string(a.Provider), string(b.Provider))
	})
	return list, nil
}

// Disconnect removes an integration if it has no connected repositories.
// GitHub integrations cannot be removed.
func (s *IntegrationService) Disconnect(ctx context.Context, userID, integrationID string) error {
	ig, err := s.integrations.FindByID(ctx, integrationID)
	if err != nil {
		return fmt.Errorf("integration service: disconnect find: %w", err)
	}
	if ig == nil || ig.UserID != userID {
		return ErrNotFound
	}
	if ig.Provider == domain.GitProviderGitHub {
		return ErrCannotDisconnectGitHub
	}

	connectedRepos, err := s.repos.FindByIntegrationID(ctx, integrationID)
	if err != nil {
		return fmt.Errorf("integration service: disconnect check repos: %w", err)
	}
	if len(connectedRepos) > 0 {
		return ErrIntegrationInUse
	}

	if err := s.integrations.Delete(ctx, integrationID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return fmt.Errorf("integration service: disconnect delete: %w", err)
	}
	return nil
}

// FindForUser returns an integration by ID, validating ownership.
func (s *IntegrationService) FindForUser(ctx context.Context, userID, integrationID string) (*domain.GitIntegration, error) {
	ig, err := s.integrations.FindByID(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("integration service: find for user: %w", err)
	}
	if ig == nil || ig.UserID != userID {
		return nil, nil
	}
	return ig, nil
}

// listGitHubRepos fetches repos from the GitHub API using the decrypted token.
func (s *IntegrationService) listGitHubRepos(ctx context.Context, token string) ([]ProviderRepo, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://api.github.com/user/repos?per_page=100&sort=updated", nil)
	if err != nil {
		return nil, fmt.Errorf("integration service: github repos request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("integration service: github repos call: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("integration service: github repos status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("integration service: github repos read: %w", err)
	}
	type ghRepo struct {
		ID            int64  `json:"id"`
		FullName      string `json:"full_name"`
		DefaultBranch string `json:"default_branch"`
		Language      string `json:"language"`
		Permissions   *struct {
			Push bool `json:"push"`
		} `json:"permissions"`
	}
	var raw []ghRepo
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("integration service: github repos decode: %w", err)
	}
	out := make([]ProviderRepo, 0, len(raw))
	for _, r := range raw {
		if r.Permissions != nil && !r.Permissions.Push {
			continue
		}
		out = append(out, ProviderRepo{
			ExternalID:    fmt.Sprintf("%d", r.ID),
			FullName:      r.FullName,
			CloneURL:      "https://github.com/" + r.FullName + ".git",
			DefaultBranch: r.DefaultBranch,
			Language:      r.Language,
			Provider:      domain.GitProviderGitHub,
		})
	}
	return out, nil
}

// ConnectAzureDevOps validates a PAT against the Azure DevOps API and saves the integration.
func (s *IntegrationService) ConnectAzureDevOps(ctx context.Context, userID, pat, orgURL string) (*domain.GitIntegration, error) {
	existing, err := s.integrations.FindByUserAndProvider(ctx, userID, domain.GitProviderAzureDevOps)
	if err != nil {
		return nil, fmt.Errorf("integration service: connect azure-devops check: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyConnected
	}

	orgURL = strings.TrimRight(orgURL, "/")
	if err := s.validateAzureDevOpsPAT(ctx, orgURL, pat); err != nil {
		return nil, err
	}

	enc, err := infra.Encrypt(pat, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("integration service: encrypt azure-devops token: %w", err)
	}
	ig := &domain.GitIntegration{
		ID:             uuid.New().String(),
		UserID:         userID,
		Provider:       domain.GitProviderAzureDevOps,
		EncryptedToken: enc,
		TokenType:      domain.TokenTypePAT,
		Scopes:         []string{"vso.code_write"},
		OrgURL:         orgURL,
		ConnectedAt:    time.Now().UTC(),
	}
	if err := s.integrations.Upsert(ctx, ig); err != nil {
		return nil, fmt.Errorf("integration service: connect azure-devops upsert: %w", err)
	}
	out, err := s.integrations.FindByUserAndProvider(ctx, userID, domain.GitProviderAzureDevOps)
	if err != nil {
		return nil, fmt.Errorf("integration service: connect azure-devops find: %w", err)
	}
	return out, nil
}

func (s *IntegrationService) azureBasicAuth(pat string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+pat))
}

func (s *IntegrationService) validateAzureDevOpsPAT(ctx context.Context, orgURL, pat string) error {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	url := orgURL + "/_apis/projects?api-version=7.0&$top=1"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("integration service: azure-devops validate request: %w", err)
	}
	req.Header.Set("Authorization", s.azureBasicAuth(pat))
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("integration service: azure-devops validate call: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrInvalidToken
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("integration service: azure-devops validate status %d", resp.StatusCode)
	}
	return nil
}

// listAzureDevOpsRepos fetches repositories from Azure DevOps using the decrypted PAT.
func (s *IntegrationService) listAzureDevOpsRepos(ctx context.Context, orgURL, pat string) ([]ProviderRepo, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	url := orgURL + "/_apis/git/repositories?api-version=7.0"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("integration service: azure-devops repos request: %w", err)
	}
	req.Header.Set("Authorization", s.azureBasicAuth(pat))
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("integration service: azure-devops repos call: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("integration service: azure-devops repos status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("integration service: azure-devops repos read: %w", err)
	}
	type azProject struct {
		Name string `json:"name"`
	}
	type azRepo struct {
		ID            string    `json:"id"`
		Name          string    `json:"name"`
		DefaultBranch string    `json:"defaultBranch"`
		RemoteURL     string    `json:"remoteUrl"`
		Project       azProject `json:"project"`
	}
	var raw struct {
		Value []azRepo `json:"value"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("integration service: azure-devops repos decode: %w", err)
	}

	// Extract org name from orgURL for FullName construction.
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(orgURL, "https://"), "http://"), "/")
	org := ""
	if len(parts) >= 2 {
		org = parts[1] // dev.azure.com/{org}
	} else if len(parts) == 1 {
		org = parts[0] // legacy visualstudio.com URL: already org name
	}

	out := make([]ProviderRepo, 0, len(raw.Value))
	for _, r := range raw.Value {
		defaultBranch := strings.TrimPrefix(r.DefaultBranch, "refs/heads/")
		if defaultBranch == "" {
			defaultBranch = "main"
		}
		fullName := org + "/" + r.Project.Name + "/" + r.Name
		cloneURL := orgURL + "/" + r.Project.Name + "/_git/" + r.Name
		out = append(out, ProviderRepo{
			ExternalID:    r.ID,
			FullName:      fullName,
			CloneURL:      cloneURL,
			DefaultBranch: defaultBranch,
			Provider:      domain.GitProviderAzureDevOps,
		})
	}
	return out, nil
}

// listGitLabProjects fetches projects from the GitLab API using the decrypted PAT.
func (s *IntegrationService) listGitLabProjects(ctx context.Context, token string) ([]ProviderRepo, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	url := s.gitlabBase + "/api/v4/projects?membership=true&min_access_level=30&per_page=100"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("integration service: gitlab projects request: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("integration service: gitlab projects call: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("integration service: gitlab projects status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("integration service: gitlab projects read: %w", err)
	}
	type glProject struct {
		ID                int64  `json:"id"`
		PathWithNamespace string `json:"path_with_namespace"`
		DefaultBranch     string `json:"default_branch"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		Language          string `json:"language"`
	}
	var raw []glProject
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("integration service: gitlab projects decode: %w", err)
	}
	out := make([]ProviderRepo, 0, len(raw))
	for _, p := range raw {
		out = append(out, ProviderRepo{
			ExternalID:    fmt.Sprintf("%d", p.ID),
			FullName:      p.PathWithNamespace,
			CloneURL:      p.HTTPURLToRepo,
			DefaultBranch: p.DefaultBranch,
			Language:      p.Language,
			Provider:      domain.GitProviderGitLab,
		})
	}
	return out, nil
}
