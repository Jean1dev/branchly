package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/branchly/branchly-runner/internal/domain"
)

// Client implements domain.GitProviderClient for GitLab.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:    "https://gitlab.com",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) OpenPR(ctx context.Context, token string, opts domain.PROptions) (string, error) {
	encodedName := url.PathEscape(opts.RepoFullName)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests", c.baseURL, encodedName)

	payload := map[string]string{
		"source_branch": opts.Head,
		"target_branch": opts.Base,
		"title":         opts.Title,
		"description":   opts.Body,
		"labels":        "branchly",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gitlab client: marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gitlab client: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gitlab client: do request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("gitlab client: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gitlab client: create MR status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("gitlab client: decode response: %w", err)
	}
	return result.WebURL, nil
}
