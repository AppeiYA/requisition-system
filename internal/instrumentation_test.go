package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Test_Otelgin_MiddlewareTracesTheRequest checks that the middleware the app
// installs turns an incoming request into a span, so the handlers run inside
// it and the database and queue calls they make hang under the same trace.
func Test_Otelgin_MiddlewareTracesTheRequest(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	// The middleware reads the global providers when it is built, so they
	// have to be installed first, exactly like in the application.
	otel.SetTracerProvider(provider)
	otel.SetMeterProvider(metricnoop.NewMeterProvider())
	t.Cleanup(func() {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		otel.SetMeterProvider(metricnoop.NewMeterProvider())
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown: %v", err)
		}
	})

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(otelgin.Middleware("requisition-system"))
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("ended spans = %d, want 1", len(spans))
	}

	if got := spans[0].Name(); got != "GET /healthz" {
		t.Errorf("span name = %q, want %q", got, "GET /healthz")
	}

	attrs := attribute.NewSet(spans[0].Attributes()...)

	if route, ok := attrs.Value("http.route"); !ok || route.AsString() != "/healthz" {
		t.Errorf("http.route = %v, want %q", route, "/healthz")
	}
	if code, ok := attrs.Value("http.response.status_code"); !ok || code.AsInt64() != http.StatusOK {
		t.Errorf("http.response.status_code = %v, want %d", code, http.StatusOK)
	}
}

// Test_Otelgin_MiddlewareKeepsTheCallerSpanContext checks the middleware
// hands the span context to the handler, so anything the handler does is
// recorded under the request instead of under a new root trace.
func Test_Otelgin_MiddlewareKeepsTheCallerSpanContext(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown: %v", err)
		}
	})

	gin.SetMode(gin.TestMode)

	var handlerSpanContext trace.SpanContext
	router := gin.New()
	router.Use(otelgin.Middleware("requisition-system"))
	router.GET("/healthz", func(c *gin.Context) {
		handlerSpanContext = trace.SpanFromContext(c.Request.Context()).SpanContext()
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if !handlerSpanContext.IsValid() {
		t.Fatal("expected the handler to run inside the span of the request")
	}
	if spans := recorder.Started(); len(spans) == 0 {
		t.Fatal("expected the middleware to start a span")
	} else if handlerSpanContext.SpanID() != spans[0].SpanContext().SpanID() {
		t.Error("expected the handler to see the span of the middleware")
	}
}
