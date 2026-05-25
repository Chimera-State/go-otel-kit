package setup

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

func TestInitAndShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Init(ctx,
		WithServiceName("test-service"),
		WithServiceVersion("1.0.0"),
		WithExporterEndpoint("localhost:4317"),
		WithSamplingRate(1.0),
	)

	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	tp := otel.GetTracerProvider()
	if tp == nil {
		t.Fatal("Global TracerProvider is nil after Init()")
	}

	err = Shutdown(ctx)
	if err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}
}
