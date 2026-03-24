package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SSEHandler struct {
	jobs *service.JobService
}

func NewSSEHandler(jobs *service.JobService) *SSEHandler {
	return &SSEHandler{jobs: jobs}
}

type sseLogPayload struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type sseDonePayload struct {
	Status string `json:"status"`
}

const sseInitialLogCap = 25000
const ssePollBatch = 400

func (h *SSEHandler) StreamJobLogs(c *gin.Context) {
	if _, ok := c.Writer.(http.Flusher); !ok {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "streaming unsupported")
		return
	}
	uid := c.GetString(middleware.ContextUserIDKey)
	jobID := c.Param("id")
	ctx := c.Request.Context()
	job, err := h.jobs.JobMeta(ctx, uid, jobID)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load job")
		return
	}
	if job == nil {
		respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "job not found")
		return
	}

	initial, err := h.jobs.ListJobLogsAsc(ctx, uid, jobID, sseInitialLogCap)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load logs")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	rc := http.NewResponseController(c.Writer)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "streaming setup failed")
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
	flusher := c.Writer.(http.Flusher)

	sendLog := func(e domain.LogEntry) {
		p := sseLogPayload{
			Timestamp: e.Timestamp.UTC().Format(time.RFC3339Nano),
			Level:     string(e.Level),
			Message:   e.Message,
		}
		b, err := json.Marshal(p)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", b)
		flusher.Flush()
	}
	var lastID primitive.ObjectID
	if len(initial) > 0 {
		for _, row := range initial {
			sendLog(row.Entry)
			lastID = row.ID
		}
	} else {
		for _, e := range job.Logs {
			sendLog(e)
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, j, err := h.jobs.ListJobLogsAfter(ctx, uid, jobID, lastID, ssePollBatch)
			if err != nil || j == nil {
				return
			}
			for _, row := range rows {
				sendLog(row.Entry)
				lastID = row.ID
			}
			if j.Status == domain.JobStatusCompleted || j.Status == domain.JobStatusFailed {
				done, _ := json.Marshal(sseDonePayload{Status: string(j.Status)})
				_, _ = fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", done)
				flusher.Flush()
				return
			}
		}
	}
}
