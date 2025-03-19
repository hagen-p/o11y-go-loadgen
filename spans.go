package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func main() {
	filePath := "spans.data"
	collectorEndpoint := "localhost:4317"

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Unmarshal spans
	var resourceSpans tracev1.ResourceSpans
	err = proto.Unmarshal(data, &resourceSpans)
	if err != nil {
		log.Fatalf("Failed to unmarshal spans: %v", err)
	}

	// Create OTLP gRPC exporter
	ctx := context.Background()
	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(collectorEndpoint),
		otlptracegrpc.WithInsecure(), // Change for TLS if needed
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	))
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	defer exporter.Shutdown(ctx)

	// Create a TracerProvider
	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	defer tp.Shutdown(ctx)

	// Send the spans
	if err := exporter.ExportSpans(ctx, resourceSpans.ScopeSpans); err != nil {
		log.Fatalf("Failed to send spans: %v", err)
	}

	fmt.Println("Spans successfully sent to the collector")
}
