package middleware

import (
	"net/http"

	"github.com/branchly/branchly-api/internal/respond"
	"github.com/gin-gonic/gin"
)

const InternalSecretHeader = "X-Internal-Secret"

func InternalAPI(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(InternalSecretHeader) != secret {
			respond.JSONError(c, http.StatusForbidden, "FORBIDDEN", "invalid internal secret")
			c.Abort()
			return
		}
		c.Next()
	}
}
