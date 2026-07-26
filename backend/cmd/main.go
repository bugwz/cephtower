package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cephtower/backend/internal/app"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/logging"
)

const shutdownTimeout = 15 * time.Second

func main() {
	configPath := flag.String("config", config.DefaultPath, "Path to the YAML configuration file")
	flag.Parse()

	application, err := app.New(*configPath)
	if err != nil {
		logging.Errorf("application initialization failed: config_path=%q error=%v", *configPath, err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := application.Close(closeCtx); err != nil {
			logging.Errorf("shutdown cleanup failed: error=%v", err)
		}
	}()

	if err := application.Run(ctx); err != nil {
		logging.Errorf("HTTP server stopped unexpectedly: error=%v", err)
		os.Exit(1)
	}
}
