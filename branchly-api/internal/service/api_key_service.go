package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/infra"
	"github.com/google/uuid"
)

var (
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrInvalidKeyFormat = errors.New("invalid api key format for provider")
)

// APIKeyInfo is the safe public view of a stored key — never includes the encrypted value.
type APIKeyInfo struct {
	Provider  domain.APIKeyProvider
	KeyHint   string
	UpdatedAt time.Time
}

type APIKeyService struct {
	repo   domain.APIKeyRepository
	encKey []byte
}

func NewAPIKeyService(repo domain.APIKeyRepository, encKey []byte) *APIKeyService {
	return &APIKeyService{repo: repo, encKey: encKey}
}

// Save validates, encrypts, and upserts an API key for the given user and provider.
func (s *APIKeyService) Save(ctx context.Context, userID string, provider domain.APIKeyProvider, plainKey string) error {
	if !provider.IsValid() {
		return fmt.Errorf("%w: unknown provider %q", ErrInvalidKeyFormat, provider)
	}
	if err := validateKeyFormat(provider, plainKey); err != nil {
		return err
	}

	encrypted, err := infra.Encrypt(plainKey, s.encKey)
	if err != nil {
		return fmt.Errorf("api key service: encrypt: %w", err)
	}

	hint := keyHint(plainKey)
	now := time.Now().UTC()

	key := &domain.UserAPIKey{
		ID:           uuid.New().String(),
		UserID:       userID,
		Provider:     provider,
		EncryptedKey: encrypted,
		KeyHint:      hint,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Upsert(ctx, key); err != nil {
		return fmt.Errorf("api key service: save: %w", err)
	}
	return nil
}

// Delete removes a user's key for a provider. Returns ErrAPIKeyNotFound if it doesn't exist.
func (s *APIKeyService) Delete(ctx context.Context, userID string, provider domain.APIKeyProvider) error {
	existing, err := s.repo.FindByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return fmt.Errorf("api key service: delete check: %w", err)
	}
	if existing == nil {
		return ErrAPIKeyNotFound
	}
	if err := s.repo.Delete(ctx, userID, provider); err != nil {
		return fmt.Errorf("api key service: delete: %w", err)
	}
	return nil
}

// ListByUserID returns metadata for all keys stored for a user. Never returns the encrypted key.
func (s *APIKeyService) ListByUserID(ctx context.Context, userID string) ([]*APIKeyInfo, error) {
	keys, err := s.repo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("api key service: list: %w", err)
	}
	infos := make([]*APIKeyInfo, 0, len(keys))
	for _, k := range keys {
		infos = append(infos, &APIKeyInfo{
			Provider:  k.Provider,
			KeyHint:   k.KeyHint,
			UpdatedAt: k.UpdatedAt,
		})
	}
	return infos, nil
}

// keyHint returns the last 4 characters of the plaintext key.
func keyHint(plainKey string) string {
	r := []rune(strings.TrimSpace(plainKey))
	if len(r) <= 4 {
		return string(r)
	}
	return string(r[len(r)-4:])
}

// validateKeyFormat checks that the key starts with the expected prefix for the provider.
func validateKeyFormat(provider domain.APIKeyProvider, key string) error {
	key = strings.TrimSpace(key)
	switch provider {
	case domain.APIKeyProviderAnthropic:
		if !strings.HasPrefix(key, "sk-ant-") {
			return fmt.Errorf("%w: Anthropic keys must start with \"sk-ant-\"", ErrInvalidKeyFormat)
		}
	case domain.APIKeyProviderGoogle:
		if !strings.HasPrefix(key, "AIza") {
			return fmt.Errorf("%w: Google AI keys must start with \"AIza\"", ErrInvalidKeyFormat)
		}
	case domain.APIKeyProviderOpenAI:
		if !strings.HasPrefix(key, "sk-") {
			return fmt.Errorf("%w: OpenAI keys must start with \"sk-\"", ErrInvalidKeyFormat)
		}
	}
	return nil
}
