package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"cephtower/backend/internal/api"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integrations/ceph/dashboard"
	"cephtower/backend/internal/logging"
	"cephtower/backend/internal/store"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "Path to the YAML configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	if _, err := logging.Install(cfg.Logging); err != nil {
		slog.Error("configure logging", "error", err)
		os.Exit(1)
	}

	db, err := store.Open(cfg.Database)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}

	cephClient := dashboard.NewDashboardClient(cfg.Ceph)
	server := api.NewServer(cfg, cephClient, db)
	defer func() {
		if err := server.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}()

	slog.Info("cephtower database configured", "engine", cfg.Database.Engine)
	slog.Info("cephtower backend listening", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, server.Routes()); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
