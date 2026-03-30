package domain

import "context"

type GitProvider string

const (
	GitProviderGitHub GitProvider = "github"
	GitProviderGitLab GitProvider = "gitlab"
)

type TokenType string

const (
	TokenTypeOAuth TokenType = "oauth"
	TokenTypePAT   TokenType = "pat"
)

type GitIntegration struct {
	ID             string      `bson:"_id"`
	UserID         string      `bson:"user_id"`
	Provider       GitProvider `bson:"provider"`
	EncryptedToken string      `bson:"encrypted_token"`
	TokenType      TokenType   `bson:"token_type"`
}

// IntegrationRepository provides read-only access to git_integrations.
type IntegrationRepository interface {
	FindByID(ctx context.Context, id string) (*GitIntegration, error)
}
