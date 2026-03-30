package domain

import (
	"context"
	"time"
)

type Repository struct {
	ID            string      `bson:"_id"`
	UserID        string      `bson:"user_id"`
	IntegrationID string      `bson:"integration_id"`
	Provider      GitProvider `bson:"provider"`
	ExternalID    string      `bson:"external_id"`
	GithubRepoID  int64       `bson:"github_repo_id,omitempty"` // legacy; populated during migration
	FullName      string      `bson:"full_name"`
	CloneURL      string      `bson:"clone_url"`
	DefaultBranch string      `bson:"default_branch"`
	Language      string      `bson:"language"`
	ConnectedAt   time.Time   `bson:"connected_at"`
}

type ConnectedRepositoryRepository interface {
	Create(ctx context.Context, r *Repository) error
	FindByID(ctx context.Context, id string) (*Repository, error)
	FindByUserID(ctx context.Context, userID string) ([]*Repository, error)
	Delete(ctx context.Context, id string) error
	FindByUserExternalAndProvider(ctx context.Context, userID, externalID string, provider GitProvider) (*Repository, error)
	FindByIntegrationID(ctx context.Context, integrationID string) ([]*Repository, error)
}
