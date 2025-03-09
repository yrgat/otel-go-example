package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(":4318"),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("service-b"),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp, nil
}

func processData(ctx context.Context) {
	tr := otel.Tracer("b-service")
	_, span := tr.Start(ctx, "processData")
	defer span.End()

	time.Sleep(500 * time.Millisecond) // Simülasyon için gecikme
	log.Println("Veri işlendi")
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Tracer başlatılamadı: %v", err)
	}
	defer tp.Shutdown(context.Background())

	app := fiber.New()

	// OpenTelemetry middleware ekleyelim
	app.Use(otelfiber.Middleware())

	app.Get("/process", func(c *fiber.Ctx) error {
		tr := otel.Tracer("b-service")
		ctx, span := tr.Start(c.UserContext(), "handleRequest")
		defer span.End()

		processData(ctx)

		return c.JSON(fiber.Map{"message": "Processed successfully"})
	})

	log.Println("B Servisi 4000 portunda çalışıyor...")
	log.Fatal(app.Listen(":4000"))
}
