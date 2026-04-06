package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

// jobService is the subset of service.JobService used by the handler.
type jobService interface {
	Create(ctx context.Context, userID string, in service.CreateJobInput) (*domain.Job, error)
	List(ctx context.Context, userID string, status *domain.JobStatus, repositoryID *string) ([]*domain.Job, error)
	Get(ctx context.Context, userID, jobID string) (*domain.Job, error)
	Retry(ctx context.Context, userID string, jobID string) (*domain.Job, error)
}

type JobHandler struct {
	svc jobService
}

func NewJobHandler(svc jobService) *JobHandler {
	return &JobHandler{svc: svc}
}

type createJobRequest struct {
	RepositoryID string `json:"repository_id" binding:"required"`
	Prompt       string `json:"prompt" binding:"required"`
	AgentType    string `json:"agent_type"`
}

type logEntryResponse struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type jobCostResponse struct {
	AgentType    string  `json:"agent_type"`
	ModelUsed    string  `json:"model_used"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	EstimatedUSD float64 `json:"estimated_usd"`
	DurationSecs float64 `json:"duration_secs"`
	IsEstimate   bool    `json:"is_estimate"`
}

type jobResponse struct {
	ID            string             `json:"id"`
	RepositoryID  string             `json:"repository_id"`
	Prompt        string             `json:"prompt"`
	Status        string             `json:"status"`
	AgentType     string             `json:"agent_type"`
	BranchName    string             `json:"branch_name"`
	PRUrl         string             `json:"pr_url,omitempty"`
	Logs          []logEntryResponse `json:"logs,omitempty"`
	Cost          *jobCostResponse   `json:"cost,omitempty"`
	AttemptNumber int                `json:"attempt_number"`
	MaxAttempts   int                `json:"max_attempts"`
	LastError     string             `json:"last_error,omitempty"`
	NextRetryAt   *string            `json:"next_retry_at,omitempty"`
	FailureType   string             `json:"failure_type,omitempty"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
	CompletedAt   *string            `json:"completed_at,omitempty"`
}

func jobToResponse(j *domain.Job) jobResponse {
	logs := make([]logEntryResponse, 0, len(j.Logs))
	for _, e := range j.Logs {
		logs = append(logs, logEntryResponse{
			Timestamp: e.Timestamp.UTC().Format("2006-01-02T15:04:05Z07:00"),
			Level:     string(e.Level),
			Message:   e.Message,
		})
	}
	var completed *string
	if j.CompletedAt != nil {
		s := j.CompletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		completed = &s
	}
	var nextRetryAt *string
	if j.NextRetryAt != nil {
		s := j.NextRetryAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		nextRetryAt = &s
	}
	var cost *jobCostResponse
	if j.Cost != nil {
		cost = &jobCostResponse{
			AgentType:    string(j.Cost.AgentType),
			ModelUsed:    j.Cost.ModelUsed,
			InputTokens:  j.Cost.InputTokens,
			OutputTokens: j.Cost.OutputTokens,
			TotalTokens:  j.Cost.TotalTokens,
			EstimatedUSD: j.Cost.EstimatedUSD,
			DurationSecs: j.Cost.DurationSecs,
			IsEstimate:   true,
		}
	}
	return jobResponse{
		ID:            j.ID,
		RepositoryID:  j.RepositoryID,
		Prompt:        j.Prompt,
		Status:        string(j.Status),
		AgentType:     string(j.AgentType),
		BranchName:    j.BranchName,
		PRUrl:         j.PRUrl,
		Logs:          logs,
		Cost:          cost,
		AttemptNumber: j.AttemptNumber,
		MaxAttempts:   j.MaxAttempts,
		LastError:     j.LastError,
		NextRetryAt:   nextRetryAt,
		FailureType:   string(j.FailureType),
		CreatedAt:     j.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     j.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt:   completed,
	}
}

func (h *JobHandler) Create(c *gin.Context) {
	var req createJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	if !domain.AgentType(req.AgentType).IsValid() {
		respond.JSONError(c, http.StatusBadRequest, "INVALID_AGENT_TYPE",
			"agent_type must be one of: claude-code, gemini")
		return
	}
	uid := c.GetString(middleware.ContextUserIDKey)
	job, err := h.svc.Create(c.Request.Context(), uid, service.CreateJobInput{
		RepositoryID: req.RepositoryID,
		Prompt:       req.Prompt,
		AgentType:    domain.AgentType(req.AgentType),
	})
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "repository not found")
			return
		}
		if errors.Is(err, service.ErrRateLimitExceeded) {
			respond.JSONError(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				"you have reached the maximum of 3 active jobs — wait for one to complete")
			return
		}
		respond.JSONError(c, http.StatusBadGateway, "RUNNER_ERROR", "could not start job")
		return
	}
	respond.JSONCreated(c, jobToResponse(job))
}

func (h *JobHandler) List(c *gin.Context) {
	uid := c.GetString(middleware.ContextUserIDKey)
	statusStr := c.Query("status")
	repoID := c.Query("repository_id")
	var st *domain.JobStatus
	if statusStr != "" {
		s := domain.JobStatus(statusStr)
		st = &s
	}
	var rid *string
	if repoID != "" {
		rid = &repoID
	}
	list, err := h.svc.List(c.Request.Context(), uid, st, rid)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list jobs")
		return
	}
	out := make([]jobResponse, 0, len(list))
	for _, j := range list {
		jr := jobToResponse(j)
		jr.Logs = nil
		out = append(out, jr)
	}
	respond.JSONOK(c, out)
}

func (h *JobHandler) Get(c *gin.Context) {
	uid := c.GetString(middleware.ContextUserIDKey)
	id := c.Param("id")
	j, err := h.svc.Get(c.Request.Context(), uid, id)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load job")
		return
	}
	if j == nil {
		respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "job not found")
		return
	}
	respond.JSONOK(c, jobToResponse(j))
}

func (h *JobHandler) Retry(c *gin.Context) {
	uid := c.GetString(middleware.ContextUserIDKey)
	id := c.Param("id")
	j, err := h.svc.Retry(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, service.ErrJobNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "job not found")
			return
		}
		if errors.Is(err, domain.ErrJobNotRetryable) {
			respond.JSONError(c, http.StatusConflict, "NOT_RETRYABLE",
				"Job cannot be retried. Only permanently failed jobs support manual retry.")
			return
		}
		respond.JSONError(c, http.StatusBadGateway, "RUNNER_ERROR", "could not retry job")
		return
	}
	respond.JSONOK(c, jobToResponse(j))
}
