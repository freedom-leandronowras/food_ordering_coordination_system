package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	tracer   = otel.Tracer("hello-handler")
	once     sync.Once
	tp       *trace.TracerProvider
)

func initOTel() {
	once.Do(func() {
		ctx := context.Background()

		// OTLP HTTP Exporter
		// Endpoint and headers will be picked up from environment variables:
		// OTEL_EXPORTER_OTLP_ENDPOINT (e.g., https://ingest.dash0.com/v1)
		// OTEL_EXPORTER_OTLP_HEADERS (e.g., Authorization=Bearer <TOKEN>)
		exporter, err := otlptracehttp.New(ctx)
		if err != nil {
			fmt.Printf("failed to create exporter: %v\n", err)
			return
		}

		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceNameKey.String("food-ordering-api"),
			),
		)
		if err != nil {
			fmt.Printf("failed to create resource: %v\n", err)
			return
		}

		tp = trace.NewTracerProvider(
			trace.WithBatchProcessor(trace.NewBatchSpanProcessor(exporter)),
			trace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
	})
}

// Handler is the entry point for the Vercel serverless function.
func Handler(
	w http.ResponseWriter,
	r *http.Request,
) {
	initOTel()

	// Wrap the handler logic with OTel
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "hello-request")
		defer span.End()

		fmt.Fprintf(
			w,
			"Hello from Go on Vercel! (Instrumented with OTel)",
		)
	})

	otelHandler := otelhttp.NewHandler(handler, "hello-api")
	otelHandler.ServeHTTP(w, r)
}
