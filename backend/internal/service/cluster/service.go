package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/connection"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

var ErrNotFound = errors.New("cluster not found")

const asyncProbeTimeout = 5 * time.Minute

type CreateInput struct{ Name, MonitorAddresses, ClientUsername, ClientKey string }
type UpdateInput struct {
	Name, MonitorAddresses, ClientUsername *string
	ClientKey                              *string
}
type Service struct {
	database      func() *store.Database
	encryptionKey string
	provider      cephprovider.ClusterProvider
}

func New(database func() *store.Database, encryptionKey string, provider cephprovider.ClusterProvider) *Service {
	return &Service{database: database, encryptionKey: encryptionKey, provider: provider}
}
func (s *Service) List(ctx context.Context, filter store.ClusterFilter) ([]store.CephCluster, error) {
	return s.database().ListClusters(ctx, filter)
}
func (s *Service) FilterOptions(ctx context.Context, fields []string) (store.ClusterFilterOptions, error) {
	return s.database().ClusterFilterOptions(ctx, fields)
}
func (s *Service) Get(ctx context.Context, id uint64) (store.CephCluster, error) {
	row, err := s.database().FindCluster(ctx, id)
	if errors.Is(err, store.ErrRecordNotFound) {
		return store.CephCluster{}, ErrNotFound
	}
	return row, err
}
func (s *Service) Create(ctx context.Context, input CreateInput) (store.CephCluster, cephdomain.ActionResult, error) {
	name, mon, user, key, err := validateCreate(input)
	if err != nil {
		return store.CephCluster{}, cephdomain.ActionResult{}, err
	}
	if _, err := connection.ParseMonitorAddresses(mon); err != nil {
		return store.CephCluster{}, cephdomain.ActionResult{}, fmt.Errorf("invalid monitor addresses: %w", err)
	}
	encrypted, err := security.Encrypt([]byte(key), s.encryptionKey)
	if err != nil {
		return store.CephCluster{}, cephdomain.ActionResult{}, err
	}
	now := time.Now().UTC()
	row := store.CephCluster{Name: name, MonitorAddresses: mon, ClientUsername: user, ClientKey: encrypted, DiscoveredData: "{}", Status: "unknown", Enabled: true, CreatedAt: now, UpdatedAt: now}
	if err := s.database().CreateCluster(ctx, &row); err != nil {
		return store.CephCluster{}, cephdomain.ActionResult{}, err
	}
	s.scheduleProbe(row)
	return row, clusterSavedResult(row.ID), nil
}
func (s *Service) Update(ctx context.Context, id uint64, input UpdateInput) (cephdomain.ActionResult, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	candidate := row
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if value == "" {
			return cephdomain.ActionResult{}, fmt.Errorf("name must not be empty")
		}
		candidate.Name = value
	}
	if input.MonitorAddresses != nil {
		value := strings.TrimSpace(*input.MonitorAddresses)
		if _, err := connection.ParseMonitorAddresses(value); err != nil {
			return cephdomain.ActionResult{}, fmt.Errorf("invalid monitor addresses: %w", err)
		}
		candidate.MonitorAddresses = value
	}
	if input.ClientUsername != nil {
		value := strings.TrimSpace(*input.ClientUsername)
		if !strings.HasPrefix(value, "client.") {
			return cephdomain.ActionResult{}, fmt.Errorf("client_username must start with client.")
		}
		candidate.ClientUsername = value
	}
	if input.ClientKey != nil {
		if *input.ClientKey == "" {
			return cephdomain.ActionResult{}, fmt.Errorf("client_key must not be empty")
		}
		candidate.ClientKey, err = security.Encrypt([]byte(*input.ClientKey), s.encryptionKey)
		if err != nil {
			return cephdomain.ActionResult{}, err
		}
	}
	if candidate == row {
		return cephdomain.ActionResult{}, fmt.Errorf("at least one field is required")
	}
	candidate.UpdatedAt = time.Now().UTC()
	if err := s.database().SaveCluster(ctx, &candidate); err != nil {
		return cephdomain.ActionResult{}, err
	}
	s.scheduleProbe(candidate)
	return clusterSavedResult(candidate.ID), nil
}
func (s *Service) Delete(ctx context.Context, id uint64, deleteCachedData bool) (cephdomain.ActionResult, error) {
	if !deleteCachedData {
		return cephdomain.ActionResult{}, fmt.Errorf("delete_cached_data=true is required")
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if err := s.database().DeleteCluster(ctx, &row); err != nil {
		return cephdomain.ActionResult{}, err
	}
	return cephdomain.ActionResult{Details: map[string]any{"deleted": true}}, nil
}
func (s *Service) Probe(ctx context.Context, id uint64) (cephdomain.ActionResult, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	return s.applyProbe(ctx, row, row, false)
}
func (s *Service) Capabilities(ctx context.Context, id uint64) ([]store.CephClusterCapability, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	return s.database().ListCapabilities(ctx, id)
}
func (s *Service) Access(ctx context.Context, id uint64) (cephprovider.ClusterAccess, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return cephprovider.ClusterAccess{}, err
	}
	return s.accessForRow(row)
}
func (s *Service) applyProbe(ctx context.Context, row, candidate store.CephCluster, saveCandidate bool) (cephdomain.ActionResult, error) {
	access, err := s.accessForRow(candidate)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	result, err := s.provider.Probe(ctx, access)
	access.ClientKey = ""
	now := time.Now().UTC()
	if err != nil {
		code, message := probeFailure(err)
		data, _ := json.Marshal(map[string]any{"status": "unavailable", "error_code": code, "error_message": message, "observed_at": now})
		row.DiscoveredData = string(data)
		row.Status, row.Enabled, row.LastErrorCode, row.LastErrorMessage, row.ObservedAt, row.UpdatedAt = "unavailable", true, &code, &message, &now, now
		_ = s.database().UpdateClusterDiscovery(ctx, row)
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: message, Retryable: true}
	}
	if saveCandidate {
		candidate.UpdatedAt = now
		if err := s.database().SaveCluster(ctx, &candidate); err != nil {
			return cephdomain.ActionResult{}, err
		}
	}
	version := result.Version
	fsid := result.FSID
	discovered, err := json.Marshal(result)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	row.DiscoveredData = string(discovered)
	row.FSID, row.CephVersion, row.Status, row.Enabled = &fsid, &version, "available", true
	row.Generation, row.LastSeenAt, row.ObservedAt, row.UpdatedAt = max(row.Generation, 1), &now, &now, now
	row.LastErrorCode, row.LastErrorMessage = nil, nil
	if err := s.database().UpdateClusterDiscovery(ctx, row); err != nil {
		return cephdomain.ActionResult{}, err
	}
	caps := make([]store.CephClusterCapability, 0, len(result.Capabilities))
	for _, capability := range result.Capabilities {
		var reason, version *string
		if capability.Reason != "" {
			reason = &capability.Reason
		}
		if capability.Version != "" {
			version = &capability.Version
		}
		var details *string
		if capability.Details != nil {
			encoded, _ := json.Marshal(capability.Details)
			value := string(encoded)
			details = &value
		}
		caps = append(caps, store.CephClusterCapability{ClusterID: row.ID, Name: capability.Name, Supported: capability.Supported, Reason: reason, Version: version, DetailsJSON: details, ObservedAt: now, UpdatedAt: now})
	}
	if err := s.database().UpsertCapabilities(ctx, caps); err != nil {
		return cephdomain.ActionResult{}, err
	}
	return cephdomain.ActionResult{ResourceURL: fmt.Sprintf("/api/v1/cluster/%d", row.ID), Details: map[string]any{"fsid": result.FSID, "ceph_version": result.Version}}, nil
}

