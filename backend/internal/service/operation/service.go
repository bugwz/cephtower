package operation

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

type Handler func(context.Context, store.CephOperation) (cephdomain.OperationResult, error)
type PlanChecker func(context.Context, PlanRequest) ([]string, []string, error)
type PostSuccess func(context.Context, store.CephOperation) error
type Request struct {
	ClusterID                                                   *uint64
	ClusterName                                                 string
	ActorUserID                                                 *uint64
	ActorUsername, RequestID, Action, ResourceKind, ResourceKey string
	ResourceGeneration                                          *uint64
	Risk                                                        cephdomain.Risk
	PlanID, RetryOfID                                           *string
	IdempotencyKey                                              string
	Parameters                                                  any
	LockKeys                                                    []LockKey
}
type LockKey struct{ Kind, Key string }
type PlanRequest struct {
	ClusterID                                                   uint64
	ActorUserID                                                 *uint64
	ActorUsername, RequestID, Action, ResourceKind, ResourceKey string
	ResourceGeneration                                          uint64
	Risk                                                        cephdomain.Risk
	Parameters                                                  any
	Blockers, Warnings                                          []string
}

type Service struct {
	database      func() *store.Database
	encryptionKey string
	handlersMu    sync.RWMutex
	handlers      map[string]Handler
	fallback      Handler
	planChecker   PlanChecker
	postSuccess   PostSuccess
	locksMu       sync.RWMutex
	locks         map[string][]LockKey
	activeMu      sync.Mutex
	active        map[string]context.CancelFunc
	workers       int
	perCluster    int
	poll          time.Duration
	lease         time.Duration
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

func New(database func() *store.Database, workers int, encryptionKeys ...string) *Service {
	if workers <= 0 {
		workers = 8
	}
	encryptionKey := ""
	if len(encryptionKeys) > 0 {
		encryptionKey = encryptionKeys[0]
	}
	return &Service{database: database, encryptionKey: encryptionKey, handlers: map[string]Handler{}, locks: map[string][]LockKey{}, active: map[string]context.CancelFunc{}, workers: workers, perCluster: 2, poll: 250 * time.Millisecond, lease: 30 * time.Second}
}
func (s *Service) Register(action string, handler Handler) error {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()
	if action == "" || handler == nil {
		return fmt.Errorf("action and handler are required")
	}
	if _, ok := s.handlers[action]; ok {
		return fmt.Errorf("operation action %q already registered", action)
	}
	s.handlers[action] = handler
	return nil
}
func (s *Service) SetFallback(handler Handler) {
	s.handlersMu.Lock()
	s.fallback = handler
	s.handlersMu.Unlock()
}
func (s *Service) SetPlanChecker(checker PlanChecker) {
	s.handlersMu.Lock()
	s.planChecker = checker
	s.handlersMu.Unlock()
}
func (s *Service) SetPostSuccess(callback PostSuccess) {
	s.handlersMu.Lock()
	s.postSuccess = callback
	s.handlersMu.Unlock()
}

func (s *Service) CreatePlan(ctx context.Context, request PlanRequest) (store.CephActionPlan, error) {
	if request.Risk != cephdomain.RiskHigh {
		return store.CephActionPlan{}, fmt.Errorf("plans are only valid for high-risk operations")
	}
	if !destructiveAction(request.Action) {
		return store.CephActionPlan{}, fmt.Errorf("action is not registered as high risk")
	}
	{
		lookupKey := resourceLookupKey(request.ResourceKind, request.ResourceKey)
		row, err := s.database().FindResource(ctx, request.ClusterID, request.ResourceKind, lookupKey)
		if err == nil && row.ResourceVersion != request.ResourceGeneration {
			return store.CephActionPlan{}, fmt.Errorf("resource generation changed")
		}
		if err != nil && !errors.Is(err, store.ErrRecordNotFound) {
			return store.CephActionPlan{}, err
		}
	}
	s.handlersMu.RLock()
	checker := s.planChecker
	s.handlersMu.RUnlock()
	if checker != nil {
		blockers, warnings, err := checker(ctx, request)
		if err != nil {
			return store.CephActionPlan{}, fmt.Errorf("pre-check failed: %w", err)
		}
		request.Blockers = append(request.Blockers, blockers...)
		request.Warnings = append(request.Warnings, warnings...)
	}
	payload, err := s.requestJSON(request.Parameters)
	if err != nil {
		return store.CephActionPlan{}, err
	}
	blockers, _ := json.Marshal(request.Blockers)
	warnings, _ := json.Marshal(request.Warnings)
	status := "valid"
	if len(request.Blockers) > 0 {
		status = "blocked"
	}
	now := time.Now().UTC()
	plan := store.CephActionPlan{ID: newUUID(), ClusterID: request.ClusterID, ActorUserID: request.ActorUserID, ActorUsername: request.ActorUsername, RequestID: request.RequestID, Action: request.Action, ResourceKind: request.ResourceKind, ResourceKey: request.ResourceKey, ResourceGeneration: request.ResourceGeneration, Risk: string(request.Risk), Status: status, RequestJSON: string(payload), BlockersJSON: string(blockers), WarningsJSON: string(warnings), ExpiresAt: now.Add(10 * time.Minute), CreatedAt: now}
	return plan, s.database().CreatePlan(ctx, &plan)
}
func destructiveAction(action string) bool {
	return strings.HasSuffix(action, ".delete") || strings.HasSuffix(action, ".purge") || action == "device.zap" || action == "osd_deployment.create" || action == "upgrade.action" || action == "rgw_period.commit"
}

func (s *Service) Enqueue(ctx context.Context, request Request) (store.CephOperation, error) {
	if request.Action == "" || request.ResourceKind == "" || request.ResourceKey == "" {
		return store.CephOperation{}, fmt.Errorf("action and resource are required")
	}
	if request.RequestID == "" {
		request.RequestID = newUUID()
	}
	payload, err := s.requestJSON(request.Parameters)
	if err != nil {
		return store.CephOperation{}, err
	}
	var keyHash, scopeHash *string
	if request.IdempotencyKey != "" {
		key := store.SHA256(request.IdempotencyKey)
		scope := store.SHA256(idempotencySubject(request.ActorUserID) + ":" + idempotencySubject(request.ClusterID) + ":" + request.Action + ":" + request.IdempotencyKey)
		keyHash = &key
		scopeHash = &scope
		if existing, err := s.database().FindIdempotentOperation(ctx, scope); err == nil {
			return existing, nil
		} else if !errors.Is(err, store.ErrRecordNotFound) {
			return store.CephOperation{}, err
		}
	}
	if request.Risk == cephdomain.RiskHigh {
		if request.PlanID == nil || request.ClusterID == nil || request.ResourceGeneration == nil {
			return store.CephOperation{}, fmt.Errorf("a valid plan is required for high-risk operations")
		}
		lookupKey := resourceLookupKey(request.ResourceKind, request.ResourceKey)
		row, err := s.database().FindResource(ctx, *request.ClusterID, request.ResourceKind, lookupKey)
		if err == nil && row.ResourceVersion != *request.ResourceGeneration {
			return store.CephOperation{}, fmt.Errorf("resource generation changed")
		}
		if err != nil && !errors.Is(err, store.ErrRecordNotFound) {
			return store.CephOperation{}, err
		}
		plan, err := s.database().FindPlan(ctx, *request.PlanID)
		if err != nil {
			return store.CephOperation{}, fmt.Errorf("plan is invalid, expired, blocked, consumed, or stale")
		}
		planned, err := s.decodeRequestJSON(plan.RequestJSON)
		if err != nil {
			return store.CephOperation{}, fmt.Errorf("plan parameters could not be verified")
		}
		requested, err := normalizeJSONBytes(request.Parameters)
		if err != nil || !bytes.Equal(planned, requested) {
			return store.CephOperation{}, fmt.Errorf("plan parameters do not match the operation request")
		}
		if err := s.database().ConsumePlan(ctx, *request.PlanID, request.ActorUserID, *request.ClusterID, request.Action, request.ResourceKind, request.ResourceKey, *request.ResourceGeneration, time.Now().UTC()); err != nil {
			return store.CephOperation{}, fmt.Errorf("plan is invalid, expired, blocked, consumed, or stale")
		}
	}
	now := time.Now().UTC()
	operation := store.CephOperation{ID: newUUID(), ClusterID: request.ClusterID, ClusterName: request.ClusterName, ActorUserID: request.ActorUserID, ActorUsername: request.ActorUsername, PlanID: request.PlanID, RetryOfID: request.RetryOfID, RequestID: request.RequestID, Action: request.Action, ResourceKind: request.ResourceKind, ResourceKey: request.ResourceKey, ResourceGeneration: request.ResourceGeneration, Risk: string(request.Risk), Status: "queued", Stage: "queued", Progress: 0, MaxAttempts: 1, IdempotencyKeyHash: keyHash, IdempotencyScopeHash: scopeHash, RequestJSON: string(payload), ScheduledAt: now, CreatedAt: now, UpdatedAt: now}
	event := store.CephOperationEvent{Sequence: 1, EventType: "state", Stage: "queued", Message: "operation queued", CreatedAt: now}
	auditParameters, err := safeJSON(request.Parameters)
	if err != nil {
		return store.CephOperation{}, err
	}
	auditParametersJSON := string(auditParameters)
	httpStatus := 202
	audit := store.AuditEvent{OccurredAt: now, EventType: "accepted", RequestID: request.RequestID, ActorUserID: request.ActorUserID, ActorUsername: request.ActorUsername, ClusterID: request.ClusterID, ClusterName: stringPointer(request.ClusterName), Action: request.Action, ResourceKind: &request.ResourceKind, ResourceKey: &request.ResourceKey, Outcome: "accepted", Risk: stringPointer(string(request.Risk)), HTTPStatus: &httpStatus, PlanID: request.PlanID, BeforeGeneration: request.ResourceGeneration, ParametersJSON: &auditParametersJSON}
	if err := s.database().CreateOperation(ctx, &operation, &event, &audit); err != nil {
		if scopeHash != nil {
			if existing, findErr := s.database().FindIdempotentOperation(ctx, *scopeHash); findErr == nil {
				return existing, nil
			}
		}
		return store.CephOperation{}, err
	}
	s.locksMu.Lock()
	s.locks[operation.ID] = append([]LockKey{}, request.LockKeys...)
	s.locksMu.Unlock()
	return operation, nil
}
func (s *Service) Start(ctx context.Context) error {
	if s.cancel != nil {
		return nil
	}
	workerCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	if err := s.database().RecoverOperations(workerCtx, time.Now().UTC().Add(-30*time.Second), time.Now().UTC()); err != nil {
		return err
	}
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(workerCtx)
	}
	return nil
}
func (s *Service) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
		s.cancel = nil
	}
}
func (s *Service) worker(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOne(ctx)
		}
	}
}
func (s *Service) runOne(ctx context.Context) {
	operation, err := s.database().ClaimQueuedOperation(ctx, time.Now().UTC(), s.perCluster)
	if err != nil {
		return
	}
	_ = s.database().AppendOperationEvent(ctx, &store.CephOperationEvent{OperationID: operation.ID, EventType: "state", Stage: "pre_check", Progress: intPointer(5), Message: "operation started", CreatedAt: time.Now().UTC()})
	s.audit(ctx, operation, "started", "running", nil)
	operationCtx, cancel := context.WithCancel(ctx)
	s.activeMu.Lock()
	s.active[operation.ID] = cancel
	s.activeMu.Unlock()
	defer func() { cancel(); s.activeMu.Lock(); delete(s.active, operation.ID); s.activeMu.Unlock() }()
	keys := s.lockRows(operation)
	if len(keys) > 0 {
		if err := s.database().AcquireLocks(ctx, operation, keys, time.Now().UTC()); err != nil {
			s.fail(ctx, operation, "resource_busy", err.Error(), true)
			return
		}
		defer s.database().ReleaseLocks(context.Background(), operation.ID)
	}
	if err := s.precheckGeneration(ctx, operation); err != nil {
		s.fail(ctx, operation, "resource_conflict", err.Error(), false)
		return
	}
	if err := s.precheckPlan(ctx, operation); err != nil {
		var operationErr *cephdomain.OperationError
		if errors.As(err, &operationErr) {
			s.fail(ctx, operation, operationErr.Code, operationErr.Message, operationErr.Retryable)
		} else {
			s.fail(ctx, operation, "pre_check_failed", err.Error(), false)
		}
		return
	}
	heartbeatDone := make(chan struct{})
	go s.heartbeat(operationCtx, cancel, operation.ID, heartbeatDone)
	defer func() {
		cancel()
		<-heartbeatDone
	}()
	s.handlersMu.RLock()
	handler := s.handlers[operation.Action]
	if handler == nil {
		handler = s.fallback
	}
	s.handlersMu.RUnlock()
	if handler == nil {
		s.fail(ctx, operation, "unsupported_action", "operation action is not registered", false)
		return
	}
	result, err := handler(operationCtx, operation)
	if err != nil {
		if operationCtx.Err() != nil {
			now := time.Now().UTC()
			_ = s.database().UpdateOperation(ctx, operation.ID, map[string]any{"status": "cancelled", "stage": "cancelled", "error_code": nil, "error_message": nil, "retryable": false, "completed_at": now})
			_ = s.database().AppendOperationEvent(ctx, &store.CephOperationEvent{OperationID: operation.ID, EventType: "state", Stage: "cancelled", Progress: intPointer(operation.Progress), Message: "operation cancelled", CreatedAt: now})
			s.audit(ctx, operation, "completed", "cancelled", nil)
			return
		}
		var operationErr *cephdomain.OperationError
		if errors.As(err, &operationErr) {
			if operationErr.Code == "post_check_failed" {
				s.needsReview(ctx, operation, operationErr.Code, operationErr.Message)
			} else {
				s.fail(ctx, operation, operationErr.Code, operationErr.Message, operationErr.Retryable)
			}
		} else {
			s.fail(ctx, operation, "operation_failed", security.Redact(err.Error()), false)
		}
		return
	}
	s.handlersMu.RLock()
	postSuccess := s.postSuccess
	s.handlersMu.RUnlock()
	if postSuccess != nil {
		if err := postSuccess(operationCtx, operation); err != nil {
			s.needsReview(ctx, operation, "post_reconcile_failed", security.Redact(err.Error()))
			return
		}
	}
	encoded, _ := safeJSON(result)
	now := time.Now().UTC()
	_ = s.database().UpdateOperation(ctx, operation.ID, map[string]any{"status": "succeeded", "stage": "completed", "progress": 100, "result_json": string(encoded), "completed_at": now})
	_ = s.database().AppendOperationEvent(ctx, &store.CephOperationEvent{OperationID: operation.ID, EventType: "state", Stage: "completed", Progress: intPointer(100), Message: "operation succeeded", CreatedAt: now})
	s.audit(ctx, operation, "completed", "succeeded", nil)
}

