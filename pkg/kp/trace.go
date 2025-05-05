package kp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func startTracing(appName, endpoint string) (*trace.TracerProvider, error) {
	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(
			exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Millisecond),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(appName),
			),
		),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

func ginOpenTelemetryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract context from incoming request
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		// Start a new span
		tr := otel.GetTracerProvider().Tracer("gokp-dev")
		ctx, span := tr.Start(ctx, fmt.Sprintf("%s %s", strings.ToUpper(c.Request.Method), c.FullPath()))
		defer span.End()

		// Pass the new context down
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Capture status code
		statusCode := c.Writer.Status()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		// Capture meaningful errors
		for _, e := range c.Errors {
			if e.Err == nil {
				continue
			}

			// Skip benign EOF errors (like empty body on POST)
			if errors.Is(e.Err, io.EOF) {
				continue
			}

			// Log and trace other errors
			span.SetAttributes(
				attribute.Bool("error", true),
				attribute.String("http.error", e.Err.Error()),
				attribute.String("exception.message", e.Err.Error()),
				attribute.String("exception.type", fmt.Sprintf("%T", e.Err)),
			)
		}
	}
}
