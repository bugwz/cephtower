package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"cephtower/backend/internal/api"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/logging"
	"cephtower/backend/internal/scheduler"
	"cephtower/backend/internal/service/ceph"
	"cephtower/backend/internal/store"
	"gorm.io/gorm"
)

type App struct {
	Config    config.Config
	Logger    *slog.Logger
	Database  *gorm.DB
	Scheduler *scheduler.Scheduler
	API       *api.Server

	closeLog   func() error
	httpServer *http.Server
}

func New(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger, closeLog, err := logging.Install(cfg.Logging, cfg.Server.Dir)
	if err != nil {
		return nil, fmt.Errorf("configure logging: %w", err)
	}

	taskScheduler, err := scheduler.Start()
	if err != nil {
		_ = closeLog()
		return nil, fmt.Errorf("configure scheduled tasks: %w", err)
	}

	db, err := store.Open(cfg.Database, cfg.Server.Dir)
	if err != nil {
		taskScheduler.Stop()
		_ = closeLog()
		return nil, fmt.Errorf("open database: %w", err)
	}
	runtimeDir := config.ResolveRuntimeDir(cfg)
	if err := ceph.SyncCephRuntimeFiles(context.Background(), db, runtimeDir); err != nil {
		_ = store.Close(db)
		taskScheduler.Stop()
		_ = closeLog()
		return nil, fmt.Errorf("sync ceph runtime files: %w", err)
	}

	return &App{
		Config:    cfg,
		Logger:    logger,
		Database:  db,
		Scheduler: taskScheduler,
		API:       api.NewServer(cfg, nil, db),
		closeLog:  closeLog,
	}, nil
}

func (a *App) Run() error {
	if a == nil || a.API == nil {
		return fmt.Errorf("app is not initialized")
	}
	listenAddr := net.JoinHostPort(a.Config.Server.Address, strconv.Itoa(a.Config.Server.Port))
	a.Logger.Info("cephtower database configured", "engine", a.Config.Database.Engine)
	a.Logger.Info("cephtower backend listening", "addr", listenAddr)
	a.httpServer = &http.Server{Addr: listenAddr, Handler: a.API.Routes()}
	err := a.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	var firstErr error
	if a.httpServer != nil {
		if err := a.httpServer.Close(); err != nil && err != http.ErrServerClosed {
			firstErr = err
		}
	}
	if a.API != nil {
		if err := a.API.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	} else if a.Database != nil {
		if err := store.Close(a.Database); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}
	if a.closeLog != nil {
		if err := a.closeLog(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
