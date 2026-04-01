package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/repository"
)

// ---- in-memory mocks ----

type mockIntegrationRepo struct {
	byProvider map[string]*domain.GitIntegration // key: "<userID>:<provider>"
	byID       map[string]*domain.GitIntegration
	upsertErr  error
	deleteErr  error
	upserted   []*domain.GitIntegration
	deleted    []string
}

func newMockIntegrationRepo() *mockIntegrationRepo {
	return &mockIntegrationRepo{
		byProvider: make(map[string]*domain.GitIntegration),
		byID:       make(map[string]*domain.GitIntegration),
	}
}

func integKey(userID string, provider domain.GitProvider) string {
	return userID + ":" + string(provider)
}

func (m *mockIntegrationRepo) Upsert(_ context.Context, ig *domain.GitIntegration) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	cp := *ig
	m.byProvider[integKey(ig.UserID, ig.Provider)] = &cp
	m.byID[ig.ID] = &cp
	m.upserted = append(m.upserted, &cp)
	return nil
}

func (m *mockIntegrationRepo) FindByUserAndProvider(_ context.Context, userID string, provider domain.GitProvider) (*domain.GitIntegration, error) {
	return m.byProvider[integKey(userID, provider)], nil
}

func (m *mockIntegrationRepo) FindByID(_ context.Context, id string) (*domain.GitIntegration, error) {
	return m.byID[id], nil
}

func (m *mockIntegrationRepo) FindAllByUserID(_ context.Context, userID string) ([]*domain.GitIntegration, error) {
	var list []*domain.GitIntegration
	for _, ig := range m.byID {
		if ig.UserID == userID {
			cp := *ig
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (m *mockIntegrationRepo) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	ig, ok := m.byID[id]
	if !ok {
		return errors.New("not found")
	}
	delete(m.byProvider, integKey(ig.UserID, ig.Provider))
	delete(m.byID, id)
	m.deleted = append(m.deleted, id)
	return nil
}

// stubConnectedRepoRepo stubs ConnectedRepositoryRepository for Disconnect tests.
type stubConnectedRepoRepo struct {
	byIntegration map[string][]*domain.Repository
}

func newStubRepoRepo() *stubConnectedRepoRepo {
	return &stubConnectedRepoRepo{byIntegration: make(map[string][]*domain.Repository)}
}

func (s *stubConnectedRepoRepo) Create(_ context.Context, _ *domain.Repository) error { return nil }
func (s *stubConnectedRepoRepo) FindByID(_ context.Context, _ string) (*domain.Repository, error) {
	return nil, nil
}
func (s *stubConnectedRepoRepo) FindByUserID(_ context.Context, _ string) ([]*domain.Repository, error) {
	return nil, nil
}
func (s *stubConnectedRepoRepo) Delete(_ context.Context, _ string) error { return nil }
func (s *stubConnectedRepoRepo) FindByUserExternalAndProvider(_ context.Context, _, _ string, _ domain.GitProvider) (*domain.Repository, error) {
	return nil, nil
}
func (s *stubConnectedRepoRepo) FindByIntegrationID(_ context.Context, id string) ([]*domain.Repository, error) {
	return s.byIntegration[id], nil
}

// ---- helper ----

// testEncryptionKey is a valid 32-byte key for AES-256-GCM.
var testEncryptionKey = []byte("12345678901234567890123456789012")

func newTestIntegSvc(
	integRepo repository.IntegrationRepository,
	repoRepo domain.ConnectedRepositoryRepository,
	httpClient *http.Client,
) *IntegrationService {
	cfg := &config.Config{EncryptionKey: testEncryptionKey}
	svc := NewIntegrationService(cfg, integRepo, repoRepo, httpClient)
	return svc
}

// ---- ConnectGitLab ----

func TestConnectGitLab_ValidPAT_Succeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") == "valid-pat" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	integRepo := newMockIntegrationRepo()
	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), server.Client())
	svc.gitlabBase = server.URL

	ig, err := svc.ConnectGitLab(context.Background(), "user-1", "valid-pat")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ig == nil {
		t.Fatal("expected integration, got nil")
	}
	if ig.Provider != domain.GitProviderGitLab {
		t.Errorf("expected provider %q, got %q", domain.GitProviderGitLab, ig.Provider)
	}
	if ig.TokenType != domain.TokenTypePAT {
		t.Errorf("expected token_type %q, got %q", domain.TokenTypePAT, ig.TokenType)
	}
	if ig.EncryptedToken == "" {
		t.Error("encrypted token should not be empty")
	}
	if ig.EncryptedToken == "valid-pat" {
		t.Error("token must be stored encrypted, not in plaintext")
	}
	if len(integRepo.upserted) != 1 {
		t.Errorf("expected 1 upsert call, got %d", len(integRepo.upserted))
	}
}