func (s *Service) precheckPlan(ctx context.Context, operation store.CephOperation) error {
	if operation.Risk != string(cephdomain.RiskHigh) || operation.ClusterID == nil {
		return nil
	}
	s.handlersMu.RLock()
	checker := s.planChecker
	s.handlersMu.RUnlock()
	if checker == nil {
		return &cephdomain.OperationError{Code: "pre_check_unavailable", Message: "high-risk pre-check is unavailable"}
	}
	decoded, err := s.decodeRequestJSON(operation.RequestJSON)
	if err != nil {
		return fmt.Errorf("decode operation request for pre-check: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.UseNumber()
	var parameters any
	if err := decoder.Decode(&parameters); err != nil {
		return fmt.Errorf("decode operation parameters for pre-check: %w", err)
	}
	generation := uint64(0)
	if operation.ResourceGeneration != nil {
		generation = *operation.ResourceGeneration
	}
	blockers, warnings, err := checker(ctx, PlanRequest{ClusterID: *operation.ClusterID, ActorUserID: operation.ActorUserID, ActorUsername: operation.ActorUsername, RequestID: operation.RequestID, Action: operation.Action, ResourceKind: operation.ResourceKind, ResourceKey: operation.ResourceKey, ResourceGeneration: generation, Risk: cephdomain.RiskHigh, Parameters: parameters})
	if err != nil {
		return err
	}
	if len(warnings) > 0 {
		data, _ := safeJSON(map[string]any{"warnings": warnings})
		encoded := string(data)
		_ = s.database().AppendOperationEvent(ctx, &store.CephOperationEvent{OperationID: operation.ID, EventType: "pre_check", Stage: "pre_check", Message: "pre-check completed with warnings", DataJSON: &encoded, CreatedAt: time.Now().UTC()})
	}
	if len(blockers) > 0 {
		return &cephdomain.OperationError{Code: "precondition_blocked", Message: strings.Join(blockers, "; ")}
	}
	return nil
}

func (s *Service) precheckGeneration(ctx context.Context, operation store.CephOperation) error {
	if operation.ClusterID == nil || operation.ResourceGeneration == nil {
		return nil
	}
	key := resourceLookupKey(operation.ResourceKind, operation.ResourceKey)
	row, err := s.database().FindResource(ctx, *operation.ClusterID, operation.ResourceKind, key)
	if errors.Is(err, store.ErrRecordNotFound) {
		if *operation.ResourceGeneration == 0 {
			return nil
		}
		return fmt.Errorf("resource generation changed")
	}
	if err != nil {
		return err
	}
	if row.ResourceVersion != *operation.ResourceGeneration {
		return fmt.Errorf("resource generation changed")
	}
	return nil
}

func (s *Service) heartbeat(ctx context.Context, cancel context.CancelFunc, operationID string, done chan<- struct{}) {
	defer close(done)
	interval := s.lease / 3
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			now = now.UTC()
			if err := s.database().RenewOperationLease(ctx, operationID, now, now.Add(s.lease)); err != nil {
				cancel()
				return
			}
		}
	}
}
func (s *Service) fail(ctx context.Context, operation store.CephOperation, code, message string, retryable bool) {
	now := time.Now().UTC()
	_ = s.database().UpdateOperation(ctx, operation.ID, map[string]any{"status": "failed", "stage": "failed", "error_code": code, "error_message": security.Redact(message), "retryable": retryable, "completed_at": now})
	_ = s.database().AppendOperationEvent(ctx, &store.CephOperationEvent{OperationID: operation.ID, EventType: "error", Stage: "failed", Message: security.Redact(message), ErrorCode: &code, CreatedAt: now})
	s.audit(ctx, operation, "completed", "failed", &code)
}
func (s *Service) needsReview(ctx context.Context, operation store.CephOperation, code, message string) {
	now := time.Now().UTC()
	_ = s.database().UpdateOperation(ctx, operation.ID, map[string]any{"status": "needs_review", "stage": "needs_review", "error_code": code, "error_message": security.Redact(message), "retryable": false, "completed_at": now})
	_ = s.database().AppendOperationEvent(ctx, &store.CephOperationEvent{OperationID: operation.ID, EventType: "warning", Stage: "needs_review", Message: security.Redact(message), ErrorCode: &code, CreatedAt: now})
	s.audit(ctx, operation, "completed", "needs_review", &code)
}
func (s *Service) lockRows(operation store.CephOperation) []store.CephOperationLock {
	s.locksMu.Lock()
	keys := s.locks[operation.ID]
	delete(s.locks, operation.ID)
	s.locksMu.Unlock()
	if len(keys) == 0 {
		keys = []LockKey{{Kind: operation.ResourceKind, Key: operation.ResourceKey}}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Kind+":"+keys[i].Key < keys[j].Kind+":"+keys[j].Key })
	rows := make([]store.CephOperationLock, 0, len(keys))
	for _, key := range keys {
		canonical := fmt.Sprintf("cluster:%d:%s:%s", *operation.ClusterID, key.Kind, key.Key)
		rows = append(rows, store.CephOperationLock{LockKey: store.SHA256(canonical), ResourceKind: key.Kind, ResourceKey: key.Key, LeaseExpiresAt: time.Now().UTC().Add(s.lease)})
	}
	return rows
}
func (s *Service) Get(ctx context.Context, id string) (store.CephOperation, error) {
	return s.database().FindOperation(ctx, id)
}
func (s *Service) List(ctx context.Context, clusterID uint64, status, action string, limit int) ([]store.CephOperation, error) {
	return s.database().ListOperations(ctx, clusterID, status, action, limit)
}
func (s *Service) ListFiltered(ctx context.Context, clusterID uint64, filter store.OperationFilter) ([]store.CephOperation, error) {
	return s.database().ListOperationsFiltered(ctx, clusterID, filter)
}
func (s *Service) Events(ctx context.Context, id string) ([]store.CephOperationEvent, error) {
	return s.database().ListOperationEvents(ctx, id)
}
func (s *Service) Cancel(ctx context.Context, id string) error {
	operation, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	switch operation.Status {
	case "queued":
		now := time.Now().UTC()
		return s.database().UpdateOperation(ctx, id, map[string]any{"status": "cancelled", "stage": "cancelled", "completed_at": now})
	case "running":
		if err := s.database().UpdateOperation(ctx, id, map[string]any{"status": "cancel_requested", "cancel_requested_at": time.Now().UTC()}); err != nil {
			return err
		}
		s.activeMu.Lock()
		cancel := s.active[id]
		s.activeMu.Unlock()
		if cancel != nil {
			cancel()
		}
		return nil
	default:
		return fmt.Errorf("operation cannot be cancelled from status %s", operation.Status)
	}
}

