package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cephtower/backend/internal/app"
)

const shutdownTimeout = 15 * time.Second

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to the YAML configuration file")
	flag.Parse()

	application, err := app.New(*configPath)
	if err != nil {
		slog.Error("initialize application", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := application.Close(closeCtx); err != nil {
			slog.Error("close application", "error", err)
		}
	}()

	if err := application.Run(ctx); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
