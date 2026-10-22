package telemetry

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func Test_NewSampler_KnownSamplers(t *testing.T) {
	tests := []struct {
		name   string
		sample string
		ratio  float64
		// wantSampled is the decision the sampler takes for a span that has
		// no parent.
		wantSampled bool
	}{
		{
			name:        "always on",
			sample:      SamplerAlwaysOn,
			wantSampled: true,
		},
		{
			name:        "always off",
			sample:      SamplerAlwaysOff,
			wantSampled: false,
		},
		{
			name:        "parent based always on",
			sample:      SamplerParentBasedAlwaysOn,
			wantSampled: true,
		},
		{
			name:        "parent based trace id ratio with a full ratio",
			sample:      SamplerParentBasedTraceIDRatio,
			ratio:       1,
			wantSampled: true,
		},
		{
			name:        "parent based trace id ratio with no ratio",
			sample:      SamplerParentBasedTraceIDRatio,
			wantSampled: false,
		},
		{
			name:        "trace id ratio with a full ratio",
			sample:      SamplerTraceIDRatio,
			ratio:       1,
			wantSampled: true,
		},
		{
			name:        "trace id ratio with no ratio",
			sample:      SamplerTraceIDRatio,
			wantSampled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler, err := newSampler(tt.sample, tt.ratio)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			decision := sampler.ShouldSample(sdktrace.SamplingParameters{
				TraceID: testTraceID(),
			})

			if got := decision.Decision == sdktrace.RecordAndSample; got != tt.wantSampled {
				t.Errorf("sampled = %v, want %v", got, tt.wantSampled)
			}
		})
	}
}

func Test_NewSampler_UnknownSampler(t *testing.T) {
	if _, err := newSampler("sometimes", 1); err == nil {
		t.Fatal("expected an error for an unknown sampler, got none")
	}
}

// Test_NewSampler_ParentBasedKeepsTheParentDecision checks that a parent
// based sampler honours the decision of the upstream service, so a trace is
// never sampled on one side and dropped on the other.
func Test_NewSampler_ParentBasedKeepsTheParentDecision(t *testing.T) {
	sampler, err := newSampler(SamplerParentBasedTraceIDRatio, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    testTraceID(),
		SpanID:     testSpanID(),
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	decision := sampler.ShouldSample(sdktrace.SamplingParameters{
		TraceID:       parent.TraceID(),
		ParentContext: trace.ContextWithSpanContext(context.Background(), parent),
	})

	if decision.Decision != sdktrace.RecordAndSample {
		t.Error("expected the sampled parent to keep the trace sampled")
	}
}

// testTraceID returns a valid trace id, so the ratio based samplers have a
// value to hash.
func testTraceID() trace.TraceID {
	return trace.TraceID{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
}

// testSpanID returns a valid span id.
func testSpanID() trace.SpanID {
	return trace.SpanID{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	}
}
