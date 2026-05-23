package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders = "Authorization, Content-Type"
	corsMaxAge       = "3600"
)

// CORS returns a middleware that allows cross-origin requests from any origin in allowedOrigins.
// "*" in the list disables origin checking and echoes any origin back (useful only for development).
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAny := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			allowAny = true
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			_, originAllowed := allowed[origin]
			if allowAny || originAllowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Methods", corsAllowMethods)
				c.Header("Access-Control-Allow-Headers", corsAllowHeaders)
				c.Header("Access-Control-Max-Age", corsMaxAge)
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
