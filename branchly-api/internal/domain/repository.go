package domain

import (
	"context"
	"time"
)

type Repository struct {
	ID            string    `bson:"_id"`
	UserID        string    `bson:"user_id"`
	GithubRepoID  int64     `bson:"github_repo_id"`
	FullName      string    `bson:"full_name"`
	DefaultBranch string    `bson:"default_branch"`
	Language      string    `bson:"language"`
	ConnectedAt   time.Time `bson:"connected_at"`
}

type ConnectedRepositoryRepository interface {
	Create(ctx context.Context, r *Repository) error
	FindByID(ctx context.Context, id string) (*Repository, error)
	FindByUserID(ctx context.Context, userID string) ([]*Repository, error)
	Delete(ctx context.Context, id string) error
	FindByUserAndGithubRepoID(ctx context.Context, userID string, githubRepoID int64) (*Repository, error)
}
