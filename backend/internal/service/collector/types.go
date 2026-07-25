package collector

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"cephtower/backend/internal/store"
)

type Result struct {
	Source          string
	RecordsUpserted int
	RecordsDeleted  int
}

type RunFilter struct {
	ClusterID uint
	Module    string
	Limit     int
}

func (service Service) RunModule(ctx context.Context, clusterID uint, module string) (Result, error) {
	module = strings.TrimSpace(module)
	config, found, err := service.configByModule(ctx, module)
	if err != nil {
		return Result{}, err
	}
	if !found {
		return Result{}, store.ErrRecordNotFound
	}
	err = service.RunDataFetchConfig(ctx, clusterID, config)
	run, queryErr := service.database().LatestDataFetchRun(ctx, clusterID, module)
	if queryErr != nil {
		return Result{}, queryErr
	}
	return Result{Source: run.Source, RecordsUpserted: run.RecordsUpserted, RecordsDeleted: run.RecordsDeleted}, err
}

func (service Service) RefreshModule(ctx context.Context, clusterID uint, module string) (Result, error) {
	module = strings.TrimSpace(module)
	key := fmt.Sprintf("%d:%s", clusterID, module)
	if !service.startRun(key) {
		return Result{}, fmt.Errorf("%w: %s", ErrModuleRunning, key)
	}
	defer service.finishRun(key)
	result, err := service.fetchClusterModule(ctx, clusterID, module)
	return Result{Source: result.source, RecordsUpserted: result.recordsUpserted, RecordsDeleted: result.recordsDeleted}, err
}

func (service Service) ListRuns(ctx context.Context, filter RunFilter) ([]store.CephDataFetchRun, error) {
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 50
	}
	clusterID := ""
	if filter.ClusterID != 0 {
		clusterID = strconv.FormatUint(uint64(filter.ClusterID), 10)
	}
	return service.database().ListDataFetchRuns(ctx, store.DataFetchRunFilter{ClusterID: clusterID, Module: strings.TrimSpace(filter.Module), Limit: filter.Limit})
}

func (service Service) configByModule(ctx context.Context, module string) (DataFetchConfig, bool, error) {
	configs, err := dataFetchConfigs(ctx, service.database())
	if err != nil {
		return DataFetchConfig{}, false, err
	}
	for _, config := range configs {
		if config.Module == module {
			return config, true, nil
		}
	}
	return DataFetchConfig{}, false, nil
}
