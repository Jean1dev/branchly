package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type DispatchJobPayload struct {
	JobID          string `json:"job_id"`
	UserID         string `json:"user_id"`
	RepositoryID   string `json:"repository_id"`
	RepositoryName string `json:"repository_name"`
	DefaultBranch  string `json:"default_branch"`
	Prompt         string `json:"prompt"`
	IntegrationID  string `json:"integration_id"`
	Provider       string `json:"provider"`
	AgentType      string `json:"agent_type"`
}

type RunnerClient struct {
	baseURL      string
	runnerSecret string
	httpClient   *http.Client
}

func NewRunnerClient(baseURL, runnerSecret string) *RunnerClient {
	return &RunnerClient{
		baseURL:      baseURL,
		runnerSecret: runnerSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *RunnerClient) DispatchJob(ctx context.Context, payload DispatchJobPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("runner client: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/jobs", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("runner client: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Runner-Secret", c.runnerSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("runner client: do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		slog.Warn("runner dispatch non-success", "status", resp.StatusCode, "body_len", len(b))
		return fmt.Errorf("runner client: status %d", resp.StatusCode)
	}
	return nil
}
