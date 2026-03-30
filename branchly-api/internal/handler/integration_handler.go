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

type IntegrationHandler struct {
	integSvc *service.IntegrationService
	repoSvc  *service.RepositoryService
}

func NewIntegrationHandler(integSvc *service.IntegrationService, repoSvc *service.RepositoryService) *IntegrationHandler {
	return &IntegrationHandler{integSvc: integSvc, repoSvc: repoSvc}
}

type integrationResponse struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	TokenType   string `json:"token_type"`
	ConnectedAt string `json:"connected_at"`
}

func toIntegrationResponse(ig *domain.GitIntegration) integrationResponse {
	return integrationResponse{
		ID:          ig.ID,
		Provider:    string(ig.Provider),
		TokenType:   string(ig.TokenType),
		ConnectedAt: ig.ConnectedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *IntegrationHandler) List(c *gin.Context) {
	uid := c.GetString(middleware.ContextUserIDKey)
	list, err := h.integSvc.ListByUserID(c.Request.Context(), uid)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list integrations")
		return
	}
	out := make([]integrationResponse, 0, len(list))
	for _, ig := range list {
		out = append(out, toIntegrationResponse(ig))
	}
	respond.JSONOK(c, out)
}

type connectGitLabRequest struct {
	PAT string `json:"pat" binding:"required"`
}

func (h *IntegrationHandler) ConnectGitLab(c *gin.Context) {
	var req connectGitLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	uid := c.GetString(middleware.ContextUserIDKey)
	ig, err := h.integSvc.ConnectGitLab(c.Request.Context(), uid, req.PAT)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			respond.JSONError(c, http.StatusUnprocessableEntity, "INVALID_TOKEN", "invalid or expired GitLab token")
			return
		}
		if errors.Is(err, service.ErrAlreadyConnected) {
			respond.JSONError(c, http.StatusConflict, "CONFLICT", "GitLab already connected")
			return
		}
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not connect GitLab")
		return
	}
	respond.JSONCreated(c, toIntegrationResponse(ig))
}

func (h *IntegrationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString(middleware.ContextUserIDKey)
	err := h.integSvc.Disconnect(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "integration not found")
			return
		}
		if errors.Is(err, service.ErrIntegrationInUse) {
			respond.JSONError(c, http.StatusConflict, "INTEGRATION_IN_USE", "integration has connected repositories")
			return
		}
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not disconnect integration")
		return
	}
	respond.JSONOK(c, gin.H{"deleted": true})
}

type providerRepoResponse struct {
	ExternalID    string `json:"external_id"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Language      string `json:"language"`
	Provider      string `json:"provider"`
}

func (h *IntegrationHandler) ListRepositories(c *gin.Context) {
	id := c.Param("id")
	uid := c.GetString(middleware.ContextUserIDKey)
	repos, err := h.repoSvc.ListFromProvider(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "integration not found")
			return
		}
		respond.JSONError(c, http.StatusBadGateway, "PROVIDER_ERROR", "could not list repositories from provider")
		return
	}
	out := make([]providerRepoResponse, 0, len(repos))
	for _, r := range repos {
		out = append(out, providerRepoResponse{
			ExternalID:    r.ExternalID,
			FullName:      r.FullName,
			CloneURL:      r.CloneURL,
			DefaultBranch: r.DefaultBranch,
			Language:      r.Language,
			Provider:      string(r.Provider),
		})
	}
	respond.JSONOK(c, out)
}
