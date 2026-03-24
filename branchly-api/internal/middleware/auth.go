package middleware

import (
	"net/http"
	"strings"

	"github.com/branchly/branchly-api/internal/respond"
	"github.com/branchly/branchly-api/internal/service"
	"github.com/gin-gonic/gin"
)

const ContextUserIDKey = "userID"

const AccessTokenCookie = "branchly_access_token"

func AuthJWT(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearerToken(c)
		if raw == "" {
			if ck, err := c.Request.Cookie(AccessTokenCookie); err == nil {
				raw = strings.TrimSpace(ck.Value)
			}
		}
		if raw == "" {
			respond.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid session")
			c.Abort()
			return
		}
		sub, err := auth.ValidateAccessToken(raw)
		if err != nil {
			respond.JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, sub)
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	h := strings.TrimSpace(c.GetHeader("Authorization"))
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
