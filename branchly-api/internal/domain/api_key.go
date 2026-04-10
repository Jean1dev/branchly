package domain

import (
	"context"
	"time"
)

type APIKeyProvider string

const (
	APIKeyProviderAnthropic APIKeyProvider = "anthropic"
	APIKeyProviderGoogle    APIKeyProvider = "google"
	APIKeyProviderOpenAI    APIKeyProvider = "openai"
)

func (p APIKeyProvider) IsValid() bool {
	switch p {
	case APIKeyProviderAnthropic, APIKeyProviderGoogle, APIKeyProviderOpenAI:
		return true
	}
	return false
}

// RequiredKeyProvider maps an agent type to the API key provider it needs.
func RequiredKeyProvider(agent AgentType) APIKeyProvider {
	switch agent {
	case AgentTypeClaudeCode:
		return APIKeyProviderAnthropic
	case AgentTypeGemini:
		return APIKeyProviderGoogle
	default:
		return ""
	}
}

// UserAPIKey stores an encrypted API key for a given provider, scoped to a user.
// The plaintext key is never stored — only EncryptedKey and a KeyHint (last 4 chars).
type UserAPIKey struct {
	ID           string         `bson:"_id"`
	UserID       string         `bson:"user_id"`
	Provider     APIKeyProvider `bson:"provider"`
	EncryptedKey string         `bson:"encrypted_key"`
	KeyHint      string         `bson:"key_hint"` // last 4 chars of the plaintext key
	CreatedAt    time.Time      `bson:"created_at"`
	UpdatedAt    time.Time      `bson:"updated_at"`
}

type APIKeyRepository interface {
	// Upsert inserts or replaces the key for (user_id, provider).
	Upsert(ctx context.Context, key *UserAPIKey) error

	FindByUserAndProvider(ctx context.Context, userID string, provider APIKeyProvider) (*UserAPIKey, error)

	FindAllByUserID(ctx context.Context, userID string) ([]*UserAPIKey, error)

	Delete(ctx context.Context, userID string, provider APIKeyProvider) error
}
