package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cephtower/backend/internal/api"
	v1handler "cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/logging"
	authservice "cephtower/backend/internal/service/auth"
	clusterservice "cephtower/backend/internal/service/cluster"
	endpointservice "cephtower/backend/internal/service/endpoint"
	externalservice "cephtower/backend/internal/service/external"
	mutationservice "cephtower/backend/internal/service/mutation"
	reconcilerservice "cephtower/backend/internal/service/reconciler"
	setupservice "cephtower/backend/internal/service/setup"
	"cephtower/backend/internal/store"
)

const shutdownTimeout = 15 * time.Second

type App struct {
	config     config.Config
	apiServer  *api.API
	database   *store.Manager
	reconciler *reconcilerservice.Service
	httpServer *http.Server
	closeLog   func() error
	closeOnce  sync.Once
	closeErr   error
}

func New(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.Database.EncryptionKey == "" {
		key, err := config.GenerateDatabaseEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("generate database encryption key: %w", err)
		}
		cfg.Database.EncryptionKey = key
	}
	if err := config.EnsureDirectories(cfg); err != nil {
		return nil, err
	}
	_, _, closeLog, err := logging.InstallManaged(cfg.Logging, cfg.Server.Dir)
	if err != nil {
		return nil, fmt.Errorf("configure logging: %w", err)
	}
	var db *store.Database
	if !cfg.Server.Bootstrap {
		db, err = store.OpenExisting(cfg.Database, cfg.Server.Dir)
		if err != nil {
			_ = closeLog()
			return nil, fmt.Errorf("open database: %w", err)
		}
	}
	manager := store.NewManager(db)
	configMu := &sync.RWMutex{}
	currentConfig := func() config.Config {
		configMu.RLock()
		defer configMu.RUnlock()
		return cfg
	}
	updateConfig := func(next config.Config) {
		configMu.Lock()
		defer configMu.Unlock()
		cfg = next
	}
	auth := authservice.New(manager.Current, currentConfig)
	runner := &executor.Runner{}
	native := &cephprovider.NativeProvider{Executor: runner}
	clusters := clusterservice.New(manager.Current, cfg.Database.EncryptionKey, native)
	endpoints := endpointservice.New(manager.Current, cfg.Database.EncryptionKey)
	external := externalservice.New(endpoints, cfg.Database.EncryptionKey)
	mutations := mutationservice.New(clusters, runner)
	reconciler := reconcilerservice.New(manager.Current, clusters, native)
	setup := &setupservice.Service{Manager: manager, CurrentConfig: currentConfig, UpdateConfig: updateConfig, OnInitialized: func() {
		reconciler.Start(context.Background())
	}}
	handler := v1handler.New(v1handler.Dependencies{Auth: auth, Clusters: clusters, Endpoints: endpoints, External: external, Mutations: mutations, Reconciler: reconciler, Setup: setup, Database: manager.Current})
	apiServer := api.NewAPI(handler)
	server := &http.Server{Addr: net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port)), Handler: apiServer.Routes(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 2 * time.Minute}
	return &App{config: cfg, apiServer: apiServer, database: manager, reconciler: reconciler, httpServer: server, closeLog: closeLog}, nil
}
func (a *App) Run(ctx context.Context) error {
	if a.database != nil && a.database.Current() != nil {
		a.reconciler.Start(ctx)
	}
	logging.Infof("backend listening: addr=%s", a.httpServer.Addr)
	errCh := make(chan error, 1)
	go func() { errCh <- a.httpServer.ListenAndServe() }()
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
			return err
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
func (a *App) Close(ctx context.Context) error {
	if a == nil {
		return nil
	}
	a.closeOnce.Do(func() {
		var errs []error
		if a.httpServer != nil {
			if err := a.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errs = append(errs, err)
			}
		}
		if a.reconciler != nil {
			a.reconciler.Stop()
		}
		if a.database != nil {
			if err := a.database.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if a.closeLog != nil {
			if err := a.closeLog(); err != nil {
				errs = append(errs, err)
			}
		}
		a.closeErr = errors.Join(errs...)
	})
	return a.closeErr
}
