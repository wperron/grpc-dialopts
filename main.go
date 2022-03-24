package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancelNotify := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancelNotify()

	timeout, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTimeout()

	exp, err := otlptracegrpc.New(timeout,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		panic(err)
	}

	rs, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String("grpc-dialopts")),
	)
	if err != nil {
		panic(err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(rs),
	)
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("grpc-dialopts")
	func() {
		_, span := tracer.Start(ctx, "spanny")
		fmt.Println("Hello")
		defer span.End()
	}()

	<-ctx.Done()
}
