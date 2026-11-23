package logger

import (
	"context"
	"log/slog"
	"strings"
)

const RedactedPlaceholder = "[REDACTED]"

var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"auth":          {},
	"authorization": {},
	"api_key":       {},
	"apikey":        {},
	"cookie":        {},
	"set-cookie":    {},
}

// MaskingHandler scrubs sensitive keys from log attributes before they are formatted.
type MaskingHandler struct {
	handler slog.Handler
}

// NewMaskingHandler creates a handler that redacts sensitive attributes.
func NewMaskingHandler(handler slog.Handler) *MaskingHandler {
	return &MaskingHandler{
		handler: handler,
	}
}

var _ slog.Handler = (*MaskingHandler)(nil)

func (h *MaskingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *MaskingHandler) Handle(ctx context.Context, r slog.Record) error {
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	r.Attrs(func(attr slog.Attr) bool {
		newRecord.AddAttrs(maskAttr(attr))
		return true
	})

	return h.handler.Handle(ctx, newRecord)
}

func (h *MaskingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	masked := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		masked[i] = maskAttr(a)
	}
	return &MaskingHandler{
		handler: h.handler.WithAttrs(masked),
	}
}

func (h *MaskingHandler) WithGroup(name string) slog.Handler {
	return &MaskingHandler{
		handler: h.handler.WithGroup(name),
	}
}

func maskAttr(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()

	if isSensitiveKey(attr.Key) {
		return slog.String(attr.Key, RedactedPlaceholder)
	}

	if attr.Value.Kind() == slog.KindGroup {
		attrs := attr.Value.Group()
		maskedGroup := make([]slog.Attr, len(attrs))
		for i, a := range attrs {
			maskedGroup[i] = maskAttr(a)
		}
		return slog.Group(attr.Key, anySliceToAny(maskedGroup)...)
	}

	return attr
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	if _, ok := sensitiveKeys[lower]; ok {
		return true
	}
	// Check compound keys such as "user.password" or "req.headers.authorization"
	for sk := range sensitiveKeys {
		if strings.HasSuffix(lower, "."+sk) || strings.HasSuffix(lower, "_"+sk) || strings.HasSuffix(lower, "-"+sk) {
			return true
		}
	}
	return false
}

func anySliceToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, a := range attrs {
		result[i] = a
	}
	return result
}
