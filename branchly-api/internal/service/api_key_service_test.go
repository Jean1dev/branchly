package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
)

// ----- in-memory fake repository -----

type fakeAPIKeyRepo struct {
	keys map[string]*domain.UserAPIKey // key: userID+":"+provider
	err  error
}

func newFakeAPIKeyRepo() *fakeAPIKeyRepo {
	return &fakeAPIKeyRepo{keys: make(map[string]*domain.UserAPIKey)}
}

func repoKey(userID string, provider domain.APIKeyProvider) string {
	return userID + ":" + string(provider)
}

func (r *fakeAPIKeyRepo) Upsert(_ context.Context, key *domain.UserAPIKey) error {
	if r.err != nil {
		return r.err
	}
	r.keys[repoKey(key.UserID, key.Provider)] = key
	return nil
}

func (r *fakeAPIKeyRepo) FindByUserAndProvider(_ context.Context, userID string, provider domain.APIKeyProvider) (*domain.UserAPIKey, error) {
	if r.err != nil {
		return nil, r.err
	}
	k, ok := r.keys[repoKey(userID, provider)]
	if !ok {
		return nil, nil
	}
	return k, nil
}

func (r *fakeAPIKeyRepo) FindAllByUserID(_ context.Context, userID string) ([]*domain.UserAPIKey, error) {
	if r.err != nil {
		return nil, r.err
	}
	var out []*domain.UserAPIKey
	for _, k := range r.keys {
		if k.UserID == userID {
			cp := *k
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeAPIKeyRepo) Delete(_ context.Context, userID string, provider domain.APIKeyProvider) error {
	if r.err != nil {
		return r.err
	}
	k := repoKey(userID, provider)
	if _, ok := r.keys[k]; !ok {
		return errors.New("not found")
	}
	delete(r.keys, k)
	return nil
}

// testEncKey is a 32-byte key for tests.
var testEncKey = []byte("00000000000000000000000000000001")

func newAPIKeyTestService() (*APIKeyService, *fakeAPIKeyRepo) {
	repo := newFakeAPIKeyRepo()
	svc := NewAPIKeyService(repo, testEncKey)
	return svc, repo
}

// ----- tests -----

func TestSave_AnthropicKey_SavesWithCorrectHint(t *testing.T) {
	svc, repo := newAPIKeyTestService()
	const plainKey = "sk-ant-api01-abcdefghijklmnopqrst"
	err := svc.Save(context.Background(), "user1", domain.APIKeyProviderAnthropic, plainKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stored := repo.keys[repoKey("user1", domain.APIKeyProviderAnthropic)]
	if stored == nil {
		t.Fatal("key not stored")
	}
	wantHint := plainKey[len(plainKey)-4:]
	if stored.KeyHint != wantHint {
		t.Errorf("expected hint %q, got %q", wantHint, stored.KeyHint)
	}
	if stored.EncryptedKey == "" {
		t.Error("encrypted key must not be empty")
	}
	if stored.EncryptedKey == plainKey {
		t.Error("encrypted key must differ from plaintext")
	}
}

func TestSave_InvalidPrefix_ReturnsErrInvalidKeyFormat(t *testing.T) {
	svc, _ := newAPIKeyTestService()
	err := svc.Save(context.Background(), "user1", domain.APIKeyProviderAnthropic, "wrong-prefix-key")
	if !errors.Is(err, ErrInvalidKeyFormat) {
		t.Errorf("expected ErrInvalidKeyFormat, got %v", err)
	}
}

func TestSave_GoogleKeyWrongPrefix_ReturnsErrInvalidKeyFormat(t *testing.T) {
	svc, _ := newAPIKeyTestService()
	err := svc.Save(context.Background(), "user1", domain.APIKeyProviderGoogle, "sk-wrong")
	if !errors.Is(err, ErrInvalidKeyFormat) {
		t.Errorf("expected ErrInvalidKeyFormat, got %v", err)
	}
}

func TestSave_SameProviderTwice_Upserts(t *testing.T) {
	svc, repo := newAPIKeyTestService()
	_ = svc.Save(context.Background(), "user1", domain.APIKeyProviderAnthropic, "sk-ant-first-keyABCD")
	firstHint := repo.keys[repoKey("user1", domain.APIKeyProviderAnthropic)].KeyHint

	_ = svc.Save(context.Background(), "user1", domain.APIKeyProviderAnthropic, "sk-ant-second-keyXYZW")
	secondHint := repo.keys[repoKey("user1", domain.APIKeyProviderAnthropic)].KeyHint

	if firstHint == secondHint {
		t.Error("expected hint to change after update")
	}
	if secondHint != "XYZW" {
		t.Errorf("expected hint XYZW, got %q", secondHint)
	}
	// Only one key should be stored.
	count := 0
	for _, k := range repo.keys {
		if k.UserID == "user1" && k.Provider == domain.APIKeyProviderAnthropic {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 key, found %d", count)
	}
}

func TestDelete_ExistingKey_Succeeds(t *testing.T) {
	svc, repo := newAPIKeyTestService()
	_ = svc.Save(context.Background(), "user1", domain.APIKeyProviderAnthropic, "sk-ant-api01-test1234")
	err := svc.Delete(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.keys[repoKey("user1", domain.APIKeyProviderAnthropic)]; ok {
		t.Error("key should have been deleted")
	}
}

func TestDelete_NonExistentKey_ReturnsErrAPIKeyNotFound(t *testing.T) {
	svc, _ := newAPIKeyTestService()
	err := svc.Delete(context.Background(), "user1", domain.APIKeyProviderAnthropic)
	if !errors.Is(err, ErrAPIKeyNotFound) {
		t.Errorf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestList_NeverReturnsEncryptedKey(t *testing.T) {
	svc, _ := newAPIKeyTestService()
	_ = svc.Save(context.Background(), "user1", domain.APIKeyProviderAnthropic, "sk-ant-api01-test1234")

	infos, err := svc.ListByUserID(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(infos) == 0 {
		t.Fatal("expected at least one key info")
	}
	for _, info := range infos {
		// APIKeyInfo does not have an EncryptedKey field — this is enforced by the type.
		// The test verifies the returned type is safe.
		if info.Provider == "" {
			t.Error("provider should not be empty")
		}
		if info.KeyHint == "" {
			t.Error("key_hint should not be empty")
		}
		if info.UpdatedAt.IsZero() {
			t.Error("updated_at should not be zero")
		}
		_ = info.Provider
		_ = info.KeyHint
		_ = info.UpdatedAt
	}
}

func TestList_EmptyForUserWithNoKeys(t *testing.T) {
	svc, _ := newAPIKeyTestService()
	infos, err := svc.ListByUserID(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected empty list, got %d items", len(infos))
	}
}

func TestKeyHint_LastFourChars(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"sk-ant-api01-abcdefgh", "efgh"},
		{"AIzaSyABCD", "ABCD"},
		{"sk-1234", "1234"},
		{"abcd", "abcd"},
		{"abc", "abc"},
	}
	for _, tc := range cases {
		got := keyHint(tc.key)
		if got != tc.want {
			t.Errorf("keyHint(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func init() {
	// Silence unused import warnings for time
	_ = time.Now
}
