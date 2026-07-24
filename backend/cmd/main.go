package main

import (
	"flag"
	"log/slog"
	"os"

	"cephtower/backend/internal/app"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to the YAML configuration file")
	flag.Parse()

	application, err := app.New(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := application.Close(); err != nil {
			slog.Error("close application", "error", err)
		}
	}()
	if err := application.Run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