func TestConnectGitLab_InvalidPAT_ReturnsErrInvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	svc := newTestIntegSvc(newMockIntegrationRepo(), newStubRepoRepo(), server.Client())
	svc.gitlabBase = server.URL

	_, err := svc.ConnectGitLab(context.Background(), "user-1", "bad-pat")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestConnectGitLab_ForbiddenPAT_ReturnsErrInvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	svc := newTestIntegSvc(newMockIntegrationRepo(), newStubRepoRepo(), server.Client())
	svc.gitlabBase = server.URL

	_, err := svc.ConnectGitLab(context.Background(), "user-1", "bad-pat")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for 403, got %v", err)
	}
}

func TestConnectGitLab_AlreadyConnected_ReturnsErrAlreadyConnected(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	// Pre-seed an existing GitLab integration for user-1.
	existing := &domain.GitIntegration{
		ID:       "existing-id",
		UserID:   "user-1",
		Provider: domain.GitProviderGitLab,
	}
	integRepo.byProvider[integKey("user-1", domain.GitProviderGitLab)] = existing
	integRepo.byID["existing-id"] = existing

	// No HTTP server needed: AlreadyConnected is returned before any network call.
	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), nil)

	_, err := svc.ConnectGitLab(context.Background(), "user-1", "any-pat")
	if !errors.Is(err, ErrAlreadyConnected) {
		t.Errorf("expected ErrAlreadyConnected, got %v", err)
	}
	// No upsert should have happened.
	if len(integRepo.upserted) != 0 {
		t.Errorf("expected no upsert, got %d", len(integRepo.upserted))
	}
}

func TestConnectGitLab_DifferentUsers_CanBothConnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	integRepo := newMockIntegrationRepo()
	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), server.Client())
	svc.gitlabBase = server.URL

	_, err := svc.ConnectGitLab(context.Background(), "user-1", "pat-1")
	if err != nil {
		t.Fatalf("user-1: unexpected error: %v", err)
	}
	_, err = svc.ConnectGitLab(context.Background(), "user-2", "pat-2")
	if err != nil {
		t.Fatalf("user-2: unexpected error: %v", err)
	}
	if len(integRepo.upserted) != 2 {
		t.Errorf("expected 2 upserts (one per user), got %d", len(integRepo.upserted))
	}
}

// ---- Disconnect ----

func TestDisconnect_NoRepos_Succeeds(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	ig := &domain.GitIntegration{ID: "integ-1", UserID: "user-1", Provider: domain.GitProviderGitLab}
	integRepo.byID["integ-1"] = ig
	integRepo.byProvider[integKey("user-1", domain.GitProviderGitLab)] = ig

	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), nil)
	err := svc.Disconnect(context.Background(), "user-1", "integ-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(integRepo.deleted) != 1 || integRepo.deleted[0] != "integ-1" {
		t.Errorf("expected integration to be deleted, deleted: %v", integRepo.deleted)
	}
}

