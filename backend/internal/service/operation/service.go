package operation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/logging"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const (
	defaultWorkers      = 4
	defaultPollInterval = 250 * time.Millisecond
	operationRetention  = 90 * 24 * time.Hour
)

var ErrIdempotencyConflict = errors.New("idempotency key is already used by another operation")

type Dispatcher interface {
	Execute(context.Context, ExecutionRequest) (cephdomain.ActionResult, error)
}

type ExecutionRequest struct {
	ClusterID    uint64
	Action       string
	ResourceKind string
	ResourceKey  string
	Parameters   map[string]any
}

type EnqueueRequest struct {
	ClusterID       uint64
	ActorUserID     *uint64
	RequestID       string
	IdempotencyKey  string
	Action          string
	ResourceKind    string
	ResourceKey     string
	Risk            string
	LockKey         string
	ExpectedVersion *uint64
	Parameters      map[string]any
}

type Options struct {
	Workers      int
	PollInterval time.Duration
}

type clusterLock struct {
	mu   sync.Mutex
	refs int
}

type Service struct {
	database      func() *store.Database
	encryptionKey string
	dispatcher    Dispatcher
	workers       int
	pollInterval  time.Duration
	wake          chan struct{}

	lifecycleMu sync.Mutex
	cancel      context.CancelFunc
	wg          sync.WaitGroup

	lockMu sync.Mutex
	locks  map[uint64]*clusterLock
}

func New(database func() *store.Database, encryptionKey string, dispatcher Dispatcher, options Options) *Service {
	workers := options.Workers
	if workers <= 0 {
		workers = defaultWorkers
	}
	pollInterval := options.PollInterval
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}
	return &Service{
		database: database, encryptionKey: encryptionKey, dispatcher: dispatcher,
		workers: workers, pollInterval: pollInterval, wake: make(chan struct{}, 1),
		locks: map[uint64]*clusterLock{},
	}
}

func (s *Service) Enqueue(ctx context.Context, request EnqueueRequest) (store.CephOperation, error) {
	if request.ClusterID == 0 || request.Action == "" || request.ResourceKind == "" {
		return store.CephOperation{}, fmt.Errorf("cluster, action, and resource kind are required")
	}
	payload, err := json.Marshal(request.Parameters)
	if err != nil {
		return store.CephOperation{}, fmt.Errorf("encode operation parameters: %w", err)
	}
	if string(payload) == "null" {
		payload = []byte("{}")
	}
	if request.IdempotencyKey != "" {
		if existing, err := s.database().FindOperationByIdempotencyKey(ctx, request.ClusterID, request.IdempotencyKey); err == nil {
			if !s.sameOperation(existing, request, payload) {
				return store.CephOperation{}, ErrIdempotencyConflict
			}
			return existing, nil
		} else if !errors.Is(err, store.ErrRecordNotFound) {
			return store.CephOperation{}, err
		}
	}
	ciphertext, err := security.Encrypt(payload, s.encryptionKey)
	if err != nil {
		return store.CephOperation{}, fmt.Errorf("encrypt operation parameters: %w", err)
	}
	now := time.Now().UTC()
	row := store.CephOperation{
		ClusterID: request.ClusterID, ActorUserID: request.ActorUserID, RequestID: request.RequestID,
		Action: request.Action, ResourceKind: request.ResourceKind, ResourceKey: request.ResourceKey,
		Risk: request.Risk, LockKey: request.LockKey, Status: store.OperationQueued,
		ParametersCiphertext: ciphertext, ExpectedVersion: request.ExpectedVersion,
		MaxAttempts: maxAttemptsFor(request.Action), CreatedAt: now, UpdatedAt: now,
	}
	if request.IdempotencyKey != "" {
		row.IdempotencyKey = &request.IdempotencyKey
	}
	if err := s.database().CreateOperation(ctx, &row); err != nil {
		if request.IdempotencyKey != "" {
			if existing, findErr := s.database().FindOperationByIdempotencyKey(ctx, request.ClusterID, request.IdempotencyKey); findErr == nil {
				if !s.sameOperation(existing, request, payload) {
					return store.CephOperation{}, ErrIdempotencyConflict
				}
				return existing, nil
			}
		}
		return store.CephOperation{}, err
	}
	s.signal()
	return row, nil
}

