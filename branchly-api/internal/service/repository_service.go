package service

import (
	"context"
	"errors"
	"fmt"
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
	ErrNotFound           = errors.New("not found")
	ErrAlreadyConnectedRepo = errors.New("repository already connected")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrRateLimitExceeded  = errors.New("active jobs limit reached")
)

// ProviderRepo is a normalised repository from any Git provider.
type ProviderRepo struct {
	ExternalID    string
	FullName      string
	CloneURL      string
	DefaultBranch string
	Language      string
	Provider      domain.GitProvider
}

type RepositoryService struct {
	cfg          *config.Config
	integrations repository.IntegrationRepository
	repos        domain.ConnectedRepositoryRepository
	integSvc     *IntegrationService
}

func NewRepositoryService(
	cfg *config.Config,
	integrations repository.IntegrationRepository,
	repos domain.ConnectedRepositoryRepository,
	integSvc *IntegrationService,
) *RepositoryService {
	return &RepositoryService{cfg: cfg, integrations: integrations, repos: repos, integSvc: integSvc}
}

func (s *RepositoryService) ListConnected(ctx context.Context, userID string) ([]*domain.Repository, error) {
	list, err := s.repos.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository service: list: %w", err)
	}
	return list, nil
}

// ConnectRepositoryInput carries the data needed to connect a repository.
type ConnectRepositoryInput struct {
	IntegrationID string
	ExternalID    string
	FullName      string
	CloneURL      string
	DefaultBranch string
	Language      string
	Provider      domain.GitProvider
}

func (s *RepositoryService) Connect(ctx context.Context, userID string, in ConnectRepositoryInput) (*domain.Repository, error) {
	ig, err := s.integrations.FindByID(ctx, in.IntegrationID)
	if err != nil {
		return nil, fmt.Errorf("repository service: connect find integration: %w", err)
	}
	if ig == nil || ig.UserID != userID {
		return nil, ErrNotFound
	}

	defaultBranch := strings.TrimSpace(in.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	existing, err := s.repos.FindByUserExternalAndProvider(ctx, userID, in.ExternalID, in.Provider)
	if err != nil {
		return nil, fmt.Errorf("repository service: connect find existing: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyConnectedRepo
	}

	r := &domain.Repository{
		ID:            uuid.New().String(),
		UserID:        userID,
		IntegrationID: in.IntegrationID,
		Provider:      in.Provider,
		ExternalID:    in.ExternalID,
		FullName:      in.FullName,
		CloneURL:      in.CloneURL,
		DefaultBranch: defaultBranch,
		Language:      in.Language,
		ConnectedAt:   time.Now().UTC(),
	}
	if err := s.repos.Create(ctx, r); err != nil {
		return nil, fmt.Errorf("repository service: create: %w", err)
	}
	return r, nil
}

func (s *RepositoryService) Disconnect(ctx context.Context, userID, repoID string) error {
	r, err := s.repos.FindByID(ctx, repoID)
	if err != nil {
		return fmt.Errorf("repository service: disconnect find: %w", err)
	}
	if r == nil || r.UserID != userID {
		return ErrNotFound
	}
	if err := s.repos.Delete(ctx, repoID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return fmt.Errorf("repository service: delete: %w", err)
	}
	return nil
}

// ListFromProvider lists repositories available to connect from a given integration.
// Repos already connected via this integration are filtered out.
func (s *RepositoryService) ListFromProvider(ctx context.Context, userID, integrationID string) ([]ProviderRepo, error) {
	ig, err := s.integrations.FindByID(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("repository service: list from provider find integration: %w", err)
	}
	if ig == nil || ig.UserID != userID {
		return nil, ErrNotFound
	}

	token, err := infra.Decrypt(ig.EncryptedToken, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("repository service: decrypt token: %w", err)
	}

	var all []ProviderRepo
	switch ig.Provider {
	case domain.GitProviderGitHub:
		all, err = s.integSvc.listGitHubRepos(ctx, token)
	case domain.GitProviderGitLab:
		all, err = s.integSvc.listGitLabProjects(ctx, token)
	default:
		return nil, fmt.Errorf("repository service: unsupported provider %s", ig.Provider)
	}
	if err != nil {
		return nil, fmt.Errorf("repository service: list from provider fetch: %w", err)
	}

	// Filter already-connected repos for this integration.
	connected, err := s.repos.FindByIntegrationID(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("repository service: list connected: %w", err)
	}
	alreadyConnected := make(map[string]struct{}, len(connected))
	for _, r := range connected {
		alreadyConnected[r.ExternalID] = struct{}{}
	}

	filtered := make([]ProviderRepo, 0, len(all))
	for _, pr := range all {
		if _, ok := alreadyConnected[pr.ExternalID]; !ok {
			filtered = append(filtered, pr)
		}
	}

	slices.SortFunc(filtered, func(a, b ProviderRepo) int {
		return strings.Compare(strings.ToLower(a.FullName), strings.ToLower(b.FullName))
	})
	return filtered, nil
}

func (s *RepositoryService) GetOwned(ctx context.Context, userID, repoID string) (*domain.Repository, error) {
	r, err := s.repos.FindByID(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("repository service: get: %w", err)
	}
	if r == nil || r.UserID != userID {
		return nil, nil
	}
	return r, nil
}
