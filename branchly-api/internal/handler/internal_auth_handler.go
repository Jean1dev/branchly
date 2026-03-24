package handler

import (
	"net/http"

	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

type InternalAuthHandler struct {
	auth *service.AuthService
}

func NewInternalAuthHandler(auth *service.AuthService) *InternalAuthHandler {
	return &InternalAuthHandler{auth: auth}
}

type internalAuthUpsertRequest struct {
	ProviderID  string  `json:"provider_id" binding:"required"`
	GithubToken string  `json:"github_token" binding:"required"`
	Email       *string `json:"email"`
	Name        *string `json:"name"`
	AvatarURL   *string `json:"avatar_url"`
}

type internalAuthUpsertResponse struct {
	UserID        string `json:"user_id"`
	InternalToken string `json:"internal_token"`
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (h *InternalAuthHandler) Upsert(c *gin.Context) {
	var req internalAuthUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}
	ctx := c.Request.Context()
	user, err := h.auth.UpsertFromInternalOAuth(
		ctx,
		req.ProviderID,
		strPtr(req.Email),
		strPtr(req.Name),
		strPtr(req.AvatarURL),
		req.GithubToken,
	)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not upsert user")
		return
	}
	jwtStr, err := h.auth.IssueAccessToken(user)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not issue session")
		return
	}
	respond.JSONOK(c, internalAuthUpsertResponse{
		UserID:        user.ID,
		InternalToken: jwtStr,
	})
}
