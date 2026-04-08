// Package infra provides infrastructure adapters including distributed tracing.
//
// telemetry.go implements OpenTelemetry-compatible tracing with:
//   - Configurable TracerProvider (OTLP exporter or noop fallback)
//   - Gin middleware for HTTP request span creation
//   - Trace ID extraction for structured log correlation
//   - JSON structured logging via slog with trace context
//
// Requirements: 44.1, 44.2, 44.3, 44.4
package infra

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "ezistock-backend"

// InitTracer initialises the OpenTelemetry TracerProvider.
//
// When OTEL_EXPORTER_OTLP_ENDPOINT is set, traces are exported via gRPC to
// that endpoint (Jaeger, AWS X-Ray collector, or any OTLP-compatible backend).
// Otherwise a noop provider is used so the application runs without overhead.
//
// The returned shutdown function must be called on application exit to flush
// pending spans.
func InitTracer(ctx context.Context, serviceName string) (shutdown func(context.Context) error, err error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		// Noop — no exporter configured.
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// TracingMiddleware returns Gin middleware that creates a span for each HTTP
// request and propagates the trace context from incoming headers.
//
// The middleware:
//   - Extracts W3C Trace Context from request headers
//   - Creates a server span named "HTTP <method> <path>"
//   - Records status code, method, path, and client IP as span attributes
//   - Sets span status to Error for 5xx responses
//   - Injects trace ID into Gin context for downstream log correlation
func TracingMiddleware() gin.HandlerFunc {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		// Extract trace context from incoming request headers.
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		spanName := fmt.Sprintf("HTTP %s %s", c.Request.Method, c.FullPath())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Request.Method),
				semconv.URLPath(c.Request.URL.Path),
				semconv.ClientAddress(c.ClientIP()),
			),
		)
		defer span.End()

		// Replace request context so downstream handlers see the span.
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))
		if status >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
		}
	}
}

// TraceIDFromContext extracts the trace ID string from a context.
// Returns an empty string when no active span exists.
// Use this to correlate structured log entries with distributed traces.
func TraceIDFromContext(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		return ""
	}
	return sc.TraceID().String()
}

// SpanIDFromContext extracts the span ID string from a context.
func SpanIDFromContext(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasSpanID() {
		return ""
	}
	return sc.SpanID().String()
}

// LoggerWithTrace returns an slog.Logger enriched with trace_id and span_id
// fields extracted from the context. When no active span exists the base
// logger is returned unchanged.
func LoggerWithTrace(ctx context.Context, base *slog.Logger) *slog.Logger {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return base
	}
	return base.With(
		slog.String("trace_id", traceID),
		slog.String("span_id", SpanIDFromContext(ctx)),
	)
}

// StartSpan is a convenience wrapper that starts a new child span with the
// package-level tracer. Callers must call span.End() when done.
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name, trace.WithAttributes(attrs...))
}

// SetupJSONLogging configures the default slog logger to output JSON to
// stdout, suitable for structured log aggregation (ELK, CloudWatch, etc.).
func SetupJSONLogging(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}
