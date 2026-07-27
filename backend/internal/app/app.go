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
	operationservice "cephtower/backend/internal/service/operation"
	reconcilerservice "cephtower/backend/internal/service/reconciler"
	"cephtower/backend/internal/store"
)

const shutdownTimeout = 15 * time.Second

type App struct {
	config     config.Config
	apiServer  *api.API
	database   *store.Manager
	operations *operationservice.Service
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
	_, _, closeLog, err := logging.InstallManaged(cfg.Logging, cfg.Server.Dir)
	if err != nil {
		return nil, fmt.Errorf("configure logging: %w", err)
	}
	db, err := store.Open(cfg.Database, cfg.Server.Dir)
	if err != nil {
		_ = closeLog()
		return nil, fmt.Errorf("open database: %w", err)
	}
	manager := store.NewManager(db)
	currentConfig := func() config.Config { return cfg }
	auth := authservice.New(manager.Current, currentConfig)
	if err := auth.EnsureRoles(context.Background()); err != nil {
		_ = manager.Close()
		_ = closeLog()
		return nil, fmt.Errorf("initialize roles: %w", err)
	}
	operations := operationservice.New(manager.Current, 8, cfg.Database.EncryptionKey)
	runner := &executor.Runner{}
	native := &cephprovider.NativeProvider{Executor: runner}
	clusters := clusterservice.New(manager.Current, cfg.Database.EncryptionKey, operations, native)
	endpoints := endpointservice.New(manager.Current, cfg.Database.EncryptionKey)
	external := externalservice.New(endpoints, operations, cfg.Database.EncryptionKey)
	mutations := mutationservice.New(clusters, runner, cfg.Database.EncryptionKey)
	operations.SetFallback(mutations.Apply)
	operations.SetPlanChecker(func(ctx context.Context, request operationservice.PlanRequest) ([]string, []string, error) {
		if externalservice.Supports(request.Action) {
			return external.CheckPlan(ctx, request)
		}
		return mutations.CheckPlan(ctx, request)
	})
	reconciler := reconcilerservice.New(manager.Current, clusters, native)
	operations.SetPostSuccess(reconciler.ReconcileAffected)
	if err := operations.Register("cluster.refresh", reconciler.ApplyRefresh); err != nil {
		_ = manager.Close()
		_ = closeLog()
		return nil, fmt.Errorf("register cluster refresh operation: %w", err)
	}
	handler := v1handler.New(v1handler.Dependencies{Auth: auth, Clusters: clusters, Operations: operations, Endpoints: endpoints, External: external, Database: manager.Current})
	apiServer := api.NewAPI(handler)
	server := &http.Server{Addr: net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port)), Handler: apiServer.Routes(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 2 * time.Minute}
	return &App{config: cfg, apiServer: apiServer, database: manager, operations: operations, reconciler: reconciler, httpServer: server, closeLog: closeLog}, nil
}
func (a *App) Run(ctx context.Context) error {
	if err := a.operations.Start(ctx); err != nil {
		return fmt.Errorf("start operation workers: %w", err)
	}
	a.reconciler.Start(ctx)
	logging.Infof("backend listening: addr=%s database_engine=%s", a.httpServer.Addr, a.config.Database.Engine)
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
		if a.operations != nil {
			a.operations.Stop()
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
