package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cephtower/backend/internal/api"
	v1handler "cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	"cephtower/backend/internal/logging"
	authservice "cephtower/backend/internal/service/auth"
	cephproxyservice "cephtower/backend/internal/service/cephproxy"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/service/collector"
	settingsservice "cephtower/backend/internal/service/settings"
	setupservice "cephtower/backend/internal/service/setup"
	"cephtower/backend/internal/store"
	"cephtower/backend/internal/task"
)

const (
	shutdownTimeout   = 15 * time.Second
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 2 * time.Minute
)

// App owns the process-level dependencies and their lifecycle.
type App struct {
	config     config.Config
	logger     *slog.Logger
	tasks      *task.Manager
	apiServer  *api.Server
	database   *store.Manager
	httpServer *http.Server
	closeLog   func() error

	closeOnce sync.Once
	closeErr  error
}

// New builds the application without starting its HTTP listener.
func New(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger, logCleanup, closeLog, err := logging.InstallManaged(cfg.Logging, cfg.Server.Dir)
	if err != nil {
		return nil, fmt.Errorf("configure logging: %w", err)
	}

	taskManager := task.NewManager(8)

	db, err := store.Open(cfg.Database, cfg.Server.Dir)
	if err != nil {
		taskManager.Stop()
		_ = closeLog()
		return nil, fmt.Errorf("open database: %w", err)
	}

	runtimeDir := config.ResolveRuntimeDir(cfg)
	if err := collector.SyncCephRuntimeFiles(context.Background(), db, runtimeDir); err != nil {
		_ = store.Close(db)
		taskManager.Stop()
		_ = closeLog()
		return nil, fmt.Errorf("sync ceph runtime files: %w", err)
	}

	databaseManager := store.NewManager(db)
	var configMu sync.RWMutex
	currentConfig := func() config.Config { configMu.RLock(); defer configMu.RUnlock(); return cfg }
	updateConfig := func(next config.Config) {
		configMu.Lock()
		cfg = next
		configMu.Unlock()
	}
	submitTask := func(name string, job func(context.Context) error) error {
		return taskManager.Submit(name, task.Job(job))
	}
	dataFetchService := collector.New(databaseManager.Current, runtimeDir)
	cephClient := cephproxyservice.New(databaseManager.Current, submitTask, dataFetchService, runtimeDir)
	authService := authservice.New(databaseManager.Current, currentConfig)
	clusterService := clusterservice.New(clusterservice.Dependencies{
		Database: databaseManager.Current,
		Discover: func(ctx context.Context, database *store.Database, cluster *store.CephCluster) error {
			return collector.DiscoverAndSyncCephClusterWithWorkDir(ctx, database, cluster, runtimeDir)
		},
		CleanRuntime: func(ctx context.Context, clusterID uint) error {
			return collector.DeleteCephClusterRuntimeFiles(runtimeDir, clusterID)
		},
	})
	settingsService := settingsservice.New(cephClient, databaseManager.Current)
	systemSettingsService := settingsservice.NewSystem(databaseManager.Current, currentConfig, settingsservice.Submitter(submitTask), dataFetchService)
	setupService := setupservice.New(databaseManager, currentConfig, updateConfig)
	apiHandler := v1handler.New(cephClient, v1handler.Dependencies{
		ClusterService: clusterService, AuthService: authService, SettingsService: settingsService,
		SystemSettingsService: systemSettingsService, SetupService: setupService,
	})
	apiServer := api.NewServer(apiHandler)
	if err := taskManager.Register(
		"ceph-data-fetch",
		time.Now(),
		task.Every(30*time.Second),
		task.CollectorJob(dataFetchService),
	); err != nil {
		taskManager.Stop()
		_ = databaseManager.Close()
		_ = closeLog()
		return nil, fmt.Errorf("register data fetch task: %w", err)
	}
	if logCleanup != nil {
		cleanupSchedule := task.DailyAt(1)
		if err := taskManager.Register(
			"log-retention-cleanup",
			cleanupSchedule(time.Now()),
			cleanupSchedule,
			task.LogCleanupJob(logCleanup),
		); err != nil {
			taskManager.Stop()
			_ = databaseManager.Close()
			_ = closeLog()
			return nil, fmt.Errorf("register log cleanup task: %w", err)
		}
	}
	listenAddr := net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port))
	return &App{
		config:    cfg,
		logger:    logger,
		tasks:     taskManager,
		apiServer: apiServer,
		database:  databaseManager,
		httpServer: &http.Server{
			Addr:              listenAddr,
			Handler:           apiServer.Routes(),
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		closeLog: closeLog,
	}, nil
}

// Run serves requests until the listener fails or ctx is canceled.
func (a *App) Run(ctx context.Context) error {
	if a == nil || a.httpServer == nil || a.apiServer == nil {
		return fmt.Errorf("app is not initialized")
	}

	a.logger.Info("cephtower database configured", "engine", a.config.Database.Engine)
	a.logger.Info("cephtower backend listening", "addr", a.httpServer.Addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

// Close releases every process-level dependency. It is safe to call more than once.
func (a *App) Close(ctx context.Context) error {
	if a == nil {
		return nil
	}

	a.closeOnce.Do(func() {
		var errs []error
		if a.httpServer != nil {
			if err := a.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errs = append(errs, fmt.Errorf("shut down HTTP server: %w", err))
			}
		}
		if a.tasks != nil {
			a.tasks.Stop()
		}
		if a.database != nil {
			if err := a.database.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close database: %w", err))
			}
		}
		if a.closeLog != nil {
			if err := a.closeLog(); err != nil {
				errs = append(errs, fmt.Errorf("close logging: %w", err))
			}
		}
		a.closeErr = errors.Join(errs...)
	})
	return a.closeErr
}
