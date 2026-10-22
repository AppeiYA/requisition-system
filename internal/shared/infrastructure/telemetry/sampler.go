package telemetry

import (
	"errors"
	"fmt"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// The sampler names follow the OpenTelemetry specification, so the standard
// OTEL_TRACES_SAMPLER variable selects one of them.
const (
	SamplerAlwaysOn                = "always_on"
	SamplerAlwaysOff               = "always_off"
	SamplerParentBasedAlwaysOn     = "parentbased_always_on"
	SamplerParentBasedTraceIDRatio = "parentbased_traceidratio"
	SamplerTraceIDRatio            = "traceidratio"
)

// ErrSamplerUnsupported is returned when the configured sampler is not one of
// the supported ones.
var ErrSamplerUnsupported = errors.New("telemetry: unsupported sampler")

// newSampler builds the trace sampler from its name and its ratio.
//
// The parent based samplers keep the decision of the upstream service: a
// trace that was sampled upstream stays sampled here, so a trace is never cut
// in half between the services it crosses.
func newSampler(name string, ratio float64) (sdktrace.Sampler, error) {
	switch name {
	case SamplerAlwaysOn:
		return sdktrace.AlwaysSample(), nil
	case SamplerAlwaysOff:
		return sdktrace.NeverSample(), nil
	case SamplerParentBasedAlwaysOn:
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case SamplerParentBasedTraceIDRatio:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)), nil
	case SamplerTraceIDRatio:
		return sdktrace.TraceIDRatioBased(ratio), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrSamplerUnsupported, name)
	}
}