func (s *Service) audit(ctx context.Context, operation store.CephOperation, eventType, outcome string, errorCode *string) {
	now := time.Now().UTC()
	resourceKind, resourceKey := operation.ResourceKind, operation.ResourceKey
	event := store.AuditEvent{OccurredAt: now, EventType: eventType, RequestID: operation.RequestID, ActorUserID: operation.ActorUserID, ActorUsername: operation.ActorUsername, ClusterID: operation.ClusterID, ClusterName: stringPointer(operation.ClusterName), Action: operation.Action, ResourceKind: &resourceKind, ResourceKey: &resourceKey, Risk: stringPointer(operation.Risk), Outcome: outcome, ErrorCode: errorCode, OperationID: &operation.ID, EventHash: store.SHA256(operation.ID + ":" + eventType + ":" + outcome + ":" + now.Format(time.RFC3339Nano))}
	_ = s.database().CreateAuditEvent(ctx, &event)
}

func safeJSON(value any) ([]byte, error) {
	redacted, err := security.RedactJSON(value)
	if err != nil {
		return nil, err
	}
	return json.Marshal(redacted)
}
func (s *Service) requestJSON(value any) ([]byte, error) {
	if s.encryptionKey == "" {
		return safeJSON(value)
	}
	protected, err := security.ProtectJSON(value, s.encryptionKey)
	if err != nil {
		return nil, err
	}
	return json.Marshal(protected)
}
func (s *Service) decodeRequestJSON(value string) ([]byte, error) {
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()
	var stored any
	if err := decoder.Decode(&stored); err != nil {
		return nil, err
	}
	if s.encryptionKey != "" {
		var err error
		stored, err = security.UnprotectJSON(stored, s.encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(stored)
}
func normalizeJSONBytes(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}
func idempotencySubject(value *uint64) string {
	if value == nil {
		return "none"
	}
	return fmt.Sprintf("%d", *value)
}
func resourceLookupKey(kind, resourceKey string) string {
	path := strings.Split(strings.Trim(resourceKey, "/"), "/")
	after := func(segment string) string {
		for index := 0; index+1 < len(path); index++ {
			if path[index] == segment {
				return path[index+1]
			}
		}
		return ""
	}
	lastResource := func() string {
		for index := len(path) - 1; index >= 0; index-- {
			switch path[index] {
			case "action", "delete", "purge", "zap", "commit", "clone":
				continue
			default:
				return path[index]
			}
		}
		return ""
	}
	switch kind {
	case "device":
		encoded := after("device")
		decoded, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
		if err == nil {
			parts := bytes.SplitN(decoded, []byte{0}, 2)
			if len(parts) == 2 {
				return string(parts[0]) + ":" + string(parts[1])
			}
		}
		return encoded
	case "rbd_image":
		return after("image")
	case "rbd_snapshot":
		return after("image") + "@" + after("snapshot")
	case "rbd_namespace":
		return after("namespace") + "/" + lastResource()
	case "subvolume_group":
		return after("filesystem") + "/" + after("subvolume-group")
	case "subvolume":
		return after("filesystem") + "/" + after("subvolume")
	case "cephfs_snapshot":
		return after("filesystem") + "/" + after("subvolume") + "/" + after("snapshot")
	case "filesystem":
		return after("filesystem")
	case "osd":
		return after("osd")
	case "pool":
		return after("pool")
	case "host":
		return after("host")
	case "service":
		return after("service")
	case "crush_rule":
		return after("crush-rule")
	case "erasure_code_profile":
		return after("erasure-code-profile")
	case "nfs_cluster", "smb_cluster":
		return after("cluster")
	case "nfs_export":
		return after("export")
	case "smb_share":
		return after("share")
	case "rgw_bucket":
		return after("bucket")
	case "nvmeof_subsystem":
		return after("subsystem")
	case "nvmeof_namespace":
		return after("namespace")
	case "nvmeof_listener":
		return after("listener")
	case "nvmeof_host":
		return after("host")
	case "iscsi_target":
		return after("target")
	case "upgrade":
		return "upgrade"
	default:
		return lastResource()
	}
}
func newUUID() string {
	data := make([]byte, 16)
	_, _ = rand.Read(data)
	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", data[:4], data[4:6], data[6:8], data[8:10], data[10:])
}
func intPointer(v int) *int          { return &v }
func stringPointer(v string) *string { return &v }
