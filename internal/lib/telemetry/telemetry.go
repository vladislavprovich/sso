package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/vladislavprovich/sso/internal/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	timeOutSeconds = 10
	metricsTimeout = 5
)

// InitLogger initializes the logger for Loki.
func InitLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// InitMetrics initializes Prometheus metrics server.
func InitMetrics(ctx context.Context, log *slog.Logger, cfg *config.Config) (*prometheus.Registry, error) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	server := &http.Server{
		Addr:              cfg.Otel.Adr,
		ReadTimeout:       metricsTimeout * time.Second,
		WriteTimeout:      metricsTimeout * time.Second,
		ReadHeaderTimeout: metricsTimeout * time.Second,
		Handler:           promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	}

	go func() {
		log.InfoContext(ctx, "Prometheus metrics available", slog.Int("port", cfg.Otel.MetricsPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.ErrorContext(ctx, "Error starting metrics server", slog.String("error", err.Error()))
		}
	}()

	return registry, nil
}

// InitTracing initializes tracing using Tempo (OTLP).
func InitTracing(ctx context.Context, otelEndpoint string, log *slog.Logger) (*trace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(ctx, timeOutSeconds*time.Second)
	defer cancel()

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelEndpoint),
	)

	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.ErrorContext(ctx, "Error creating trace exporter", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	resource, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("sso-service"),
		),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error creating resource", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(resource),
		trace.WithSampler(trace.AlwaysSample()),
	)

	return tracerProvider, nil
}
