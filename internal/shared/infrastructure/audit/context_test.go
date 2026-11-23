package audit

import (
	"context"
	"testing"
)

func TestRequestMetaRoundTrip(t *testing.T) {
	want := RequestMeta{
		RequestID: "req-123",
		IPAddress: "203.0.113.7",
		UserAgent: "test-agent",
	}

	got, ok := RequestMetaFromContext(WithRequestMeta(context.Background(), want))
	if !ok {
		t.Fatal("expected metadata in context")
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestRequestMetaMissing(t *testing.T) {
	if _, ok := RequestMetaFromContext(context.Background()); ok {
		t.Fatal("expected no metadata in a bare context")
	}
}
