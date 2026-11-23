package logger

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

// GinLogger returns a gin.HandlerFunc middleware that emits a structured log
// record for each completed HTTP request with OpenTelemetry trace correlation.
func GinLogger(l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		status := c.Writer.Status()

		attrs := []any{
			slog.Group("http",
				slog.String("method", c.Request.Method),
				slog.String("route", route),
				slog.String("target", c.Request.URL.RequestURI()),
				slog.Int("status_code", status),
				slog.Float64("duration_ms", float64(duration.Microseconds())/1000.0),
				slog.String("client_ip", c.ClientIP()),
				slog.String("user_agent", c.Request.UserAgent()),
				slog.Int("response_size", c.Writer.Size()),
			),
		}

		if userID, ok := authcontext.GetUserID(c); ok && userID != "" {
			var userAttrs []any
			userAttrs = append(userAttrs, slog.String("id", userID))
			if role, ok := authcontext.GetUserRole(c); ok && role != "" {
				userAttrs = append(userAttrs, slog.String("role", role))
			}
			attrs = append(attrs, slog.Group("user", userAttrs...))
		}

		ctx := c.Request.Context()

		// Health checks log at DEBUG when successful to reduce noise
		if route == "/api/healthz" && status < http.StatusBadRequest {
			l.DebugContext(ctx, "http request completed", attrs...)
			return
		}

		switch {
		case status >= http.StatusInternalServerError:
			l.ErrorContext(ctx, "http request failed with server error", attrs...)
		case status >= http.StatusBadRequest:
			l.WarnContext(ctx, "http request completed with client error", attrs...)
		default:
			l.InfoContext(ctx, "http request completed", attrs...)
		}
	}
}

// GinRecovery returns a gin.HandlerFunc middleware that recovers from panics,
// logs the stack trace in a structured format with trace context, and sends a 500 error.
func GinRecovery(l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				ctx := c.Request.Context()

				route := c.FullPath()
				if route == "" {
					route = c.Request.URL.Path
				}

				l.ErrorContext(
					ctx,
					"panic recovered in http handler",
					slog.Any("panic", r),
					slog.String("error.stack", stack),
					slog.String("http.route", route),
					slog.String("http.method", c.Request.Method),
					slog.String("http.client_ip", c.ClientIP()),
				)

				response.Error(c, "internal server error", http.StatusInternalServerError, nil)
			}
		}()
		c.Next()
	}
}