func (s *Service) sameOperation(existing store.CephOperation, request EnqueueRequest, payload []byte) bool {
	if existing.Action != request.Action ||
		existing.ResourceKind != request.ResourceKind ||
		existing.ResourceKey != request.ResourceKey {
		return false
	}
	existingPayload, err := security.Decrypt(existing.ParametersCiphertext, s.encryptionKey)
	return err == nil && bytes.Equal(existingPayload, payload)
}

func (s *Service) Start(ctx context.Context) error {
	if s.dispatcher == nil {
		return fmt.Errorf("operation dispatcher is required")
	}
	if s.database == nil || s.database() == nil {
		return fmt.Errorf("operation database is unavailable")
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.cancel != nil {
		return nil
	}
	runCtx, cancel := context.WithCancel(ctx)
	if _, err := s.database().RecoverRunningOperations(runCtx, time.Now().UTC()); err != nil {
		cancel()
		return fmt.Errorf("recover interrupted operations: %w", err)
	}
	s.cancel = cancel
	s.wg.Add(1)
	go s.maintainRetention(runCtx)
	for range s.workers {
		s.wg.Add(1)
		go s.worker(runCtx)
	}
	return nil
}

func (s *Service) maintainRetention(ctx context.Context) {
	defer s.wg.Done()
	s.pruneCompleted(ctx, time.Now().UTC())
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.pruneCompleted(ctx, now.UTC())
		}
	}
}

func (s *Service) pruneCompleted(ctx context.Context, now time.Time) {
	if _, err := s.database().PruneCompletedOperations(ctx, now.Add(-operationRetention)); err != nil && ctx.Err() == nil {
		logging.Warnf("operation retention failed: error=%v", err)
	}
}

func (s *Service) Stop() {
	s.lifecycleMu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.lifecycleMu.Unlock()
	if cancel != nil {
		cancel()
		s.wg.Wait()
	}
}

func (s *Service) worker(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		for {
			row, err := s.database().ClaimNextOperation(ctx, time.Now().UTC())
			if errors.Is(err, store.ErrRecordNotFound) {
				break
			}
			if err != nil {
				if ctx.Err() == nil {
					logging.Warnf("operation claim failed: error=%v", err)
				}
				break
			}
			s.recordAudit(ctx, row, "operation_started", "started", nil, nil)
			s.execute(ctx, row)
			if ctx.Err() != nil {
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-s.wake:
		case <-ticker.C:
		}
	}
}

func (s *Service) execute(ctx context.Context, row store.CephOperation) {
	unlock := s.lockCluster(row.ClusterID)
	defer unlock()
	if err := s.validateExpectedVersion(ctx, row); err != nil {
		code, message, retryable := normalizeError(err)
		s.fail(ctx, row, code, message, retryable)
		return
	}
	parameters, err := s.parameters(row)
	if err != nil {
		s.fail(ctx, row, "invalid_operation_payload", err.Error(), false)
		return
	}
	result, err := s.dispatcher.Execute(ctx, ExecutionRequest{
		ClusterID: row.ClusterID, Action: row.Action, ResourceKind: row.ResourceKind,
		ResourceKey: row.ResourceKey, Parameters: parameters,
	})
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		code, message, retryable := normalizeError(err)
		s.fail(ctx, row, code, message, retryable)
		return
	}
	redacted, err := security.RedactJSON(result)
	if err != nil {
		s.fail(ctx, row, "result_encoding_failed", "operation result could not be encoded", false)
		return
	}
	encoded, err := json.Marshal(redacted)
	if err != nil {
		s.fail(ctx, row, "result_encoding_failed", "operation result could not be encoded", false)
		return
	}
	if err := s.database().CompleteOperation(ctx, row.ID, string(encoded), time.Now().UTC()); err != nil {
		if ctx.Err() == nil {
			logging.Errorf("operation completion persistence failed: operation_id=%d error=%v", row.ID, err)
		}
		return
	}
	s.recordAudit(ctx, row, "operation_completed", "succeeded", nil, nil)
}

func (s *Service) validateExpectedVersion(ctx context.Context, row store.CephOperation) error {
	if row.ExpectedVersion == nil {
		return nil
	}
	resource, err := s.database().FindResource(ctx, row.ClusterID, row.ResourceKind, row.LockKey)
	if errors.Is(err, store.ErrRecordNotFound) && *row.ExpectedVersion == 0 {
		return nil
	}
	if err != nil {
		return &cephdomain.ActionError{
			Code: "resource_conflict", Message: "resource version could not be verified before execution",
		}
	}
	if resource.ResourceVersion != *row.ExpectedVersion {
		return &cephdomain.ActionError{Code: "resource_conflict", Message: "resource generation changed"}
	}
	return nil
}

