package infra_test

import (
	"context"
	"errors"
	"testing"

	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/infra"
)

// encryptionKey is a 32-byte test key.
var testEncKey = []byte("00000000000000000000000000000001")

// stubAPIKeyRepo is an in-memory fake of domain.APIKeyRepository.
type stubAPIKeyRepo struct {
	key *domain.UserAPIKey
	err error
}

func (s *stubAPIKeyRepo) FindByUserAndProvider(_ context.Context, _ string, _ domain.APIKeyProvider) (*domain.UserAPIKey, error) {
	return s.key, s.err
}

func encryptedKey(t *testing.T, plain string) string {
	t.Helper()
	ct, err := infra.Encrypt(plain, testEncKey)
	if err != nil {
		t.Fatalf("encrypt test key: %v", err)
	}
	return ct
}

func TestKeyResolver_UserKeyTakesPriority(t *testing.T) {
	userKey := &domain.UserAPIKey{
		UserID:       "user1",
		Provider:     domain.APIKeyProviderAnthropic,
		EncryptedKey: encryptedKey(t, "sk-ant-user-secret"),
	}
	repo := &stubAPIKeyRepo{key: userKey}
	resolver := infra.NewKeyResolver(repo, testEncKey, map[domain.APIKeyProvider]string{
		domain.APIKeyProviderAnthropic: "sk-ant-global-fallback",
	})

	got, err := resolver.Resolve(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "sk-ant-user-secret" {
		t.Errorf("expected user key, got %q", got)
	}
}

func TestKeyResolver_FallsBackToGlobalKeyWhenUserHasNone(t *testing.T) {
	repo := &stubAPIKeyRepo{key: nil} // no user key
	resolver := infra.NewKeyResolver(repo, testEncKey, map[domain.APIKeyProvider]string{
		domain.APIKeyProviderAnthropic: "sk-ant-global-key",
	})

	got, err := resolver.Resolve(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "sk-ant-global-key" {
		t.Errorf("expected global key, got %q", got)
	}
}

func TestKeyResolver_ErrorWhenNoKeyAvailable(t *testing.T) {
	repo := &stubAPIKeyRepo{key: nil}
	resolver := infra.NewKeyResolver(repo, testEncKey, map[domain.APIKeyProvider]string{}) // no globals

	_, err := resolver.Resolve(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if err == nil {
		t.Fatal("expected error when no key available, got nil")
	}
}

func TestKeyResolver_CorruptedUserKeyReturnsError(t *testing.T) {
	// Corrupted key — must NOT silently fall back to global.
	userKey := &domain.UserAPIKey{
		UserID:       "user1",
		Provider:     domain.APIKeyProviderAnthropic,
		EncryptedKey: "not-valid-base64!!!",
	}
	repo := &stubAPIKeyRepo{key: userKey}
	resolver := infra.NewKeyResolver(repo, testEncKey, map[domain.APIKeyProvider]string{
		domain.APIKeyProviderAnthropic: "sk-ant-global-fallback",
	})

	_, err := resolver.Resolve(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if err == nil {
		t.Fatal("expected error for corrupted key, got nil")
	}
}

func TestKeyResolver_RepoLookupErrorPropagates(t *testing.T) {
	repo := &stubAPIKeyRepo{key: nil, err: errors.New("db timeout")}
	resolver := infra.NewKeyResolver(repo, testEncKey, map[domain.APIKeyProvider]string{
		domain.APIKeyProviderAnthropic: "sk-ant-global-fallback",
	})

	_, err := resolver.Resolve(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if err == nil {
		t.Fatal("expected error when repo lookup fails, got nil")
	}
}
