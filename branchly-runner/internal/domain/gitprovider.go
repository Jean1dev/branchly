package domain

import "context"

// PROptions contains the parameters needed to open a pull/merge request.
type PROptions struct {
	RepoFullName string
	Title        string
	Body         string
	Head         string
	Base         string
}

// GitProviderClient abstracts PR/MR creation across Git hosting providers.
type GitProviderClient interface {
	OpenPR(ctx context.Context, token string, opts PROptions) (prURL string, err error)
}