func (s *Service) parameters(row store.CephOperation) (map[string]any, error) {
	payload, err := security.Decrypt(row.ParametersCiphertext, s.encryptionKey)
	if err != nil {
		return nil, err
	}
	var parameters map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&parameters); err != nil {
		return nil, err
	}
	if parameters == nil {
		parameters = map[string]any{}
	}
	return parameters, nil
}

func (s *Service) fail(ctx context.Context, row store.CephOperation, code, message string, retryable bool) {
	message = security.Redact(message)
	if retryable && row.Attempts < row.MaxAttempts {
		now := time.Now().UTC()
		nextAttemptAt := now.Add(retryDelay(row.Attempts))
		if err := s.database().RequeueOperation(ctx, row.ID, code, message, nextAttemptAt, now); err != nil {
			if ctx.Err() == nil {
				logging.Errorf("operation retry persistence failed: operation_id=%d error=%v", row.ID, err)
			}
			return
		}
		s.recordAudit(ctx, row, "operation_retry_scheduled", "retrying", &code, &retryable)
		s.signal()
		return
	}
	if err := s.database().FailOperation(ctx, row.ID, code, message, retryable, time.Now().UTC()); err != nil {
		if ctx.Err() == nil {
			logging.Errorf("operation failure persistence failed: operation_id=%d error=%v", row.ID, err)
		}
		return
	}
	s.recordAudit(ctx, row, "operation_completed", "failed", &code, &retryable)
}

func maxAttemptsFor(action string) uint32 {
	if action == "cluster.refresh" {
		return 3
	}
	return 1
}

func retryDelay(attempt uint32) time.Duration {
	if attempt == 0 {
		attempt = 1
	}
	delay := time.Second << min(attempt-1, 5)
	return min(delay, 30*time.Second)
}

func (s *Service) recordAudit(ctx context.Context, row store.CephOperation, eventType, outcome string, errorCode *string, retryable *bool) {
	clusterID := row.ClusterID
	resourceKind, resourceKey, risk := row.ResourceKind, row.ResourceKey, row.Risk
	actorUsername := "system"
	if row.ActorUserID != nil {
		if user, err := s.database().FindUserByID(ctx, *row.ActorUserID); err == nil {
			actorUsername = user.Username
		}
	}
	var clusterName *string
	if cluster, err := s.database().FindCluster(ctx, row.ClusterID); err == nil {
		clusterName = &cluster.Name
	}
	details := map[string]any{
		"operation_id": row.ID,
		"attempt":      row.Attempts,
		"max_attempts": row.MaxAttempts,
	}
	if retryable != nil {
		details["retryable"] = *retryable
	}
	encoded, err := json.Marshal(details)
	if err != nil {
		return
	}
	detailsJSON := string(encoded)
	event := store.AuditEvent{
		OccurredAt: time.Now().UTC(), EventType: eventType, RequestID: row.RequestID,
		ActorUserID: row.ActorUserID, ActorUsername: actorUsername,
		ClusterID: &clusterID, ClusterName: clusterName, Action: row.Action,
		ResourceKind: &resourceKind, ResourceKey: &resourceKey, Risk: &risk,
		Outcome: outcome, ErrorCode: errorCode, BeforeGeneration: row.ExpectedVersion,
		DetailsJSON: &detailsJSON,
	}
	if err := s.database().CreateAuditEvent(ctx, &event); err != nil && ctx.Err() == nil {
		logging.Errorf("operation audit persistence failed: operation_id=%d event=%s error=%v", row.ID, eventType, err)
	}
}

func normalizeError(err error) (string, string, bool) {
	var actionError *cephdomain.ActionError
	if errors.As(err, &actionError) {
		return actionError.Code, actionError.Message, actionError.Retryable
	}
	return "action_failed", err.Error(), true
}

func (s *Service) lockCluster(clusterID uint64) func() {
	s.lockMu.Lock()
	lock := s.locks[clusterID]
	if lock == nil {
		lock = &clusterLock{}
		s.locks[clusterID] = lock
	}
	lock.refs++
	s.lockMu.Unlock()
	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		s.lockMu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(s.locks, clusterID)
		}
		s.lockMu.Unlock()
	}
}

func (s *Service) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}
