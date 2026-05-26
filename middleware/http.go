package middleware

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/Chimera-State/go-otel-kit/middleware"

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		tracer := otel.Tracer(tracerName)
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)

		if rw.status >= http.StatusBadRequest {
			msg := fmt.Sprintf("HTTP %d: %s", rw.status, http.StatusText(rw.status))
			err := fmt.Errorf("http request failed with status %d", rw.status)

			span.RecordError(err)
			span.SetStatus(codes.Error, msg)
		} else {
			span.SetStatus(codes.Ok, "success")
		}
	})
}
