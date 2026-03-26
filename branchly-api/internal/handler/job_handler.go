package handler

import (
	"errors"
	"net/http"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	svc *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{svc: svc}
}

type createJobRequest struct {
	RepositoryID string `json:"repository_id" binding:"required"`
	Prompt       string `json:"prompt" binding:"required"`
}

type logEntryResponse struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type jobResponse struct {
	ID           string              `json:"id"`
	RepositoryID string              `json:"repository_id"`
	Prompt       string              `json:"prompt"`
	Status       string              `json:"status"`
	BranchName   string              `json:"branch_name"`
	PRUrl        string              `json:"pr_url,omitempty"`
	Logs         []logEntryResponse  `json:"logs,omitempty"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	CompletedAt  *string             `json:"completed_at,omitempty"`
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
	return jobResponse{
		ID:           j.ID,
		RepositoryID: j.RepositoryID,
		Prompt:       j.Prompt,
		Status:       string(j.Status),
		BranchName:   j.BranchName,
		PRUrl:        j.PRUrl,
		Logs:         logs,
		CreatedAt:    j.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    j.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt:  completed,
	}
}

func (h *JobHandler) Create(c *gin.Context) {
	var req createJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	uid := c.GetString(middleware.ContextUserIDKey)
	job, err := h.svc.Create(c.Request.Context(), uid, service.CreateJobInput{
		RepositoryID: req.RepositoryID,
		Prompt:       req.Prompt,
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
