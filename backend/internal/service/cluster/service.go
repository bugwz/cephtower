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
	"cephtower/backend/internal/security"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

var ErrNotFound = errors.New("cluster not found")

type CreateInput struct{ Name, MonitorAddresses, ClientUsername, ClientKey string }
type UpdateInput struct {
	Name, MonitorAddresses, ClientUsername *string
	ClientKey                              *string
}
type Service struct {
	database      func() *store.Database
	encryptionKey string
	operations    *operationservice.Service
	provider      cephprovider.ClusterProvider
}

func New(database func() *store.Database, encryptionKey string, operations *operationservice.Service, provider cephprovider.ClusterProvider) *Service {
	s := &Service{database: database, encryptionKey: encryptionKey, operations: operations, provider: provider}
	if operations != nil {
		_ = operations.Register("cluster.probe", s.probeOperation)
		_ = operations.Register("cluster.update", s.probeOperation)
		_ = operations.Register("cluster.delete", s.deleteOperation)
	}
	return s
}
func (s *Service) List(ctx context.Context) ([]store.CephCluster, error) {
	return s.database().ListClusters(ctx)
}
func (s *Service) Get(ctx context.Context, id uint64) (store.CephCluster, error) {
	row, err := s.database().FindCluster(ctx, id)
	if errors.Is(err, store.ErrRecordNotFound) {
		return store.CephCluster{}, ErrNotFound
	}
	return row, err
}
func (s *Service) Create(ctx context.Context, input CreateInput, actorUserID *uint64, actor, requestID, idempotencyKey string) (store.CephCluster, store.CephOperation, error) {
	name, mon, user, key, err := validateCreate(input)
	if err != nil {
		return store.CephCluster{}, store.CephOperation{}, err
	}
	if _, err := connection.ParseMonitorAddresses(mon); err != nil {
		return store.CephCluster{}, store.CephOperation{}, fmt.Errorf("invalid monitor addresses: %w", err)
	}
	encrypted, err := security.Encrypt([]byte(key), s.encryptionKey)
	if err != nil {
		return store.CephCluster{}, store.CephOperation{}, err
	}
	now := time.Now().UTC()
	row := store.CephCluster{Name: name, MonitorAddresses: mon, ClientUsername: user, ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := s.database().CreateCluster(ctx, &row); err != nil {
		return store.CephCluster{}, store.CephOperation{}, err
	}
	operation, err := s.operations.Enqueue(ctx, operationservice.Request{ClusterID: &row.ID, ClusterName: row.Name, ActorUserID: actorUserID, ActorUsername: actor, RequestID: requestID, Action: "cluster.probe", ResourceKind: "cluster", ResourceKey: fmt.Sprintf("%d", row.ID), Risk: cephdomain.RiskLow, IdempotencyKey: idempotencyKey, Parameters: map[string]any{"cluster_id": row.ID}, LockKeys: []operationservice.LockKey{{Kind: "cluster", Key: fmt.Sprintf("%d", row.ID)}}})
	if err != nil {
		_ = s.database().DeleteCluster(ctx, &row)
		return store.CephCluster{}, store.CephOperation{}, err
	}
	return row, operation, nil
}
func (s *Service) Update(ctx context.Context, id uint64, input UpdateInput, actorUserID *uint64, actor, requestID, idempotencyKey string) (store.CephOperation, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return store.CephOperation{}, err
	}
	parameters := map[string]any{}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if value == "" {
			return store.CephOperation{}, fmt.Errorf("name must not be empty")
		}
		parameters["name"] = value
	}
	if input.MonitorAddresses != nil {
		value := strings.TrimSpace(*input.MonitorAddresses)
		if _, err := connection.ParseMonitorAddresses(value); err != nil {
			return store.CephOperation{}, fmt.Errorf("invalid monitor addresses: %w", err)
		}
		parameters["monitor_addresses"] = value
	}
	if input.ClientUsername != nil {
		value := strings.TrimSpace(*input.ClientUsername)
		if !strings.HasPrefix(value, "client.") {
			return store.CephOperation{}, fmt.Errorf("client_username must start with client.")
		}
		parameters["client_username"] = value
	}
	if input.ClientKey != nil {
		if *input.ClientKey == "" {
			return store.CephOperation{}, fmt.Errorf("client_key must not be empty")
		}
		parameters["client_key"] = *input.ClientKey
	}
	if len(parameters) == 0 {
		return store.CephOperation{}, fmt.Errorf("at least one field is required")
	}
	return s.operations.Enqueue(ctx, operationservice.Request{ClusterID: &row.ID, ClusterName: row.Name, ActorUserID: actorUserID, ActorUsername: actor, RequestID: requestID, Action: "cluster.update", ResourceKind: "cluster", ResourceKey: fmt.Sprintf("%d", row.ID), Risk: cephdomain.RiskMedium, IdempotencyKey: idempotencyKey, Parameters: parameters})
}
func (s *Service) Delete(ctx context.Context, id uint64, deleteCachedData bool, actorUserID *uint64, actor, requestID, idempotencyKey string) (store.CephOperation, error) {
	if !deleteCachedData {
		return store.CephOperation{}, fmt.Errorf("delete_cached_data=true is required")
	}
	row, err := s.Get(ctx, id)
	if err != nil {
		return store.CephOperation{}, err
	}
	active, err := s.database().CountNonTerminalOperations(ctx, id, "")
	if err != nil {
		return store.CephOperation{}, err
	}
	if active != 0 {
		return store.CephOperation{}, fmt.Errorf("cluster has %d non-terminal operations", active)
	}
	return s.operations.Enqueue(ctx, operationservice.Request{ClusterID: &row.ID, ClusterName: row.Name, ActorUserID: actorUserID, ActorUsername: actor, RequestID: requestID, Action: "cluster.delete", ResourceKind: "cluster", ResourceKey: fmt.Sprintf("%d", row.ID), Risk: cephdomain.RiskMedium, IdempotencyKey: idempotencyKey, Parameters: map[string]any{"delete_cached_data": true}})
}
func (s *Service) Probe(ctx context.Context, id uint64, actorUserID *uint64, actor, requestID, idempotencyKey string) (store.CephOperation, error) {
	row, err := s.Get(ctx, id)
	if err != nil {
		return store.CephOperation{}, err
	}
	return s.operations.Enqueue(ctx, operationservice.Request{ClusterID: &row.ID, ClusterName: row.Name, ActorUserID: actorUserID, ActorUsername: actor, RequestID: requestID, Action: "cluster.probe", ResourceKind: "cluster", ResourceKey: fmt.Sprintf("%d", row.ID), Risk: cephdomain.RiskLow, IdempotencyKey: idempotencyKey, Parameters: map[string]any{"probe": true}})
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
func (s *Service) probeOperation(ctx context.Context, operation store.CephOperation) (cephdomain.OperationResult, error) {
	if operation.ClusterID == nil {
		return cephdomain.OperationResult{}, fmt.Errorf("cluster operation has no cluster")
	}
	row, err := s.Get(ctx, *operation.ClusterID)
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	candidate := row
	if operation.Action == "cluster.update" {
		var stored any
		if err := json.Unmarshal([]byte(operation.RequestJSON), &stored); err != nil {
			return cephdomain.OperationResult{}, err
		}
		stored, err = security.UnprotectJSON(stored, s.encryptionKey)
		if err != nil {
			return cephdomain.OperationResult{}, &cephdomain.OperationError{Code: "invalid_credential", Message: "cluster update secrets could not be decrypted"}
		}
		parameters, ok := stored.(map[string]any)
		if !ok {
			return cephdomain.OperationResult{}, fmt.Errorf("cluster update parameters are invalid")
		}
		if value, ok := parameters["name"].(string); ok {
			candidate.Name = value
		}
		if value, ok := parameters["monitor_addresses"].(string); ok {
			candidate.MonitorAddresses = value
		}
		if value, ok := parameters["client_username"].(string); ok {
			candidate.ClientUsername = value
		}
		if value, ok := parameters["client_key"].(string); ok {
			candidate.ClientKey, err = security.Encrypt([]byte(value), s.encryptionKey)
			if err != nil {
				return cephdomain.OperationResult{}, err
			}
		}
	}
	access, err := s.accessForRow(candidate)
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	result, err := s.provider.Probe(ctx, access)
	access.ClientKey = ""
	now := time.Now().UTC()
	if err != nil {
		code := "ceph_unavailable"
		message := security.Redact(err.Error())
		_ = s.database().UpsertObservation(ctx, &store.CephClusterObservation{ClusterID: *operation.ClusterID, Status: "unavailable", Enabled: true, LastErrorCode: &code, LastErrorMessage: &message, ObservedAt: &now, UpdatedAt: now})
		return cephdomain.OperationResult{}, &cephdomain.OperationError{Code: code, Message: message, Retryable: true}
	}
	if operation.Action == "cluster.update" {
		candidate.UpdatedAt = now
		if err := s.database().SaveCluster(ctx, &candidate); err != nil {
			return cephdomain.OperationResult{}, err
		}
	}
	version := result.Version
	fsid := result.FSID
	observation := store.CephClusterObservation{ClusterID: *operation.ClusterID, FSID: &fsid, CephVersion: &version, Status: "available", Enabled: true, Generation: 1, LastSeenAt: &now, ObservedAt: &now, UpdatedAt: now}
	if err := s.database().UpsertObservation(ctx, &observation); err != nil {
		return cephdomain.OperationResult{}, err
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
		caps = append(caps, store.CephClusterCapability{ClusterID: *operation.ClusterID, Name: capability.Name, Supported: capability.Supported, Reason: reason, Version: version, DetailsJSON: details, ObservedAt: now, UpdatedAt: now})
	}
	if err := s.database().UpsertCapabilities(ctx, caps); err != nil {
		return cephdomain.OperationResult{}, err
	}
	return cephdomain.OperationResult{ResourceURL: fmt.Sprintf("/api/v1/cluster/%d", *operation.ClusterID), Details: map[string]any{"fsid": result.FSID, "ceph_version": result.Version}}, nil
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
func (s *Service) deleteOperation(ctx context.Context, operation store.CephOperation) (cephdomain.OperationResult, error) {
	if operation.ClusterID == nil {
		return cephdomain.OperationResult{}, nil
	}
	row, err := s.database().FindCluster(ctx, *operation.ClusterID)
	if errors.Is(err, store.ErrRecordNotFound) {
		return cephdomain.OperationResult{}, nil
	}
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	active, err := s.database().CountNonTerminalOperations(ctx, *operation.ClusterID, operation.ID)
	if err != nil {
		return cephdomain.OperationResult{}, err
	}
	if active != 0 {
		return cephdomain.OperationResult{}, &cephdomain.OperationError{Code: "cluster_busy", Message: "cluster has non-terminal operations", Retryable: true}
	}
	if err := s.database().DeleteCluster(ctx, &row); err != nil {
		return cephdomain.OperationResult{}, err
	}
	return cephdomain.OperationResult{Details: map[string]any{"deleted": true}}, nil
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
