package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/infra"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyConnected = errors.New("repository already connected")
)

type RepositoryService struct {
	cfg  *config.Config
	users domain.UserRepository
	repos domain.ConnectedRepositoryRepository
}

func NewRepositoryService(cfg *config.Config, users domain.UserRepository, repos domain.ConnectedRepositoryRepository) *RepositoryService {
	return &RepositoryService{cfg: cfg, users: users, repos: repos}
}

func (s *RepositoryService) githubHTTPClient(ctx context.Context) *http.Client {
	return &http.Client{Timeout: 8 * time.Second}
}

func (s *RepositoryService) decryptUserToken(u *domain.User) (string, error) {
	tok, err := infra.Decrypt(u.EncryptedToken, s.cfg.EncryptionKey)
	if err != nil {
		return "", fmt.Errorf("repository service: decrypt token: %w", err)
	}
	return tok, nil
}

func (s *RepositoryService) ListConnected(ctx context.Context, userID string) ([]*domain.Repository, error) {
	list, err := s.repos.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository service: list: %w", err)
	}
	return list, nil
}

type ConnectRepositoryInput struct {
	GithubRepoID  int64
	FullName      string
	DefaultBranch string
	Language      string
}

func (s *RepositoryService) Connect(ctx context.Context, userID string, in ConnectRepositoryInput) (*domain.Repository, error) {
	existing, err := s.repos.FindByUserAndGithubRepoID(ctx, userID, in.GithubRepoID)
	if err != nil {
		return nil, fmt.Errorf("repository service: connect find: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyConnected
	}
	r := &domain.Repository{
		ID:            uuid.New().String(),
		UserID:        userID,
		GithubRepoID:  in.GithubRepoID,
		FullName:      in.FullName,
		DefaultBranch: in.DefaultBranch,
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

type GitHubRepoListItem struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	Language      string `json:"language"`
	Private       bool   `json:"private"`
}

func (s *RepositoryService) ListGitHubAvailable(ctx context.Context, userID string) ([]GitHubRepoListItem, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("repository service: user: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	token, err := s.decryptUserToken(u)
	if err != nil {
		return nil, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://api.github.com/user/repos?per_page=100&sort=updated", nil)
	if err != nil {
		return nil, fmt.Errorf("repository service: request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := s.githubHTTPClient(ctx).Do(req)
	if err != nil {
		return nil, fmt.Errorf("repository service: github repos: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repository service: github repos status %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("repository service: read body: %w", err)
	}
	var out []GitHubRepoListItem
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("repository service: decode repos: %w", err)
	}
	return out, nil
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
