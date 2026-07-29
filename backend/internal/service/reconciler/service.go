package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/store"
)

type Module struct {
	Name     string
	Interval time.Duration
	Kinds    []string
}

var DefaultModules = []Module{
	{Name: "fast", Interval: 15 * time.Second, Kinds: []string{"overview", "health_check"}},
	{Name: "topology", Interval: 30 * time.Second, Kinds: []string{"host", "service", "daemon", "mon", "mgr", "mds", "upgrade"}},
	{Name: "storage", Interval: time.Minute, Kinds: []string{"osd", "osd_flag", "osd_removal", "pool", "filesystem", "subvolume_group", "subvolume", "cephfs_snapshot", "rbd_image", "rbd_snapshot", "rbd_namespace", "rbd_trash", "rbd_group", "rbd_mirroring", "rgw_status", "rgw_user", "rgw_account", "rgw_role", "rgw_bucket", "rgw_realm", "rgw_zonegroup", "rgw_zone", "nfs_cluster", "nfs_export", "smb_cluster", "smb_share"}},
	{Name: "inventory", Interval: 5 * time.Minute, Kinds: []string{"device", "capability"}},
	{Name: "configuration", Interval: 10 * time.Minute, Kinds: []string{"config_value", "config_option", "mgr_module", "crush_rule", "erasure_code_profile"}},
}

type breaker struct {
	failures int
	next     time.Time
}
type Service struct {
	database func() *store.Database
	clusters *clusterservice.Service
	provider cephprovider.CollectorProvider
	modules  []Module
	mu       sync.Mutex
	breakers map[uint64]*breaker
	runCtx   context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func New(database func() *store.Database, clusters *clusterservice.Service, provider cephprovider.CollectorProvider) *Service {
	return &Service{database: database, clusters: clusters, provider: provider, modules: DefaultModules, breakers: map[uint64]*breaker{}}
}

func (s *Service) Refresh(ctx context.Context, clusterID uint64, modules []string) (cephdomain.ActionResult, error) {
	if clusterID == 0 {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "invalid_request", Message: "cluster is required"}
	}
	selected := make(map[string]bool, len(modules))
	for _, name := range modules {
		selected[name] = true
	}
	available := make(map[string]Module, len(s.modules))
	for _, module := range s.modules {
		available[module.Name] = module
	}
	if len(selected) == 0 {
		for name := range available {
			selected[name] = true
		}
	}
	for name := range selected {
		if _, ok := available[name]; !ok {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "invalid_request", Message: "unsupported refresh module " + name}
		}
	}
	names := make([]string, 0, len(selected))
	for _, module := range s.modules {
		if !selected[module.Name] {
			continue
		}
		if err := s.Reconcile(ctx, clusterID, module); err != nil {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "refresh_failed", Message: security.Redact(err.Error()), Retryable: true}
		}
		names = append(names, module.Name)
	}
	return cephdomain.ActionResult{Details: map[string]any{"modules": names}}, nil
}

func (s *Service) RefreshKind(ctx context.Context, clusterID uint64, kind string) (cephdomain.ActionResult, error) {
	module, ok := s.moduleForKind(kind)
	if !ok {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "invalid_request", Message: "unsupported refresh kind " + kind}
	}
	if err := s.ReconcileKinds(ctx, clusterID, module, []string{kind}); err != nil {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "refresh_failed", Message: security.Redact(err.Error()), Retryable: true}
	}
	return cephdomain.ActionResult{Details: map[string]any{"kinds": []string{kind}, "module": module.Name}}, nil
}

func (s *Service) RefreshKinds(ctx context.Context, clusterID uint64, kinds []string) (cephdomain.ActionResult, error) {
	if len(kinds) == 0 {
		return s.Refresh(ctx, clusterID, nil)
	}
	grouped := map[string][]string{}
	modules := map[string]Module{}
	for _, kind := range kinds {
		module, ok := s.moduleForKind(kind)
		if !ok {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "invalid_request", Message: "unsupported refresh kind " + kind}
		}
		grouped[module.Name] = append(grouped[module.Name], kind)
		modules[module.Name] = module
	}
	for _, module := range s.modules {
		selected := grouped[module.Name]
		if len(selected) == 0 {
			continue
		}
		if err := s.ReconcileKinds(ctx, clusterID, modules[module.Name], selected); err != nil {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "refresh_failed", Message: security.Redact(err.Error()), Retryable: true}
		}
	}
	return cephdomain.ActionResult{Details: map[string]any{"kinds": kinds}}, nil
}

func (s *Service) Start(ctx context.Context) {
	if s.cancel != nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.runCtx = runCtx
	s.cancel = cancel
	for _, module := range s.modules {
		module := module
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.runModule(runCtx, module)
			ticker := time.NewTicker(module.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-runCtx.Done():
					return
				case <-ticker.C:
					s.runModule(runCtx, module)
				}
			}
		}()
	}
}
func (s *Service) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
		s.cancel = nil
		s.runCtx = nil
	}
}
func (s *Service) runModule(ctx context.Context, module Module) {
	clusters, err := s.clusters.List(ctx)
	if err != nil {
		return
	}
	for _, cluster := range clusters {
		if s.allowed(cluster.ID, time.Now()) {
			_ = s.Reconcile(ctx, cluster.ID, module)
		}
	}
}
func (s *Service) Reconcile(ctx context.Context, clusterID uint64, module Module) error {
	return s.reconcile(ctx, clusterID, module, nil)
}

