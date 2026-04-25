package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTracingMiddlewareCreatesSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	r := chi.NewRouter()
	r.Use(TracingMiddleware("test-service"))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		span := SpanFromContext(r.Context())
		if !span.SpanContext().IsValid() {
			t.Error("expected valid span in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTracingMiddlewareRecordsStatus500(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	r := chi.NewRouter()
	r.Use(TracingMiddleware("test-service"))
	r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestTracingMiddlewarePropagatesContext(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	var receivedCtx context.Context
	r := chi.NewRouter()
	r.Use(TracingMiddleware("test-service"))
	r.Get("/ctx", func(w http.ResponseWriter, r *http.Request) {
		receivedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ctx", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if receivedCtx == nil {
		t.Fatal("expected context to be captured")
	}
	traceID := TraceIDFromContext(receivedCtx)
	if traceID == "" {
		t.Error("expected non-empty trace ID in propagated context")
	}
}

func TestTraceIDFromEmptyContext(t *testing.T) {
	traceID := TraceIDFromContext(context.Background())
	if traceID != "" {
		t.Error("expected empty trace ID without tracer")
	}
}

func TestSpanIDFromEmptyContext(t *testing.T) {
	spanID := SpanIDFromContext(context.Background())
	if spanID != "" {
		t.Error("expected empty span ID without tracer")
	}
}

func TestInitTracerNoEndpoint(t *testing.T) {
	shutdown, err := InitTracer(TracingConfig{
		ServiceName: "test",
		Endpoint:    "",
		Insecure:    true,
	})
	if err != nil {
		t.Fatalf("InitTracer failed: %v", err)
	}
	if shutdown == nil {
		t.Error("expected non-nil shutdown function")
	}
	_ = shutdown(context.Background())
}

func TestInitTracerWithSampleRate(t *testing.T) {
	shutdown, err := InitTracer(TracingConfig{
		ServiceName: "test",
		Endpoint:    "",
		Insecure:    true,
		SampleRate:  0.5,
	})
	if err != nil {
		t.Fatalf("InitTracer with sample rate failed: %v", err)
	}
	_ = shutdown(context.Background())
}

func TestMeasureDBCall(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := context.Background()
	err := MeasureDBCall(ctx, "get_device", func(innerCtx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestMeasureDBCallWithError(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := context.Background()
	expectedErr := errors.New("db timeout")
	err := MeasureDBCall(ctx, "put_device", func(innerCtx context.Context) error {
		return expectedErr
	})
	if err != expectedErr {
		t.Fatalf("expected db timeout error, got %v", err)
	}
}

func TestStartSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test_operation")
	defer span.End()

	if !span.SpanContext().IsValid() {
		t.Error("expected valid span")
	}
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		t.Error("expected non-empty trace ID")
	}
}

func TestSpanFromContextReturnsValidSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx, span := StartSpan(context.Background(), "check")
	defer span.End()

	retrieved := SpanFromContext(ctx)
	if !retrieved.SpanContext().IsValid() {
		t.Error("retrieved span should be valid")
	}
}
