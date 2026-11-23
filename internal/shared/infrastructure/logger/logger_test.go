package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestLogger_JSONFormattingAndLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:          "info",
		Format:         "json",
		ServiceName:    "test-service",
		ServiceVersion: "v1.0.0",
		Environment:    "test",
		Output:         buf,
	})

	// Debug should be filtered out
	l.Debug("debug message")
	if buf.Len() > 0 {
		t.Fatalf("expected debug message to be filtered out, got %s", buf.String())
	}

	// Info should appear
	l.Info("info message", slog.String("key", "value"))
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v, raw: %s", err, buf.String())
	}

	if parsed["msg"] != "info message" {
		t.Errorf("msg = %v, want 'info message'", parsed["msg"])
	}
	if parsed["level"] != "INFO" {
		t.Errorf("level = %v, want 'INFO'", parsed["level"])
	}
	if parsed["key"] != "value" {
		t.Errorf("key = %v, want 'value'", parsed["key"])
	}

	svc, ok := parsed["service"].(map[string]any)
	if !ok {
		t.Fatalf("missing service group: %v", parsed)
	}
	if svc["name"] != "test-service" {
		t.Errorf("service.name = %v, want 'test-service'", svc["name"])
	}
	if svc["version"] != "v1.0.0" {
		t.Errorf("service.version = %v, want 'v1.0.0'", svc["version"])
	}
	if svc["environment"] != "test" {
		t.Errorf("service.environment = %v, want 'test'", svc["environment"])
	}
}

func TestLogger_TraceContextEnrichment(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	tracer := tp.Tracer("test-tracer")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	sc := span.SpanContext()
	wantTraceID := sc.TraceID().String()
	wantSpanID := sc.SpanID().String()

	l.InfoContext(ctx, "span message", slog.String("action", "test"))

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if parsed["trace_id"] != wantTraceID {
		t.Errorf("trace_id = %v, want %q", parsed["trace_id"], wantTraceID)
	}
	if parsed["span_id"] != wantSpanID {
		t.Errorf("span_id = %v, want %q", parsed["span_id"], wantSpanID)
	}
	if parsed["trace_flags"] == "" {
		t.Errorf("expected non-empty trace_flags")
	}
}

func TestLogger_MaskingSensitiveKeys(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	l.Info("login attempt",
		slog.String("username", "alice"),
		slog.String("password", "super-secret-password"),
		slog.String("refresh_token", "eyJhbGciOi..."),
		slog.String("authorization", "Bearer xyz123"),
		slog.Group("credentials",
			slog.String("api_key", "secret-key"),
			slog.String("safe_id", "id-123"),
		),
	)

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if parsed["username"] != "alice" {
		t.Errorf("username = %v, want alice", parsed["username"])
	}
	if parsed["password"] != RedactedPlaceholder {
		t.Errorf("password = %v, want %s", parsed["password"], RedactedPlaceholder)
	}
	if parsed["refresh_token"] != RedactedPlaceholder {
		t.Errorf("refresh_token = %v, want %s", parsed["refresh_token"], RedactedPlaceholder)
	}
	if parsed["authorization"] != RedactedPlaceholder {
		t.Errorf("authorization = %v, want %s", parsed["authorization"], RedactedPlaceholder)
	}

	creds, ok := parsed["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("missing credentials group")
	}
	if creds["api_key"] != RedactedPlaceholder {
		t.Errorf("api_key = %v, want %s", creds["api_key"], RedactedPlaceholder)
	}
	if creds["safe_id"] != "id-123" {
		t.Errorf("safe_id = %v, want 'id-123'", creds["safe_id"])
	}
}

func TestLogger_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Options{
		Level:  "warn",
		Format: "text",
		Output: buf,
	})

	l.Warn("warning event", slog.String("reason", "high load"))

	output := buf.String()
	if !strings.Contains(output, "WARN") && !strings.Contains(output, "level=WARN") {
		t.Errorf("expected level=WARN in output, got: %s", output)
	}
	if !strings.Contains(output, "warning event") {
		t.Errorf("expected message in output, got: %s", output)
	}
	if !strings.Contains(output, "reason=\"high load\"") && !strings.Contains(output, "reason=\"high load\"") && !strings.Contains(output, "reason=") {
		t.Errorf("expected reason in output, got: %s", output)
	}
}
