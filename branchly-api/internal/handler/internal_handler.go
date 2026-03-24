package handler

import (
	"net/http"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

type InternalHandler struct {
	jobs *service.JobService
}

func NewInternalHandler(jobs *service.JobService) *InternalHandler {
	return &InternalHandler{jobs: jobs}
}

type internalStatusRequest struct {
	Status     string `json:"status" binding:"required"`
	PRUrl      string `json:"pr_url"`
	BranchName string `json:"branch_name"`
}

func (h *InternalHandler) UpdateStatus(c *gin.Context) {
	jobID := c.Param("id")
	var req internalStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	st := domain.JobStatus(req.Status)
	switch st {
	case domain.JobStatusPending, domain.JobStatusRunning, domain.JobStatusCompleted, domain.JobStatusFailed:
	default:
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid status")
		return
	}
	if err := h.jobs.UpdateStatusInternal(c.Request.Context(), jobID, st, req.PRUrl, req.BranchName); err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update job")
		return
	}
	respond.JSONOK(c, gin.H{"ok": true})
}

type internalLogRequest struct {
	Timestamp *time.Time `json:"timestamp"`
	Level     string     `json:"level" binding:"required"`
	Message   string     `json:"message" binding:"required"`
}

func (h *InternalHandler) AppendLog(c *gin.Context) {
	jobID := c.Param("id")
	var req internalLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	lvl := domain.LogLevel(req.Level)
	switch lvl {
	case domain.LogLevelInfo, domain.LogLevelSuccess, domain.LogLevelError:
	default:
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid log level")
		return
	}
	ts := time.Now().UTC()
	if req.Timestamp != nil {
		ts = req.Timestamp.UTC()
	}
	entry := domain.LogEntry{
		Timestamp: ts,
		Level:     lvl,
		Message:   req.Message,
	}
	if err := h.jobs.AppendLogInternal(c.Request.Context(), jobID, entry); err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not append log")
		return
	}
	respond.JSONOK(c, gin.H{"ok": true})
}
