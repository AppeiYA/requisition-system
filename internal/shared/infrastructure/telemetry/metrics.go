package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("requisition-system")

type Counter struct {
	instrument metric.Int64Counter
}

type Histogram struct {
	instrument metric.Float64Histogram
}

func NewHistogram(name, description string) (*Histogram, error) {
	histogram, err := meter.Float64Histogram(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		return nil, err
	}

	return &Histogram{
		instrument: histogram,
	}, nil
}

func (h *Histogram) Record(ctx context.Context, seconds float64) {
	h.instrument.Record(ctx, seconds)
}

func NewCounter(name, description string) (*Counter, error) {
	counter, err := meter.Int64Counter(
		name,
		metric.WithDescription(description),
	)
	if err != nil {
		return nil, err
	}

	return &Counter{
		instrument: counter,
	}, nil
}

func (c *Counter) Inc(ctx context.Context) {
	c.instrument.Add(ctx, 1)
}

func (c *Counter) Add(ctx context.Context, value int64) {
	c.instrument.Add(ctx, value)
}

func (c *Counter) AddWithAttributes(
	ctx context.Context,
	value int64,
	attributes ...attribute.KeyValue,
) {
	c.instrument.Add(ctx, value, metric.WithAttributes(attributes...))
}

func (c *Counter) Measure(ctx context.Context, histogram *Histogram, operation func(context.Context) error) error {
	start := time.Now()

	err := operation(ctx)

	histogram.Record(ctx, time.Since(start).Seconds())

	return err
}
