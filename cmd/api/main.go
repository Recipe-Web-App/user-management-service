package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/app"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/config"
	customLogger "github.com/jsamuelsen/recipe-web-app/user-management-service/internal/logger"
	"github.com/jsamuelsen/recipe-web-app/user-management-service/internal/server"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// Load config
	cfg := config.Load()

	setupLogger()

	// Create dependency container
	container, err := app.NewContainer(app.ContainerConfig{
		Config: cfg,
	})
	if err != nil {
		slog.Error("failed to create application container", "error", err)
		os.Exit(1)
	}

	defer func() {
		err := container.Close()
		if err != nil {
			slog.Error("failed to close container", "error", err)
		}
	}()

	runServerWithContainer(container)
}

func setupLogger() {
	// Initialize structured logger
	var handlers []slog.Handler

	// Console Handler
	if config.Instance.Logging.ConsoleEnabled {
		level := parseLevel(config.Instance.Logging.ConsoleLevel)
		opts := &slog.HandlerOptions{
			Level: level,
		}
		if config.Instance.Logging.Format == "json" {
			handlers = append(handlers, slog.NewJSONHandler(os.Stdout, opts))
		} else {
			handlers = append(handlers, slog.NewTextHandler(os.Stdout, opts))
		}
	}

	// File Handler
	if config.Instance.Logging.FileEnabled && config.Instance.Logging.File != "" {
		level := parseLevel(config.Instance.Logging.FileLevel)
		opts := &slog.HandlerOptions{
			Level: level,
		}
		writer := &lumberjack.Logger{
			Filename:   config.Instance.Logging.File,
			MaxSize:    config.Instance.Logging.MaxSize,
			MaxBackups: config.Instance.Logging.MaxBackups,
			MaxAge:     config.Instance.Logging.MaxAge,
			Compress:   config.Instance.Logging.Compress,
		}
		// File logging always uses JSON for structured data parsing usually, but respecting format config is fine too.
		// Let's stick to JSON for file to be safe/standard, or use the configured format.
		// User requested common format, so we use config.Instance.Logging.Format
		if config.Instance.Logging.Format == "json" {
			handlers = append(handlers, slog.NewJSONHandler(writer, opts))
		} else {
			handlers = append(handlers, slog.NewTextHandler(writer, opts))
		}
	}

	// If no handlers are enabled, discard
	if len(handlers) == 0 {
		handlers = append(handlers, slog.NewTextHandler(io.Discard, nil))
	}

	logger := slog.New(customLogger.NewFanoutHandler(handlers...))
	slog.SetDefault(logger)
}

func runServerWithContainer(container *app.Container) {
	srv := server.NewServerWithContainer(container)

	// Server run context
	done := make(chan bool, 1)

	go func() {
		slog.Info("Server listening", "addr", srv.Addr)

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	// Graceful shutdown
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-wait
	slog.Info("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err := srv.Shutdown(ctx)
	if err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
	close(done)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
