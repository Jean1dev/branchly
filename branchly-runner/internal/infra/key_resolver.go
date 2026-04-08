package infra

import (
	"context"
	"fmt"

	"github.com/branchly/branchly-runner/internal/domain"
)

// KeyResolver resolves which API key to use for a job, following the priority:
//  1. User's own key stored in the database (decrypted at runtime)
//  2. Global fallback key from environment configuration
//  3. No key available → error (job fails with clear message)
type KeyResolver struct {
	apiKeyRepo    domain.APIKeyRepository
	encryptionKey []byte
	globalKeys    map[domain.APIKeyProvider]string
}

func NewKeyResolver(repo domain.APIKeyRepository, encKey []byte, globalKeys map[domain.APIKeyProvider]string) *KeyResolver {
	return &KeyResolver{
		apiKeyRepo:    repo,
		encryptionKey: encKey,
		globalKeys:    globalKeys,
	}
}

// Resolve returns the plaintext API key for the given user and provider.
// The caller is responsible for zeroing the returned string after use.
func (r *KeyResolver) Resolve(ctx context.Context, userID string, provider domain.APIKeyProvider) (string, error) {
	// 1. Try user's own key.
	userKey, err := r.apiKeyRepo.FindByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return "", fmt.Errorf("key resolver: lookup user key: %w", err)
	}
	if userKey != nil {
		plain, err := Decrypt(userKey.EncryptedKey, r.encryptionKey)
		if err != nil {
			// Corrupted key is a hard failure — do NOT silently fall back to global.
			return "", fmt.Errorf("key resolver: decrypt user key: %w", err)
		}
		return plain, nil
	}

	// 2. Fallback to global key.
	if globalKey, ok := r.globalKeys[provider]; ok && globalKey != "" {
		return globalKey, nil
	}

	// 3. No key available.
	return "", fmt.Errorf(
		"key resolver: no API key available for provider %q — configure your own key at Settings → API keys",
		provider,
	)
}
