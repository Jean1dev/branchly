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
	GithubRepoID  int64  `json:"github_repo_id" binding:"required"`
	FullName      string `json:"full_name" binding:"required"`
	DefaultBranch string `json:"default_branch"`
	Language      string `json:"language"`
}

type repositoryResponse struct {
	ID            string `json:"id"`
	GithubRepoID  int64  `json:"github_repo_id"`
	FullName      string `json:"full_name"`
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
		GithubRepoID:  r.GithubRepoID,
		FullName:      r.FullName,
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
	repo, err := h.svc.Connect(c.Request.Context(), uid, service.ConnectRepositoryInput{
		GithubRepoID:  req.GithubRepoID,
		FullName:      req.FullName,
		DefaultBranch: req.DefaultBranch,
		Language:      req.Language,
	})
	if err != nil {
		if errors.Is(err, service.ErrAlreadyConnected) {
			respond.JSONError(c, http.StatusConflict, "CONFLICT", "repository already connected")
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

func (h *RepositoryHandler) ListGitHub(c *gin.Context) {
	uid := c.GetString(middleware.ContextUserIDKey)
	list, err := h.svc.ListGitHubAvailable(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		respond.JSONError(c, http.StatusBadGateway, "GITHUB_ERROR", "could not list github repositories")
		return
	}
	respond.JSONOK(c, list)
}
