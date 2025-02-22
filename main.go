package main

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang-fiber-poc/pkg/config"
	_ "golang-fiber-poc/pkg/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	appConfig := config.Read()
	defer zap.L().Sync()

	zap.L().Info("Starting server...")

	app := fiber.New()

	app.Get("/healthcheck", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

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
