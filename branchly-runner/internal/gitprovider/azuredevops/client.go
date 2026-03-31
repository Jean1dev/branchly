package azuredevops

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

// Client implements domain.GitProviderClient for Azure DevOps.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:    "https://dev.azure.com",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func basicAuth(pat string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+pat))
}

// OpenPR creates a pull request on Azure DevOps.
// opts.RepoFullName must be in the format "{org}/{project}/{repo}".
func (c *Client) OpenPR(ctx context.Context, token string, opts domain.PROptions) (string, error) {
	parts := strings.SplitN(opts.RepoFullName, "/", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("azuredevops client: invalid repo full name %q, expected org/project/repo", opts.RepoFullName)
	}
	org, project, repo := parts[0], parts[1], parts[2]

	apiURL := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/pullrequests?api-version=7.0",
		c.baseURL, org, project, repo)

	payload := map[string]any{
		"title":         opts.Title,
		"description":   opts.Body,
		"sourceRefName": "refs/heads/" + opts.Head,
		"targetRefName": "refs/heads/" + opts.Base,
		"labels":        []map[string]string{{"name": "branchly"}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("azuredevops client: marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("azuredevops client: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", basicAuth(token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("azuredevops client: do request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("azuredevops client: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("azuredevops client: create PR status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		PullRequestID int `json:"pullRequestId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("azuredevops client: decode response: %w", err)
	}

	prURL := fmt.Sprintf("%s/%s/%s/_git/%s/pullrequest/%d",
		c.baseURL, org, project, repo, result.PullRequestID)
	return prURL, nil
}