func (s *Service) ReconcileKinds(ctx context.Context, clusterID uint64, module Module, kinds []string) error {
	if len(kinds) == 0 {
		return s.Reconcile(ctx, clusterID, module)
	}
	return s.reconcile(ctx, clusterID, module, kindSet(kinds))
}

func (s *Service) reconcile(ctx context.Context, clusterID uint64, module Module, selectedKinds map[string]struct{}) error {
	generation, err := s.database().NextCollectionGeneration(ctx, clusterID, module.Name)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	run := store.CephCollectionRun{ClusterID: clusterID, Module: module.Name, Generation: generation, Status: "running", Source: "ceph_cli", StartedAt: now, CreatedAt: now}
	if err := s.database().CreateCollectionRun(ctx, &run); err != nil {
		return err
	}
	access, err := s.clusters.Access(ctx, clusterID)
	var records []store.CephResourceRecord
	if err == nil {
		var observations []cephprovider.Observation
		var authoritativeKinds []string
		if provider, ok := s.provider.(cephprovider.CollectionMetadataProvider); ok {
			var result cephprovider.CollectionResult
			result, err = provider.CollectWithMetadata(ctx, access, module.Name)
			observations = result.Observations
			authoritativeKinds = availableKinds(module.Kinds, result.UnavailableKinds)
		} else {
			observations, err = s.provider.Collect(ctx, access, module.Name)
			authoritativeKinds = observedKinds(observations)
		}
		access.ClientKey = ""
		if err == nil {
			records = make([]store.CephResourceRecord, 0, len(observations))
			for _, observation := range observations {
				redacted, redactErr := security.RedactJSON(observation.Payload)
				if redactErr != nil {
					err = redactErr
					break
				}
				payload, marshalErr := json.Marshal(redacted)
				if marshalErr != nil {
					err = marshalErr
					break
				}
				var name, status *string
				if observation.Name != "" {
					value := observation.Name
					name = &value
				}
				if observation.Status != "" {
					value := observation.Status
					status = &value
				}
				records = append(records, store.CephResourceRecord{Kind: observation.Kind, NaturalKey: observation.NaturalKey, ParentKind: optionalString(observation.ParentKind), ParentKey: optionalString(observation.ParentKey), Name: name, Status: status, Source: observation.Source, SourceVersion: optionalString(observation.SourceVersion), ObservedAt: observation.ObservedAt, PayloadSchemaVersion: 1, PayloadJSON: string(payload)})
			}
			if err == nil {
				if len(selectedKinds) > 0 {
					records = filterRecords(records, selectedKinds)
					authoritativeKinds = filterKinds(authoritativeKinds, selectedKinds)
				}
				err = s.database().ReconcileResources(ctx, clusterID, generation, records, authoritativeKinds)
			}
		}
	}
	finished := time.Now().UTC()
	if err == nil {
		_ = s.database().FinishCollectionRun(ctx, run.ID, "succeeded", uint64(len(records)), nil, nil, finished)
		s.success(clusterID)
		return nil
	}
	message := security.Redact(fmt.Sprint(err))
	code := "ceph_unavailable"
	_ = s.database().FinishCollectionRun(ctx, run.ID, "failed", 0, &code, &message, finished)
	staleKinds := module.Kinds
	if len(selectedKinds) > 0 {
		staleKinds = filterKinds(module.Kinds, selectedKinds)
	}
	_ = s.database().MarkModuleResourcesStale(ctx, clusterID, staleKinds, finished)
	s.failure(clusterID)
	return err
}

func availableKinds(moduleKinds, unavailableKinds []string) []string {
	unavailable := make(map[string]struct{}, len(unavailableKinds))
	for _, kind := range unavailableKinds {
		unavailable[kind] = struct{}{}
	}
	result := make([]string, 0, len(moduleKinds))
	for _, kind := range moduleKinds {
		if _, missing := unavailable[kind]; !missing {
			result = append(result, kind)
		}
	}
	return result
}

func observedKinds(observations []cephprovider.Observation) []string {
	set := make(map[string]struct{})
	for _, observation := range observations {
		set[observation.Kind] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for kind := range set {
		result = append(result, kind)
	}
	return result
}
func (s *Service) allowed(id uint64, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.breakers[id]
	return state == nil || !now.Before(state.next)
}
func (s *Service) success(id uint64) { s.mu.Lock(); delete(s.breakers, id); s.mu.Unlock() }
func (s *Service) failure(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.breakers[id]
	if state == nil {
		state = &breaker{}
		s.breakers[id] = state
	}
	state.failures++
	delay := time.Second * time.Duration(1<<min(state.failures, 8))
	state.next = time.Now().Add(delay)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) moduleForKind(kind string) (Module, bool) {
	for _, module := range s.modules {
		for _, candidate := range module.Kinds {
			if candidate == kind {
				return module, true
			}
		}
	}
	return Module{}, false
}

func kindSet(kinds []string) map[string]struct{} {
	result := make(map[string]struct{}, len(kinds))
	for _, kind := range kinds {
		if kind != "" {
			result[kind] = struct{}{}
		}
	}
	return result
}

func filterRecords(rows []store.CephResourceRecord, kinds map[string]struct{}) []store.CephResourceRecord {
	result := make([]store.CephResourceRecord, 0, len(rows))
	for _, row := range rows {
		if _, ok := kinds[row.Kind]; ok {
			result = append(result, row)
		}
	}
	return result
}

func filterKinds(values []string, kinds map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := kinds[value]; ok {
			result = append(result, value)
		}
	}
	return result
}
