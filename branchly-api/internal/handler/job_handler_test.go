package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

// ---- mock service ----

type mockJobSvc struct {
	job *domain.Job
	err error
}

func (m *mockJobSvc) Create(_ context.Context, _ string, _ service.CreateJobInput) (*domain.Job, error) {
	return m.job, m.err
}
func (m *mockJobSvc) List(_ context.Context, _ string, _ *domain.JobStatus, _ *string) ([]*domain.Job, error) {
	return nil, nil
}
func (m *mockJobSvc) Get(_ context.Context, _, _ string) (*domain.Job, error) {
	return nil, nil
}
func (m *mockJobSvc) Retry(_ context.Context, _, _ string) (*domain.Job, error) {
	return m.job, m.err
}

// ---- helpers ----

func testRouter(svc *mockJobSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewJobHandler(svc)
	r := gin.New()
	r.POST("/jobs", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
	}, h.Create)
	return r
}

func postJob(r *gin.Engine, body map[string]any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func aJob() *domain.Job {
	return &domain.Job{
		ID:           "job-1",
		UserID:       "user-1",
		RepositoryID: "repo-1",
		Prompt:       "add feature",
		AgentType:    domain.AgentTypeClaudeCode,
		Status:       domain.JobStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ---- tests ----

func TestCreate_AgentTypeClaudeCode_Accepted(t *testing.T) {
	r := testRouter(&mockJobSvc{job: aJob()})
	w := postJob(r, map[string]any{
		"repository_id": "repo-1",
		"prompt":        "add feature",
		"agent_type":    "claude-code",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreate_AgentTypeGemini_Accepted(t *testing.T) {
	j := aJob()
	j.AgentType = domain.AgentTypeGemini
	r := testRouter(&mockJobSvc{job: j})
	w := postJob(r, map[string]any{
		"repository_id": "repo-1",
		"prompt":        "add feature",
		"agent_type":    "gemini",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreate_EmptyAgentType_Returns400(t *testing.T) {
	r := testRouter(&mockJobSvc{})
	w := postJob(r, map[string]any{
		"repository_id": "repo-1",
		"prompt":        "add feature",
		"agent_type":    "",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	checkErrorCode(t, w.Body.Bytes(), "INVALID_AGENT_TYPE")
}

func TestCreate_UnknownAgentType_Returns400(t *testing.T) {
	r := testRouter(&mockJobSvc{})
	w := postJob(r, map[string]any{
		"repository_id": "repo-1",
		"prompt":        "add feature",
		"agent_type":    "openai",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	checkErrorCode(t, w.Body.Bytes(), "INVALID_AGENT_TYPE")
}

func TestCreate_MissingAgentType_Returns400(t *testing.T) {
	r := testRouter(&mockJobSvc{})
	w := postJob(r, map[string]any{
		"repository_id": "repo-1",
		"prompt":        "add feature",
		// agent_type omitted
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// checkErrorCode asserts that the JSON body contains the expected error code.
func checkErrorCode(t *testing.T, body []byte, wantCode string) {
	t.Helper()
	var out struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("could not parse error body: %v — raw: %s", err, body)
	}
	if out.Error.Code != wantCode {
		t.Errorf("expected error code %q, got %q", wantCode, out.Error.Code)
	}
}
