package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/service/collector"
	"cephtower/backend/internal/store"
)

var ErrDataFetchConfigNotFound = errors.New("data fetch module setting not found")

type Submitter func(string, func(context.Context) error) error

type SystemService struct {
	database      func() *store.Database
	currentConfig func() config.Config
	submit        Submitter
	collector     *collector.Service
}

func NewSystem(database func() *store.Database, currentConfig func() config.Config, submit Submitter, workers ...*collector.Service) *SystemService {
	var worker *collector.Service
	if len(workers) > 0 {
		worker = workers[0]
	}
	return &SystemService{database: database, currentConfig: currentConfig, submit: submit, collector: worker}
}

func (s *SystemService) List(ctx context.Context, prefix string) ([]store.Setting, error) {
	db := s.database()
	if err := collector.EnsureDefaultSystemSettings(ctx, db); err != nil {
		return nil, err
	}
	return db.ListSettings(ctx, strings.TrimSpace(prefix))
}

func (s *SystemService) Update(ctx context.Context, key, value string) (store.Setting, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return store.Setting{}, fmt.Errorf("setting key is required")
	}
	if strings.HasPrefix(key, collector.DataFetchSettingPrefix) {
		var item collector.DataFetchConfig
		if json.Unmarshal([]byte(value), &item) != nil {
			return store.Setting{}, fmt.Errorf("data fetch setting value must be valid JSON")
		}
		module := strings.TrimPrefix(key, collector.DataFetchSettingPrefix)
		if item.Module == "" {
			item.Module = module
		}
		if item.Module != module {
			return store.Setting{}, fmt.Errorf("data fetch setting module must match setting key")
		}
		if err := validateDataFetch(item); err != nil {
			return store.Setting{}, err
		}
		collector.NormalizeDataFetchConfig(&item)
		data, err := json.Marshal(item)
		if err != nil {
			return store.Setting{}, err
		}
		value = string(data)
	}
	return s.database().UpsertSetting(ctx, key, value)
}

func (s *SystemService) Reset(ctx context.Context) error {
	return s.database().Transaction(func(tx *store.Database) error {
		if err := tx.DeleteSettingsByPrefix(ctx, collector.DataFetchSettingPrefix); err != nil {
			return err
		}
		return collector.EnsureDefaultSystemSettings(ctx, tx)
	})
}

func (s *SystemService) RunModule(ctx context.Context, module string) error {
	if s.submit == nil {
		return fmt.Errorf("task manager is unavailable")
	}
	item, found, err := s.config(ctx, strings.TrimSpace(module))
	if err != nil {
		return err
	}
	if !found {
		return ErrDataFetchConfigNotFound
	}
	clusters, err := s.database().ListClusters(ctx)
	if err != nil {
		return err
	}
	worker := s.collector
	if worker == nil {
		worker = collector.New(s.database, config.ResolveRuntimeDir(s.currentConfig()))
	}
	return s.submit("manual-data-fetch:"+module, func(jobCtx context.Context) error {
		for _, cluster := range clusters {
			if err := worker.RunConfig(jobCtx, cluster.ID, item); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SystemService) ListRuns(ctx context.Context, clusterID, module string, limit int) ([]store.CephDataFetchRun, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	return s.database().ListDataFetchRuns(ctx, store.DataFetchRunFilter{ClusterID: strings.TrimSpace(clusterID), Module: strings.TrimSpace(module), Limit: limit})
}

func (s *SystemService) config(ctx context.Context, module string) (collector.DataFetchConfig, bool, error) {
	setting, err := s.database().FindSetting(ctx, collector.DataFetchSettingKey(module))
	if errors.Is(err, store.ErrRecordNotFound) {
		return collector.DataFetchConfig{}, false, nil
	}
	if err != nil {
		return collector.DataFetchConfig{}, false, err
	}
	var result collector.DataFetchConfig
	if err := json.Unmarshal([]byte(setting.Value), &result); err != nil {
		return result, false, err
	}
	if result.Module == "" {
		result.Module = module
	}
	collector.NormalizeDataFetchConfig(&result)
	return result, true, nil
}

func validateDataFetch(item collector.DataFetchConfig) error {
	if strings.TrimSpace(item.Module) == "" {
		return fmt.Errorf("module is required")
	}
	if item.IntervalSeconds < 10 {
		return fmt.Errorf("interval_seconds must be at least 10")
	}
	if item.TimeoutSeconds < 1 {
		return fmt.Errorf("timeout_seconds must be positive")
	}
	if item.JitterSeconds < 0 {
		return fmt.Errorf("jitter_seconds cannot be negative")
	}
	if item.FetchSource != "command" && item.FetchSource != "dashboard" && item.FetchSource != "mixed" {
		return fmt.Errorf("fetch_source must be command, dashboard or mixed")
	}
	if item.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	if item.RetryBackoffSeconds < 0 {
		return fmt.Errorf("retry_backoff_seconds cannot be negative")
	}
	return nil
}
