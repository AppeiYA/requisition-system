package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
)

// Options configures the structured logger.
type Options struct {
	Level          string
	Format         string
	AddSource      bool
	ServiceName    string
	ServiceVersion string
	Environment    string
	Output         io.Writer
}

// New builds a configured structured *slog.Logger with masking and
// OpenTelemetry context propagation handlers.
func New(opts Options) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(opts.Level) {
	case config.LogLevelDebug:
		level = slog.LevelDebug
	case config.LogLevelWarn:
		level = slog.LevelWarn
	case config.LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	hOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: opts.AddSource,
	}

	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	var baseHandler slog.Handler
	if strings.ToLower(opts.Format) == config.LogFormatJSON {
		baseHandler = slog.NewJSONHandler(out, hOpts)
	} else {
		baseHandler = slog.NewTextHandler(out, hOpts)
	}

	// Chain: ContextHandler (adds trace info) -> MaskingHandler (redacts secrets) -> BaseHandler
	handler := NewContextHandler(NewMaskingHandler(baseHandler))

	l := slog.New(handler)

	// Attach service identity resource attributes if specified
	var serviceAttrs []any
	if opts.ServiceName != "" {
		serviceAttrs = append(serviceAttrs, slog.String("name", opts.ServiceName))
	}
	if opts.ServiceVersion != "" {
		serviceAttrs = append(serviceAttrs, slog.String("version", opts.ServiceVersion))
	}
	if opts.Environment != "" {
		serviceAttrs = append(serviceAttrs, slog.String("environment", opts.Environment))
	}

	if len(serviceAttrs) > 0 {
		l = l.With(slog.Group("service", serviceAttrs...))
	}

	return l
}

// SetDefault sets the global slog default logger.
func SetDefault(l *slog.Logger) {
	slog.SetDefault(l)
}