func (s *Service) scheduleProbe(row store.CephCluster) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), asyncProbeTimeout)
		defer cancel()
		_, _ = s.applyProbe(ctx, row, row, false)
	}()
}

func clusterSavedResult(clusterID uint64) cephdomain.ActionResult {
	return cephdomain.ActionResult{ResourceURL: fmt.Sprintf("/api/v1/cluster/%d", clusterID)}
}

func (s *Service) accessForRow(row store.CephCluster) (cephprovider.ClusterAccess, error) {
	plain, err := security.Decrypt(row.ClientKey, s.encryptionKey)
	if err != nil {
		return cephprovider.ClusterAccess{}, fmt.Errorf("decrypt cluster credential: %w", err)
	}
	key := string(plain)
	for i := range plain {
		plain[i] = 0
	}
	return cephprovider.ClusterAccess{MonitorAddresses: row.MonitorAddresses, ClientUsername: row.ClientUsername, ClientKey: key}, nil
}

func probeFailure(err error) (string, string) {
	var commandErr *executor.Error
	if errors.As(err, &commandErr) && commandErr.Kind == "timeout" {
		return "ceph_probe_timeout", "Ceph probe timed out. Verify that the MON addresses and ports are reachable from the CephTower backend. Common MON ports are 3300 and 6789. Also verify the client user and key."
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "ceph_probe_timeout", "Ceph probe timed out or was cancelled. Verify that the MON addresses and ports are reachable from the CephTower backend."
	}
	return "ceph_unavailable", security.Redact(err.Error())
}

func validateCreate(input CreateInput) (string, string, string, string, error) {
	name := strings.TrimSpace(input.Name)
	mon := strings.TrimSpace(input.MonitorAddresses)
	user := strings.TrimSpace(input.ClientUsername)
	if name == "" || mon == "" || user == "" || input.ClientKey == "" {
		return "", "", "", "", fmt.Errorf("name, monitor_addresses, client_username and client_key are required")
	}
	if !strings.HasPrefix(user, "client.") {
		return "", "", "", "", fmt.Errorf("client_username must start with client.")
	}
	return name, mon, user, input.ClientKey, nil
}
