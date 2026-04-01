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

type RepositoryHandler struct {
	svc *service.RepositoryService
}

func NewRepositoryHandler(svc *service.RepositoryService) *RepositoryHandler {
	return &RepositoryHandler{svc: svc}
}

type connectRepoRequest struct {
	IntegrationID string `json:"integration_id" binding:"required"`
	ExternalID    string `json:"external_id" binding:"required"`
	FullName      string `json:"full_name" binding:"required"`
	CloneURL      string `json:"clone_url" binding:"required"`
	DefaultBranch string `json:"default_branch"`
	Language      string `json:"language"`
	Provider      string `json:"provider" binding:"required"`
}

type repositoryResponse struct {
	ID            string `json:"id"`
	IntegrationID string `json:"integration_id"`
	Provider      string `json:"provider"`
	ExternalID    string `json:"external_id"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Language      string `json:"language"`
	ConnectedAt   string `json:"connected_at"`
}

func (h *RepositoryHandler) List(c *gin.Context) {
	uid := c.GetString(middleware.ContextUserIDKey)
	list, err := h.svc.ListConnected(c.Request.Context(), uid)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list repositories")
		return
	}
	out := make([]repositoryResponse, 0, len(list))
	for _, r := range list {
		out = append(out, repoToResponse(r))
	}
	respond.JSONOK(c, out)
}

func repoToResponse(r *domain.Repository) repositoryResponse {
	return repositoryResponse{
		ID:            r.ID,
		IntegrationID: r.IntegrationID,
		Provider:      string(r.Provider),
		ExternalID:    r.ExternalID,
		FullName:      r.FullName,
		CloneURL:      r.CloneURL,
		DefaultBranch: r.DefaultBranch,
		Language:      r.Language,
		ConnectedAt:   r.ConnectedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *RepositoryHandler) Connect(c *gin.Context) {
	var req connectRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	uid := c.GetString(middleware.ContextUserIDKey)
	provider := domain.GitProvider(req.Provider)
	if !provider.IsValid() {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "unsupported provider")
		return
	}
	repo, err := h.svc.Connect(c.Request.Context(), uid, service.ConnectRepositoryInput{
		IntegrationID: req.IntegrationID,
		ExternalID:    req.ExternalID,
		FullName:      req.FullName,
		CloneURL:      req.CloneURL,
		DefaultBranch: req.DefaultBranch,
		Language:      req.Language,
		Provider:      provider,
	})
	if err != nil {
		if errors.Is(err, service.ErrAlreadyConnectedRepo) {
			respond.JSONError(c, http.StatusConflict, "CONFLICT", "repository already connected")
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "integration not found")
			return
		}
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not connect repository")
		return
	}
	respond.JSONCreated(c, repoToResponse(repo))
}

func (h *RepositoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString(middleware.ContextUserIDKey)
	if err := h.svc.Disconnect(c.Request.Context(), uid, id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "repository not found")
			return
		}
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not disconnect")
		return
	}
	respond.JSONOK(c, gin.H{"deleted": true})
}
