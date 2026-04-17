package domain

import "context"

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
	case AgentTypeGPTCodex:
		return APIKeyProviderOpenAI
	default:
		return ""
	}
}

// UserAPIKey is the minimal projection of a user API key needed by the runner.
type UserAPIKey struct {
	ID           string         `bson:"_id"`
	UserID       string         `bson:"user_id"`
	Provider     APIKeyProvider `bson:"provider"`
	EncryptedKey string         `bson:"encrypted_key"`
}

// APIKeyRepository provides read-only access to user API keys.
type APIKeyRepository interface {
	FindByUserAndProvider(ctx context.Context, userID string, provider APIKeyProvider) (*UserAPIKey, error)
}
