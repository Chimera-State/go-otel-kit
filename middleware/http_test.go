package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func setupTestProvider() *tracetest.SpanRecorder {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return sr
}

func TestTraceMiddleware_NewTrace(t *testing.T) {
	sr := setupTestProvider()

	handler := TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name() != "GET /test" {
		t.Errorf("expected span name 'GET /test', got %s", span.Name())
	}
	// Yeni trace olduğu için parent olmamalı
	if span.Parent().IsValid() {
		t.Errorf("expected no parent span, but got one")
	}
}

func TestTraceMiddleware_WithParent(t *testing.T) {
	sr := setupTestProvider()

	handler := TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test-parent", nil)

	traceID, _ := trace.TraceIDFromHex("0af7651916cd43dd8448eb211c80319c")
	spanID, _ := trace.SpanIDFromHex("b7ad6b7169203331")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Parent().SpanID() != spanID {
		t.Errorf("expected parent span ID %s, got %s", spanID, span.Parent().SpanID())
	}
	if span.SpanContext().TraceID() != traceID {
		t.Errorf("expected trace ID %s, got %s", traceID, span.SpanContext().TraceID())
	}
}

func TestTraceMiddleware_ErrorStatus(t *testing.T) {
	sr := setupTestProvider()

	handler := TraceMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // 500 hatası
	}))

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Status().Code != codes.Error {
		t.Errorf("expected error status code, got %v", span.Status().Code)
	}
}
