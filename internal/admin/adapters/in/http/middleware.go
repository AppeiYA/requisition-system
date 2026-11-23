package adminhttp

import (
	"strings"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestMetadata attaches a request id, the client IP and the user agent to
// the request context. The audit recorder reads them from there, which keeps
// the use cases free of HTTP concerns and lets every admin route inherit the
// metadata without repeating the middleware.
func RequestMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
		// Reject oversized or blank client values instead of storing them:
		// the column is 64 characters and the id ends up in logs.
		if requestID == "" || len(requestID) > 64 {
			requestID = uuid.NewString()
		}
		c.Header(requestIDHeader, requestID)

		if c.Request != nil {
			c.Request = c.Request.WithContext(audit.WithRequestMeta(
				c.Request.Context(),
				audit.RequestMeta{
					RequestID: requestID,
					IPAddress: c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
				},
			))
		}

		c.Next()
	}
}
