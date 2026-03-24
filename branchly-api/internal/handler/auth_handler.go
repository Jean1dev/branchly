package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

const oauthStateCookie = "branchly_oauth_state"

type AuthHandler struct {
	cfg  *config.Config
	auth *service.AuthService
}

func NewAuthHandler(cfg *config.Config, auth *service.AuthService) *AuthHandler {
	return &AuthHandler{cfg: cfg, auth: auth}
}

func (h *AuthHandler) GitHubStart(c *gin.Context) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not generate state")
		return
	}
	state := hex.EncodeToString(b)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.CookieSecure,
	})
	url := h.auth.OAuthAuthCodeURL(state)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	stateQ := c.Query("state")
	code := c.Query("code")
	if code == "" || stateQ == "" {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "missing code or state")
		return
	}
	ck, err := c.Request.Cookie(oauthStateCookie)
	if err != nil || ck.Value == "" || ck.Value != stateQ {
		respond.JSONError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid oauth state")
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.CookieSecure,
	})
	ctx := c.Request.Context()
	tok, err := h.auth.ExchangeGitHubCode(ctx, code)
	if err != nil {
		respond.JSONError(c, http.StatusBadGateway, "OAUTH_ERROR", "token exchange failed")
		return
	}
	if !tok.Valid() || tok.AccessToken == "" {
		respond.JSONError(c, http.StatusBadGateway, "OAUTH_ERROR", "invalid token response")
		return
	}
	gh, email, err := h.auth.FetchGitHubProfile(ctx, tok.AccessToken)
	if err != nil {
		respond.JSONError(c, http.StatusBadGateway, "GITHUB_ERROR", "could not load github profile")
		return
	}
	user, err := h.auth.UpsertGitHubUser(ctx, tok.AccessToken, gh, email)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not save user")
		return
	}
	jwtStr, err := h.auth.IssueAccessToken(user)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not issue session")
		return
	}
	h.setAccessTokenCookie(c.Writer, jwtStr)
	redir := h.cfg.FrontendURL + "/login"
	c.Redirect(http.StatusFound, redir)
}

type meResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserIDKey)
	ctx := c.Request.Context()
	u, err := h.auth.Me(ctx, userID)
	if err != nil {
		respond.JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "could not load user")
		return
	}
	if u == nil {
		respond.JSONError(c, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	respond.JSONOK(c, meResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAccessTokenCookie(c.Writer)
	respond.JSONOK(c, gin.H{"ok": true})
}

func (h *AuthHandler) setAccessTokenCookie(w http.ResponseWriter, value string) {
	ck := &http.Cookie{
		Name:     middleware.AccessTokenCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   int(h.cfg.JWTTTL().Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	if d := h.cfg.SessionCookieDomain; d != "" {
		ck.Domain = d
	}
	http.SetCookie(w, ck)
}

func (h *AuthHandler) clearAccessTokenCookie(w http.ResponseWriter) {
	ck := &http.Cookie{
		Name:     middleware.AccessTokenCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	if d := h.cfg.SessionCookieDomain; d != "" {
		ck.Domain = d
	}
	http.SetCookie(w, ck)
}
