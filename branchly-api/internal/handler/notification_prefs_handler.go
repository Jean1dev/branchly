package handler

import (
	"net/http"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/gin-gonic/gin"
)

type NotificationPrefsHandler struct {
	users domain.UserRepository
}

func NewNotificationPrefsHandler(users domain.UserRepository) *NotificationPrefsHandler {
	return &NotificationPrefsHandler{users: users}
}

func (h *NotificationPrefsHandler) Get(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	user, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load user")
		return
	}
	respond.JSONOK(c, gin.H{"notification_preferences": user.NotificationPreferences})
}

// GetInternal is called server-to-server (internal secret auth) to fetch notification
// preferences and email address for a given user ID.
func (h *NotificationPrefsHandler) GetInternal(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	respond.JSONOK(c, gin.H{
		"email":                    user.Email,
		"name":                     user.Name,
		"notification_preferences": user.NotificationPreferences,
	})
}

type patchNotificationPrefsRequest struct {
	Enabled        *bool `json:"enabled"`
	OnJobCompleted *bool `json:"on_job_completed"`
	OnJobFailed    *bool `json:"on_job_failed"`
	OnPROpened     *bool `json:"on_pr_opened"`
}

func (h *NotificationPrefsHandler) Patch(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	user, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load user")
		return
	}

	var req patchNotificationPrefsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid body")
		return
	}

	prefs := user.NotificationPreferences
	if req.Enabled != nil {
		prefs.Email.Enabled = *req.Enabled
	}
	if req.OnJobCompleted != nil {
		prefs.Email.OnJobCompleted = *req.OnJobCompleted
	}
	if req.OnJobFailed != nil {
		prefs.Email.OnJobFailed = *req.OnJobFailed
	}
	if req.OnPROpened != nil {
		prefs.Email.OnPROpened = *req.OnPROpened
	}

	if err := h.users.UpdateNotificationPreferences(c.Request.Context(), userID, prefs); err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not update preferences")
		return
	}
	respond.JSONOK(c, gin.H{"notification_preferences": prefs})
}
