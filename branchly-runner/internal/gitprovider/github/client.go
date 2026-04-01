package github

import (
	"context"
	"fmt"

	"github.com/branchly/branchly-runner/internal/domain"
	gogithub "github.com/google/go-github/v63/github"
)

// Client implements domain.GitProviderClient for GitHub.
type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) OpenPR(ctx context.Context, token string, opts domain.PROptions) (string, error) {
	owner, repo, err := splitRepo(opts.RepoFullName)
	if err != nil {
		return "", fmt.Errorf("github client: split repo: %w", err)
	}

	gh := gogithub.NewClient(nil).WithAuthToken(token)
	pr, _, err := gh.PullRequests.Create(ctx, owner, repo, &gogithub.NewPullRequest{
		Title: &opts.Title,
		Head:  &opts.Head,
		Base:  &opts.Base,
		Body:  &opts.Body,
	})
	if err != nil {
		return "", fmt.Errorf("github client: create PR: %w", err)
	}
	return pr.GetHTMLURL(), nil
}

func splitRepo(fullName string) (owner, repo string, err error) {
	for i, c := range fullName {
		if c == '/' {
			if i == 0 || i == len(fullName)-1 {
				break
			}
			return fullName[:i], fullName[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("invalid repository name: %q", fullName)
}
