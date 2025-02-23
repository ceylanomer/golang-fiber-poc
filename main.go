package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	recover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"golang-fiber-poc/app/client"
	"golang-fiber-poc/app/healthcheck"
	"golang-fiber-poc/app/product"
	"golang-fiber-poc/infra/couchbase"
	"golang-fiber-poc/pkg/config"
	_ "golang-fiber-poc/pkg/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Request any
type Response any

type HandlerInterface[R Request, Res Response] interface {
	Handle(ctx context.Context, req *R) (*Res, error)
}

func handle[R Request, Res Response](handler HandlerInterface[R, Res]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req R

		if err := c.BodyParser(&req); err != nil && !errors.Is(err, fiber.ErrUnprocessableEntity) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.ParamsParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.QueryParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.ReqHeaderParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		/*
			ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
			defer cancel()
		*/

		ctx := c.UserContext()

		res, err := handler.Handle(ctx, &req)
		if err != nil {
			zap.L().Error("Failed to handle request", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(res)
	}
}

//For V3
//func handle[R Request, Res Response](handler HandlerInterface[R, Res]) fiber.Handler {
//	return func(c fiber.Ctx) error {
//		var req R
//
//		if err := c.Bind().Body(&req); err != nil && !errors.Is(err, fiber.ErrUnprocessableEntity) {
//			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
//		}
//
//		if err := c.Bind().Query(&req); err != nil {
//			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
//		}
//
//		if err := c.Bind().Header(&req); err != nil {
//			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
//		}
//
//		if err := c.Bind().URI(&req); err != nil {
//			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
//		}
//
//		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
//		defer cancel()
//
//		res, err := handler.Handle(ctx, &req)
//
//		if err != nil {
//			zap.L().Error("Failed to handle request", zap.Error(err))
//			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
//		}
//
//		return c.JSON(res)
//	}
//}

func main() {
	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("Starting server...")

	httpClient := client.NewHttpClient()

	tp := initTracer()
	couchbaseRepository := couchbase.NewRepository(tp)

	healthcheckHandler := healthcheck.NewHealthCheckHandler()
	getProductHandler := product.NewGetProductHandler(couchbaseRepository, httpClient)
	createProductHandler := product.NewCreateProductHandler(couchbaseRepository)
	updateProductHandler := product.NewUpdateProductHandler(couchbaseRepository)

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Concurrency:  256 * 1024,
	})

	app.Use(recover.New())
	app.Use(otelfiber.Middleware())

	app.Get("/healthcheck", handle[healthcheck.Request, healthcheck.Response](healthcheckHandler))
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Get("/err", func(c *fiber.Ctx) error {
		return fiber.ErrUnprocessableEntity
	})

	mainRouter := app.Group("/api")
	v1Group := mainRouter.Group("/v1")

	productGroup := v1Group.Group("/product", basicauth.New(basicauth.Config{
		Users: map[string]string{
			"admin": "password",
		},
	}))

	productGroup.Get("/:id", handle[product.GetProductRequest, product.GetProductResponse](getProductHandler))
	productGroup.Post("/", handle[product.CreateProductRequest, product.CreateProductResponse](createProductHandler))
	productGroup.Put("/:id", handle[product.UpdateProductRequest, product.UpdateProductResponse](updateProductHandler))

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			zap.L().Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("Server started on port", zap.String("port", appConfig.Port))

	gracefulShutdown(app)
}

func initTracer() *sdktrace.TracerProvider {

	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("localhost:4318"),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		zap.L().Fatal("Failed to create stdout exporter", zap.Error(err))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("golang-fiber-poc"),
			)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func gracefulShutdown(app *fiber.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	zap.L().Info("Shutting down server...")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("Failed to shutdown server", zap.Error(err))
	}

	zap.L().Info("Server shutdown successfully")
}
