package domain

import "time"

type GitProvider string

const (
	GitProviderGitHub GitProvider = "github"
	GitProviderGitLab GitProvider = "gitlab"
)

func (p GitProvider) IsValid() bool {
	switch p {
	case GitProviderGitHub, GitProviderGitLab:
		return true
	}
	return false
}

func (p GitProvider) DisplayName() string {
	switch p {
	case GitProviderGitHub:
		return "GitHub"
	case GitProviderGitLab:
		return "GitLab"
	}
	return string(p)
}

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
	Scopes         []string    `bson:"scopes"`
	ExpiresAt      *time.Time  `bson:"expires_at,omitempty"`
	ConnectedAt    time.Time   `bson:"connected_at"`
}
