package main

import (
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

	"cephtower/backend/internal/api"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/logging"
	"cephtower/backend/internal/scheduler"
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

	_, closeLog, err := logging.Install(cfg.Logging, cfg.Server.WorkDir)
	if err != nil {
		slog.Error("configure logging", "error", err)
		os.Exit(1)
	}
	taskScheduler, err := scheduler.Start()
	if err != nil {
		_ = closeLog()
		slog.Error("configure scheduled tasks", "error", err)
		os.Exit(1)
	}
	defer taskScheduler.Stop()
	defer func() {
		if err := closeLog(); err != nil {
			slog.Error("close log file", "error", err)
		}
	}()

	db, err := store.Open(cfg.Database, cfg.Server.WorkDir)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}

	server := api.NewServer(cfg, nil, db)
	defer func() {
		if err := server.Close(); err != nil {
			slog.Error("close database", "error", err)
		}
	}()

	slog.Info("cephtower database configured", "engine", cfg.Database.Engine)
	listenAddr := net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port))
	slog.Info("cephtower backend listening", "addr", listenAddr)
	if err := http.ListenAndServe(listenAddr, server.Routes()); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
