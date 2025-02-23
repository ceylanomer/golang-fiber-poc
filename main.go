package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	recover "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang-fiber-poc/app/healthcheck"
	"golang-fiber-poc/app/product"
	"golang-fiber-poc/infra/couchbase"
	"golang-fiber-poc/pkg/config"
	"golang-fiber-poc/pkg/customvalidator"
	_ "golang-fiber-poc/pkg/log"
	"net"
	"net/http"
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
	return func(c fiber.Ctx) error {
		var req R

		if err := c.Bind().Body(&req); err != nil && !errors.Is(err, fiber.ErrUnprocessableEntity) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.Bind().Query(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.Bind().Header(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.Bind().URI(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		res, err := handler.Handle(ctx, &req)

		if err != nil {
			zap.L().Error("Failed to handle request", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(res)
	}
}

func main() {
	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("Starting server...")

	couchbaseRepository := couchbase.NewRepository()

	healthcheckHandler := healthcheck.NewHealthCheckHandler()
	getProductHandler := product.NewGetProductHandler(couchbaseRepository)
	createProductHandler := product.NewCreateProductHandler(couchbaseRepository)

	app := fiber.New(fiber.Config{
		IdleTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		Concurrency:     256 * 1024,
		StructValidator: &customvalidator.StructValidator{Validation: validator.New()}},
	)

	app.Use(recover.New())

	app.Get("/healthcheck", handle[healthcheck.Request, healthcheck.Response](healthcheckHandler))
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Get("/err", func(c fiber.Ctx) error {
		return fiber.ErrUnprocessableEntity
	})

	//mainRouter := app.Group("/api")
	//v1Group := mainRouter.Group("/v1")
	//
	//productGroup := v1Group.Group("/product", basicauth.New(basicauth.Config{
	//	Users: map[string]string{
	//		"admin": "password",
	//	},
	//}))
	//
	//productGroup.Post("/", handle[product.CreateProductRequest, product.CreateProductResponse](createProductHandler))
	//productGroup.Get("/:id", handle[product.GetProductRequest, product.GetProductResponse](getProductHandler))

	app.Post("/api/v1/product", handle[product.CreateProductRequest, product.CreateProductResponse](createProductHandler))
	app.Get("/api/v1/product/:id", handle[product.GetProductRequest, product.GetProductResponse](getProductHandler))

	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", appConfig.Port)); err != nil {
			zap.L().Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("Server started on port", zap.String("port", appConfig.Port))

	gracefulShutdown(app)
}

func httpc() {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.google.com", nil)
	if err != nil {
		zap.L().Error("Failed to create request to google", zap.Error(err))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		zap.L().Error("Failed to make request to google", zap.Error(err))
	}
	zap.L().Info("Response from google", zap.Int("status_code", resp.StatusCode))
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
