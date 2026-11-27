package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func meter() metric.Meter {
	return otel.Meter(instrumentationName)
}

type Counter struct {
	instrument metric.Int64Counter
}

type Histogram struct {
	instrument metric.Float64Histogram
}

func NewHistogram(name, description, unit string) (*Histogram, error) {
	opts := []metric.Float64HistogramOption{
		metric.WithDescription(description),
	}
	if unit != "" {
		opts = append(opts, metric.WithUnit(unit))
	}

	histogram, err := meter().Float64Histogram(name, opts...)
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

func NewCounter(name, description, unit string) (*Counter, error) {
	opts := []metric.Int64CounterOption{
		metric.WithDescription(description),
	}
	if unit != "" {
		opts = append(opts, metric.WithUnit(unit))
	}

	counter, err := meter().Int64Counter(name, opts...)
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

type Gauge struct {
	instrument metric.Int64Gauge
}

func NewGauge(name, description, unit string) (*Gauge, error) {
	opts := []metric.Int64GaugeOption{
		metric.WithDescription(description),
	}
	if unit != "" {
		opts = append(opts, metric.WithUnit(unit))
	}

	gauge, err := meter().Int64Gauge(name, opts...)
	if err != nil {
		return nil, err
	}

	return &Gauge{instrument: gauge}, nil
}

func (g *Gauge) Record(ctx context.Context, value int64) {
	g.instrument.Record(ctx, value)
}