func TestDisconnect_WithRepos_ReturnsErrIntegrationInUse(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	ig := &domain.GitIntegration{ID: "integ-1", UserID: "user-1", Provider: domain.GitProviderGitLab}
	integRepo.byID["integ-1"] = ig

	repoRepo := newStubRepoRepo()
	repoRepo.byIntegration["integ-1"] = []*domain.Repository{{ID: "repo-1"}}

	svc := newTestIntegSvc(integRepo, repoRepo, nil)
	err := svc.Disconnect(context.Background(), "user-1", "integ-1")
	if !errors.Is(err, ErrIntegrationInUse) {
		t.Errorf("expected ErrIntegrationInUse, got %v", err)
	}
	if len(integRepo.deleted) > 0 {
		t.Error("integration must not be deleted when repos are connected")
	}
}

func TestDisconnect_WrongOwner_ReturnsErrNotFound(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	ig := &domain.GitIntegration{ID: "integ-1", UserID: "user-2", Provider: domain.GitProviderGitLab}
	integRepo.byID["integ-1"] = ig

	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), nil)
	err := svc.Disconnect(context.Background(), "user-1", "integ-1")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for wrong owner, got %v", err)
	}
}

func TestDisconnect_GitHub_ReturnsErrCannotDisconnectGitHub(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	ig := &domain.GitIntegration{ID: "integ-gh", UserID: "user-1", Provider: domain.GitProviderGitHub}
	integRepo.byID["integ-gh"] = ig

	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), nil)
	err := svc.Disconnect(context.Background(), "user-1", "integ-gh")
	if !errors.Is(err, ErrCannotDisconnectGitHub) {
		t.Errorf("expected ErrCannotDisconnectGitHub, got %v", err)
	}
	if len(integRepo.deleted) > 0 {
		t.Error("GitHub integration must not be deleted")
	}
}

func TestDisconnect_NonExistent_ReturnsErrNotFound(t *testing.T) {
	svc := newTestIntegSvc(newMockIntegrationRepo(), newStubRepoRepo(), nil)
	err := svc.Disconnect(context.Background(), "user-1", "nonexistent-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing integration, got %v", err)
	}
}

// ---- ListByUserID ----

func TestListByUserID_GitHubAlwaysFirst(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	// Add GitLab first, GitHub second — they should be reordered.
	gl := &domain.GitIntegration{ID: "gl-id", UserID: "user-1", Provider: domain.GitProviderGitLab}
	gh := &domain.GitIntegration{ID: "gh-id", UserID: "user-1", Provider: domain.GitProviderGitHub}
	integRepo.byID["gl-id"] = gl
	integRepo.byID["gh-id"] = gh

	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), nil)
	list, err := svc.ListByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 integrations, got %d", len(list))
	}
	if list[0].Provider != domain.GitProviderGitHub {
		t.Errorf("expected GitHub first, got %s", list[0].Provider)
	}
	if list[1].Provider != domain.GitProviderGitLab {
		t.Errorf("expected GitLab second, got %s", list[1].Provider)
	}
}

func TestListByUserID_OnlyOwnIntegrations(t *testing.T) {
	integRepo := newMockIntegrationRepo()
	user1gh := &domain.GitIntegration{ID: "u1-gh", UserID: "user-1", Provider: domain.GitProviderGitHub}
	user2gh := &domain.GitIntegration{ID: "u2-gh", UserID: "user-2", Provider: domain.GitProviderGitHub}
	integRepo.byID["u1-gh"] = user1gh
	integRepo.byID["u2-gh"] = user2gh

	svc := newTestIntegSvc(integRepo, newStubRepoRepo(), nil)
	list, err := svc.ListByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 integration for user-1, got %d", len(list))
	}
	if list[0].ID != "u1-gh" {
		t.Errorf("expected user-1's integration, got %s", list[0].ID)
	}
}

func TestListByUserID_Empty(t *testing.T) {
	svc := newTestIntegSvc(newMockIntegrationRepo(), newStubRepoRepo(), nil)
	list, err := svc.ListByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}
