package setup

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Option func(*Config)

type Config struct {
	ServiceName      string
	ServiceVersion   string
	ExporterEndpoint string
	SamplingRate     float64
}

func defaultConfig() Config {
	return Config{
		ServiceName:      getEnvOr("OTEL_SERVICE_NAME", "unknown-service"),
		ServiceVersion:   getEnvOr("OTEL_SERVICE_VERSION", "0.0.1"),
		ExporterEndpoint: getEnvOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		SamplingRate:     getEnvFloatOr("OTEL_SAMPLING_RATE", 1.0),
	}
}

var shutdownFunc func(context.Context) error

func Init(ctx context.Context, opts ...Option) error {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
	))
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplingRate))),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdownFunc = tp.Shutdown
	return nil
}

func Shutdown(ctx context.Context) error {
	if shutdownFunc != nil {
		return shutdownFunc(ctx)
	}
	return nil
}

//helper func

func getEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloatOr(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

//builder func

func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

func WithExporterEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.ExporterEndpoint = endpoint
	}
}

func WithSamplingRate(rate float64) Option {
	return func(c *Config) {
		c.SamplingRate = rate
	}
}
