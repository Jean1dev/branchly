package handler

import (
	"context"
	"crypto/subtle"
	"encoding/binary"
	"log/slog"
	"net/http"
	"strings"

	"github.com/branchly/branchly-runner/internal/domain"
	"github.com/branchly/branchly-runner/internal/executor"
	"github.com/branchly/branchly-runner/internal/pool"
	"github.com/gin-gonic/gin"
)

type JobsHandler struct {
	secret   string
	pool     *pool.Pool
	executor *executor.Executor
}

func NewJobsHandler(secret string, p *pool.Pool, ex *executor.Executor) *JobsHandler {
	return &JobsHandler{secret: secret, pool: p, executor: ex}
}

type postJobBody struct {
	JobID          string `json:"job_id" binding:"required"`
	UserID         string `json:"user_id" binding:"required"`
	RepositoryID   string `json:"repository_id" binding:"required"`
	RepositoryName string `json:"repository_name" binding:"required"`
	DefaultBranch  string `json:"default_branch"`
	Prompt         string `json:"prompt" binding:"required"`
	EncryptedToken string `json:"encrypted_token" binding:"required"`
	AgentType      string `json:"agent_type"`
}

const maxRunnerSecretBytes = 512

func secretsEqual(expected, got string) bool {
	eb := []byte(expected)
	gb := []byte(got)
	if len(eb) > maxRunnerSecretBytes || len(gb) > maxRunnerSecretBytes {
		return false
	}
	var pe, pg [8 + maxRunnerSecretBytes]byte
	binary.BigEndian.PutUint64(pe[:8], uint64(len(eb)))
	binary.BigEndian.PutUint64(pg[:8], uint64(len(gb)))
	copy(pe[8:], eb)
	copy(pg[8:], gb)
	return subtle.ConstantTimeCompare(pe[:], pg[:]) == 1
}

func (h *JobsHandler) PostJob(c *gin.Context) {
	got := strings.TrimSpace(c.GetHeader("X-Runner-Secret"))
	if got == "" || !secretsEqual(h.secret, got) {
		c.Status(http.StatusUnauthorized)
		return
	}
	var body postJobBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	in := executor.RunJobInput{
		JobID:          strings.TrimSpace(body.JobID),
		UserID:         strings.TrimSpace(body.UserID),
		RepositoryID:   strings.TrimSpace(body.RepositoryID),
		RepositoryName: strings.TrimSpace(body.RepositoryName),
		DefaultBranch:  strings.TrimSpace(body.DefaultBranch),
		Prompt:         body.Prompt,
		EncryptedToken: body.EncryptedToken,
		AgentType:      domain.AgentType(strings.TrimSpace(body.AgentType)),
	}
	ok := h.pool.TryGo(func() {
		h.executor.Run(context.Background(), in)
	})
	if !ok {
		c.Status(http.StatusTooManyRequests)
		return
	}
	slog.Info("job accepted", "job_id", in.JobID)
	c.Status(http.StatusAccepted)
}
