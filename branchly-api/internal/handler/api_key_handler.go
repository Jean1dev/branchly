package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

type apiKeyService interface {
	Save(ctx context.Context, userID string, provider domain.APIKeyProvider, plainKey string) error
	Delete(ctx context.Context, userID string, provider domain.APIKeyProvider) error
	ListByUserID(ctx context.Context, userID string) ([]*service.APIKeyInfo, error)
}

type APIKeyHandler struct {
	svc apiKeyService
}

func NewAPIKeyHandler(svc apiKeyService) *APIKeyHandler {
	return &APIKeyHandler{svc: svc}
}

type apiKeyInfoResponse struct {
	Provider  string    `json:"provider"`
	KeyHint   string    `json:"key_hint"`
	UpdatedAt time.Time `json:"updated_at"`
}

// List returns metadata for all API keys stored by the authenticated user.
// GET /settings/api-keys
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	infos, err := h.svc.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not list api keys")
		return
	}
	resp := make([]apiKeyInfoResponse, 0, len(infos))
	for _, info := range infos {
		resp = append(resp, apiKeyInfoResponse{
			Provider:  string(info.Provider),
			KeyHint:   info.KeyHint,
			UpdatedAt: info.UpdatedAt,
		})
	}
	respond.JSONOK(c, resp)
}

type saveAPIKeyRequest struct {
	Key string `json:"key" binding:"required"`
}

// Save stores or replaces the user's key for a provider.
// PUT /settings/api-keys/:provider
func (h *APIKeyHandler) Save(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	providerParam := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	provider := domain.APIKeyProvider(providerParam)
	if !provider.IsValid() {
		respond.JSONError(c, http.StatusBadRequest, "INVALID_PROVIDER", "unknown provider: "+providerParam)
		return
	}

	var req saveAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "key is required")
		return
	}

	if err := h.svc.Save(c.Request.Context(), userID, provider, req.Key); err != nil {
		if errors.Is(err, service.ErrInvalidKeyFormat) {
			respond.JSONError(c, http.StatusBadRequest, "INVALID_KEY_FORMAT", err.Error())
			return
		}
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not save api key")
		return
	}

	// Return updated metadata.
	infos, err := h.svc.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not fetch api key info")
		return
	}
	for _, info := range infos {
		if info.Provider == provider {
			respond.JSONOK(c, apiKeyInfoResponse{
				Provider:  string(info.Provider),
				KeyHint:   info.KeyHint,
				UpdatedAt: info.UpdatedAt,
			})
			return
		}
	}
	c.Status(http.StatusOK)
}

// Delete removes the user's key for a provider.
// DELETE /settings/api-keys/:provider
func (h *APIKeyHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	providerParam := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	provider := domain.APIKeyProvider(providerParam)
	if !provider.IsValid() {
		respond.JSONError(c, http.StatusBadRequest, "INVALID_PROVIDER", "unknown provider: "+providerParam)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, provider); err != nil {
		if errors.Is(err, service.ErrAPIKeyNotFound) {
			respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "api key not found")
			return
		}
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not delete api key")
		return
	}
	c.Status(http.StatusNoContent)
}
