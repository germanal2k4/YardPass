package tracer

import (
	"context"
	"fmt"
	"strings"
	"yardpass/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

type Tracer struct {
	tp   trace.TracerProvider
	Name string
}

func NewTracer(lc fx.Lifecycle, c config.TracerConfig) (*Tracer, error) {
	propagators := []propagation.TextMapPropagator{propagation.TraceContext{}}
	propagators = append(propagators, propagation.Baggage{})

	propagator := propagation.NewCompositeTextMapPropagator(propagators...)

	if !c.Enabled {
		noopTp := noop.NewTracerProvider()
		otel.SetTracerProvider(noopTp)
		otel.SetTextMapPropagator(propagator)

		return &Tracer{tp: noopTp}, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("yardpass"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(c.Url),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	spanProc := sdktrace.NewBatchSpanProcessor(exporter)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(spanProc),
	)

	otel.SetTracerProvider(tp)

	tracer := &Tracer{
		tp:   tp,
		Name: "yardpass",
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	})

	return tracer, nil
}

func (t *Tracer) StartSpan(ctx context.Context, span string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tp.Tracer(t.Name).Start(ctx, span, opts...)
}

func TraceSafeString(s string) string {
	return strings.ToValidUTF8(s, "?")
}
