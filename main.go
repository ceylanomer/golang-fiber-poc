package main

import (
	"fmt"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	recover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang-fiber-poc/app/client"
	"golang-fiber-poc/app/healthcheck"
	"golang-fiber-poc/app/product"
	"golang-fiber-poc/infra/couchbase"
	"golang-fiber-poc/pkg/config"
	"golang-fiber-poc/pkg/handler"
	_ "golang-fiber-poc/pkg/log"
	"golang-fiber-poc/pkg/tracer"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests",
	Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
}, []string{"route", "method", "status"})

func init() {
	prometheus.MustRegister(httpRequestDuration)
}

func RequestDurationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		httpRequestDuration.WithLabelValues(
			c.Route().Path,
			c.Method(),
			status,
		).Observe(duration)
		return err
	}
}

func main() {
	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("Starting server...")

	httpClient := client.NewHttpClient()

	tp := tracer.InitTracer()
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
	app.Use(RequestDurationMiddleware())

	app.Get("/healthcheck", handler.Handle[healthcheck.Request, healthcheck.Response](healthcheckHandler))
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

	productGroup.Get("/:id", handler.Handle[product.GetProductRequest, product.GetProductResponse](getProductHandler))
	productGroup.Post("/", handler.Handle[product.CreateProductRequest, product.CreateProductResponse](createProductHandler))
	productGroup.Put("/:id", handler.Handle[product.UpdateProductRequest, product.UpdateProductResponse](updateProductHandler))

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			zap.L().Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("Server started on port", zap.String("port", appConfig.Port))

	gracefulShutdown(app)
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
