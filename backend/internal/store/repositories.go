package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *Database) ListClusters(ctx context.Context) ([]CephCluster, error) {
	var rows []CephCluster
	return rows, d.db.WithContext(ctx).Order("id asc").Find(&rows).Error
}
func (d *Database) FindCluster(ctx context.Context, id uint64) (CephCluster, error) {
	var row CephCluster
	return row, d.db.WithContext(ctx).First(&row, id).Error
}
func (d *Database) CreateCluster(ctx context.Context, row *CephCluster) error {
	return d.db.WithContext(ctx).Create(row).Error
}
func (d *Database) SaveCluster(ctx context.Context, row *CephCluster) error {
	return d.db.WithContext(ctx).Save(row).Error
}
func (d *Database) DeleteCluster(ctx context.Context, row *CephCluster) error {
	return d.db.WithContext(ctx).Delete(row).Error
}

func (d *Database) ListCapabilities(ctx context.Context, clusterID uint64) ([]CephClusterCapability, error) {
	var rows []CephClusterCapability
	return rows, d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order("name asc").Find(&rows).Error
}

func (d *Database) ListCredentials(ctx context.Context, clusterID uint64) ([]CephClusterCredential, error) {
	var rows []CephClusterCredential
	return rows, d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order("kind asc").Find(&rows).Error
}
func (d *Database) FindCredential(ctx context.Context, clusterID uint64, kind string) (CephClusterCredential, error) {
	var row CephClusterCredential
	return row, d.db.WithContext(ctx).Where("cluster_id = ? AND kind = ?", clusterID, kind).First(&row).Error
}
func (d *Database) UpsertCredential(ctx context.Context, row *CephClusterCredential) error {
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "cluster_id"}, {Name: "kind"}}, DoUpdates: clause.AssignmentColumns([]string{"credential", "fingerprint", "updated_at"})}).Create(row).Error
}
func (d *Database) DeleteCredential(ctx context.Context, clusterID uint64, kind string) error {
	result := d.db.WithContext(ctx).Where("cluster_id = ? AND kind = ?", clusterID, kind).Delete(&CephClusterCredential{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (d *Database) ListEndpoints(ctx context.Context, clusterID uint64) ([]CephClusterEndpoint, error) {
	var rows []CephClusterEndpoint
	return rows, d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order("kind asc, name asc").Find(&rows).Error
}
func (d *Database) FindEndpoint(ctx context.Context, clusterID, endpointID uint64) (CephClusterEndpoint, error) {
	var row CephClusterEndpoint
	return row, d.db.WithContext(ctx).Where("cluster_id = ? AND id = ?", clusterID, endpointID).First(&row).Error
}
func (d *Database) CreateEndpoint(ctx context.Context, row *CephClusterEndpoint) error {
	return d.db.WithContext(ctx).Create(row).Error
}
func (d *Database) SaveEndpoint(ctx context.Context, row *CephClusterEndpoint) error {
	return d.db.WithContext(ctx).Save(row).Error
}
func (d *Database) DeleteEndpoint(ctx context.Context, row *CephClusterEndpoint) error {
	return d.db.WithContext(ctx).Delete(row).Error
}

type ResourceFilter struct {
	Kind, ParentKind, ParentKey, Name, Status string
	Limit                                     int
	AfterID                                   uint64
}

type OperationFilter struct {
	Status, Action, ResourceKind, ResourceKey string
	ActorUserID                               *uint64
	Limit                                     int
}

type AuditFilter struct {
	ActorUsername, Action, ResourceKind, ResourceKey string
	ActorUserID                                      *uint64
	Limit                                            int
}

func (d *Database) ListResources(ctx context.Context, clusterID uint64, filter ResourceFilter) ([]CephResourceRecord, error) {
	query := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
	for column, value := range map[string]string{"kind": filter.Kind, "parent_kind": filter.ParentKind, "parent_key": filter.ParentKey, "name": filter.Name, "status": filter.Status} {
		if value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	if filter.AfterID > 0 {
		query = query.Where("id > ?", filter.AfterID)
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	var rows []CephResourceRecord
	return rows, query.Order("id asc").Limit(limit + 1).Find(&rows).Error
}
func (d *Database) FindResource(ctx context.Context, clusterID uint64, kind, key string) (CephResourceRecord, error) {
	var row CephResourceRecord
	return row, d.db.WithContext(ctx).Where("cluster_id = ? AND kind = ? AND natural_key = ?", clusterID, kind, key).First(&row).Error
}
func (d *Database) UpsertResources(ctx context.Context, rows []CephResourceRecord) error {
	if len(rows) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cluster_id"}, {Name: "kind"}, {Name: "natural_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"parent_kind", "parent_key", "name", "status", "generation", "resource_version", "source", "source_version", "observed_at", "stale_at", "payload_schema_version", "payload_json", "updated_at"}),
	}).Create(&rows).Error
}

func (d *Database) CreateOperation(ctx context.Context, operation *CephOperation, event *CephOperationEvent, audit *AuditEvent) error {
	d.auditMu.Lock()
	defer d.auditMu.Unlock()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(operation).Error; err != nil {
			return err
		}
		if event != nil {
			event.OperationID = operation.ID
			if err := tx.Create(event).Error; err != nil {
				return err
			}
		}
		if audit != nil {
			audit.OperationID = &operation.ID
			if err := createAuditEvent(tx, audit); err != nil {
				return err
			}
		}
		return nil
	})
}
func (d *Database) FindOperation(ctx context.Context, id string) (CephOperation, error) {
	var row CephOperation
	return row, d.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
}
func (d *Database) FindIdempotentOperation(ctx context.Context, scopeHash string) (CephOperation, error) {
	var row CephOperation
	return row, d.db.WithContext(ctx).Where("idempotency_scope_hash = ?", scopeHash).First(&row).Error
}
func (d *Database) ListOperations(ctx context.Context, clusterID uint64, status, action string, limit int) ([]CephOperation, error) {
	return d.ListOperationsFiltered(ctx, clusterID, OperationFilter{Status: status, Action: action, Limit: limit})
}
func (d *Database) ListOperationsFiltered(ctx context.Context, clusterID uint64, filter OperationFilter) ([]CephOperation, error) {
	q := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
	for column, value := range map[string]string{"status": filter.Status, "action": filter.Action, "resource_kind": filter.ResourceKind, "resource_key": filter.ResourceKey} {
		if value != "" {
			q = q.Where(column+" = ?", value)
		}
	}
	if filter.ActorUserID != nil {
		q = q.Where("actor_user_id = ?", *filter.ActorUserID)
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	var rows []CephOperation
	return rows, q.Order("created_at desc").Limit(limit).Find(&rows).Error
}
func (d *Database) CountNonTerminalOperations(ctx context.Context, clusterID uint64, excludeID string) (int64, error) {
	query := d.db.WithContext(ctx).Model(&CephOperation{}).Where("cluster_id = ? AND status NOT IN ?", clusterID, []string{"succeeded", "failed", "cancelled", "needs_review"})
	if excludeID != "" {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}
func (d *Database) ListOperationEvents(ctx context.Context, operationID string) ([]CephOperationEvent, error) {
	var rows []CephOperationEvent
	return rows, d.db.WithContext(ctx).Where("operation_id = ?", operationID).Order("sequence asc").Find(&rows).Error
}

func (d *Database) ListClusterOperationEventsAfter(ctx context.Context, clusterID, afterID uint64, limit int) ([]CephOperationEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows []CephOperationEvent
	query := d.db.WithContext(ctx).Table("ceph_operation_event AS event").Select("event.*").Joins("JOIN ceph_operation AS operation ON operation.id = event.operation_id").Where("operation.cluster_id = ? AND event.id > ?", clusterID, afterID).Order("event.id asc").Limit(limit)
	return rows, query.Scan(&rows).Error
}

func (d *Database) ClaimQueuedOperation(ctx context.Context, now time.Time, perClusterLimit int) (CephOperation, error) {
	if perClusterLimit <= 0 {
		perClusterLimit = 2
	}
	var claimed CephOperation
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var candidate CephOperation
		query := tx.Table("ceph_operation AS candidate").Select("candidate.*").
			Where("candidate.status = ? AND candidate.scheduled_at <= ?", "queued", now).
			Where("candidate.cluster_id IS NULL OR (SELECT COUNT(*) FROM ceph_operation AS active WHERE active.cluster_id = candidate.cluster_id AND active.status IN ?) < ?", []string{"running", "cancel_requested", "recovering"}, perClusterLimit).
			Order("candidate.scheduled_at asc, candidate.created_at asc")
		if err := query.Take(&candidate).Error; err != nil {
			return err
		}
		result := tx.Model(&CephOperation{}).Where("id = ? AND status = ?", candidate.ID, "queued").Updates(map[string]any{"status": "running", "stage": "pre_check", "started_at": now, "heartbeat_at": now, "attempt": gorm.Expr("attempt + 1"), "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("id = ?", candidate.ID).First(&claimed).Error
	})
	return claimed, err
}
func (d *Database) UpdateOperation(ctx context.Context, id string, updates map[string]any) error {
	updates["updated_at"] = time.Now().UTC()
	return d.db.WithContext(ctx).Model(&CephOperation{}).Where("id = ?", id).Updates(updates).Error
}
func (d *Database) AppendOperationEvent(ctx context.Context, event *CephOperationEvent) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var max *uint64
		if err := tx.Model(&CephOperationEvent{}).Where("operation_id = ?", event.OperationID).Select("MAX(sequence)").Scan(&max).Error; err != nil {
			return err
		}
		event.Sequence = 1
		if max != nil {
			event.Sequence = *max + 1
		}
		return tx.Create(event).Error
	})
}
func (d *Database) RecoverOperations(ctx context.Context, staleBefore, timeNow time.Time) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stale := tx.Model(&CephOperation{}).Where("status IN ? AND (heartbeat_at IS NULL OR heartbeat_at < ?)", []string{"running", "recovering", "cancel_requested"}, staleBefore)
		if err := stale.Where("risk = ?", "low").Updates(map[string]any{"status": "queued", "stage": "recovering", "scheduled_at": timeNow, "started_at": nil, "heartbeat_at": nil, "updated_at": timeNow}).Error; err != nil {
			return err
		}
		return tx.Model(&CephOperation{}).Where("status IN ? AND risk <> ? AND (heartbeat_at IS NULL OR heartbeat_at < ?)", []string{"running", "recovering", "cancel_requested"}, "low", staleBefore).Updates(map[string]any{"status": "needs_review", "stage": "recovery", "error_code": "worker_interrupted", "error_message": "worker stopped before completion", "retryable": true, "completed_at": timeNow, "updated_at": timeNow}).Error
	})
}

func (d *Database) RenewOperationLease(ctx context.Context, operationID string, now, leaseExpiresAt time.Time) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&CephOperation{}).Where("id = ? AND status IN ?", operationID, []string{"running", "cancel_requested"}).Updates(map[string]any{"heartbeat_at": now, "updated_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Model(&CephOperationLock{}).Where("operation_id = ?", operationID).Updates(map[string]any{"lease_expires_at": leaseExpiresAt, "updated_at": now}).Error
	})
}

func (d *Database) AcquireLocks(ctx context.Context, operation CephOperation, keys []CephOperationLock, now time.Time) error {
	sort.Slice(keys, func(i, j int) bool { return keys[i].LockKey < keys[j].LockKey })
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range keys {
			keys[i].OperationID = operation.ID
			keys[i].ClusterID = *operation.ClusterID
			keys[i].AcquiredAt = now
			keys[i].UpdatedAt = now
			takeover := tx.Model(&CephOperationLock{}).Where("lock_key = ? AND lease_expires_at <= ?", keys[i].LockKey, now).Updates(map[string]any{"cluster_id": keys[i].ClusterID, "resource_kind": keys[i].ResourceKind, "resource_key": keys[i].ResourceKey, "operation_id": keys[i].OperationID, "fencing_token": gorm.Expr("fencing_token + 1"), "lease_expires_at": keys[i].LeaseExpiresAt, "acquired_at": now, "updated_at": now})
			if takeover.Error != nil {
				return takeover.Error
			}
			if takeover.RowsAffected == 1 {
				continue
			}
			keys[i].FencingToken = 1
			if err := tx.Create(&keys[i]).Error; err != nil {
				return errors.New("resource is locked")
			}
		}
		return nil
	})
}
func (d *Database) ReleaseLocks(ctx context.Context, operationID string) error {
	return d.db.WithContext(ctx).Where("operation_id = ?", operationID).Delete(&CephOperationLock{}).Error
}

func (d *Database) CreatePlan(ctx context.Context, plan *CephActionPlan) error {
	return d.db.WithContext(ctx).Create(plan).Error
}
func (d *Database) FindPlan(ctx context.Context, id string) (CephActionPlan, error) {
	var row CephActionPlan
	return row, d.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
}
func (d *Database) ConsumePlan(ctx context.Context, id string, actorUserID *uint64, clusterID uint64, action, resourceKind, resourceKey string, generation uint64, now time.Time) error {
	result := d.db.WithContext(ctx).Model(&CephActionPlan{}).Where("id = ? AND status = ? AND risk = ? AND expires_at > ? AND cluster_id = ? AND action = ? AND resource_kind = ? AND resource_key = ? AND resource_generation = ?", id, "valid", "high", now, clusterID, action, resourceKind, resourceKey, generation)
	if actorUserID == nil {
		result = result.Where("actor_user_id IS NULL")
	} else {
		result = result.Where("actor_user_id = ?", *actorUserID)
	}
	result = result.Updates(map[string]any{"status": "consumed", "consumed_at": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *Database) UpsertObservation(ctx context.Context, row *CephClusterObservation) error {
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "cluster_id"}}, DoUpdates: clause.AssignmentColumns([]string{"fsid", "ceph_version", "status", "enabled", "generation", "last_seen_at", "last_error_code", "last_error_message", "observed_at", "updated_at"})}).Create(row).Error
}
func (d *Database) UpsertCapabilities(ctx context.Context, rows []CephClusterCapability) error {
	if len(rows) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "cluster_id"}, {Name: "name"}}, DoUpdates: clause.AssignmentColumns([]string{"supported", "reason", "version", "details_json", "observed_at", "updated_at"})}).Create(&rows).Error
}
func (d *Database) FindObservation(ctx context.Context, clusterID uint64) (CephClusterObservation, error) {
	var row CephClusterObservation
	return row, d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).First(&row).Error
}

func (d *Database) NextCollectionGeneration(ctx context.Context, clusterID uint64, module string) (uint64, error) {
	var value *uint64
	err := d.db.WithContext(ctx).Model(&CephCollectionRun{}).Where("cluster_id = ? AND module = ?", clusterID, module).Select("MAX(generation)").Scan(&value).Error
	if value == nil {
		return 1, err
	}
	return *value + 1, err
}
func (d *Database) CreateCollectionRun(ctx context.Context, row *CephCollectionRun) error {
	return d.db.WithContext(ctx).Create(row).Error
}
func (d *Database) FinishCollectionRun(ctx context.Context, id uint64, status string, count uint64, errorCode, errorMessage *string, finished time.Time) error {
	return d.db.WithContext(ctx).Model(&CephCollectionRun{}).Where("id = ?", id).Updates(map[string]any{"status": status, "record_count": count, "error_code": errorCode, "error_message": errorMessage, "finished_at": finished}).Error
}
func (d *Database) ReconcileResources(ctx context.Context, clusterID, generation uint64, rows []CephResourceRecord, authoritativeKinds []string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for i := range rows {
			row := &rows[i]
			row.ClusterID, row.Generation, row.UpdatedAt = clusterID, generation, now
			var existing CephResourceRecord
			err := tx.Where("cluster_id = ? AND kind = ? AND natural_key = ?", clusterID, row.Kind, row.NaturalKey).First(&existing).Error
			if err == nil {
				row.ID, row.CreatedAt, row.ResourceVersion = existing.ID, existing.CreatedAt, existing.ResourceVersion
				if existing.PayloadJSON != row.PayloadJSON || !equalStringPointer(existing.Status, row.Status) {
					row.ResourceVersion++
				}
				if err := tx.Save(row).Error; err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				row.ResourceVersion, row.CreatedAt = 1, now
				if err := tx.Create(row).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		for _, kind := range authoritativeKinds {
			if err := tx.Model(&CephResourceRecord{}).Where("cluster_id = ? AND kind = ? AND generation <> ? AND stale_at IS NULL", clusterID, kind, generation).Update("stale_at", now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (d *Database) MarkModuleResourcesStale(ctx context.Context, clusterID uint64, kinds []string, now time.Time) error {
	if len(kinds) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Model(&CephResourceRecord{}).Where("cluster_id = ? AND kind IN ? AND stale_at IS NULL", clusterID, kinds).Update("stale_at", now).Error
}
func equalStringPointer(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func SHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (d *Database) EnsureSettings(ctx context.Context, settings []Setting) error {
	if len(settings) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&settings).Error
}
func (d *Database) ListSettings(ctx context.Context, prefix string) ([]Setting, error) {
	q := d.db.WithContext(ctx)
	if prefix != "" {
		q = q.Where("key LIKE ?", prefix+"%")
	}
	var rows []Setting
	return rows, q.Order("key asc").Find(&rows).Error
}
func (d *Database) FindSetting(ctx context.Context, key string) (Setting, error) {
	var row Setting
	return row, d.db.WithContext(ctx).Where("key = ?", key).First(&row).Error
}
func (d *Database) UpsertSetting(ctx context.Context, key, value string) (Setting, error) {
	now := time.Now().UTC()
	row := Setting{Key: key, Value: value, CreatedAt: now, UpdatedAt: now}
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoUpdates: clause.Assignments(map[string]any{"value": value, "updated_at": now})}).Create(&row).Error
	return row, err
}
func (d *Database) DeleteSettingsByPrefix(ctx context.Context, prefix string) error {
	return d.db.WithContext(ctx).Where("key LIKE ?", prefix+"%").Delete(&Setting{}).Error
}

func (d *Database) HasUsers(ctx context.Context) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&User{}).Count(&count).Error
	return count > 0, err
}

func (d *Database) FindUserByUsername(ctx context.Context, username string) (User, error) {
	var row User
	return row, d.db.WithContext(ctx).Where("username = ?", username).First(&row).Error
}
func (d *Database) FindUserByAccount(ctx context.Context, account string) (User, error) {
	var row User
	return row, d.db.WithContext(ctx).Where("username = ? OR email = ?", account, account).First(&row).Error
}
func (d *Database) FindUserByID(ctx context.Context, id uint64) (User, error) {
	var row User
	return row, d.db.WithContext(ctx).First(&row, id).Error
}
func (d *Database) ListUsers(ctx context.Context) ([]User, error) {
	var rows []User
	return rows, d.db.WithContext(ctx).Order("id asc").Find(&rows).Error
}
func (d *Database) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&User{}).Count(&count).Error
	return count, err
}
func (d *Database) CreateUser(ctx context.Context, row *User) error {
	return d.db.WithContext(ctx).Create(row).Error
}
func (d *Database) UpdateUser(ctx context.Context, id uint64, updates map[string]any) error {
	updates["updated_at"] = time.Now().UTC()
	return d.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updates).Error
}
func (d *Database) DeleteUserSessions(ctx context.Context, userID uint64) error {
	return d.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserSession{}).Error
}
func (d *Database) CreateSessionAndTouchUser(ctx context.Context, session *UserSession, now time.Time) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(session).Error; err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", session.UserID).Updates(map[string]any{"last_login_at": now, "updated_at": now}).Error
	})
}
func (d *Database) UserForSessionHash(ctx context.Context, tokenHash string, now time.Time) (User, error) {
	var session UserSession
	err := d.db.WithContext(ctx).Preload("User").Where("token_hash = ? AND expires_at > ? AND revoked_at IS NULL", tokenHash, now).First(&session).Error
	return session.User, err
}
func (d *Database) ReplacePasswordReset(ctx context.Context, reset *PasswordResetCode) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&PasswordResetCode{}).Where("user_id = ? AND consumed_at IS NULL", reset.UserID).Update("consumed_at", time.Now().UTC()).Error; err != nil {
			return err
		}
		return tx.Create(reset).Error
	})
}
func (d *Database) FindValidPasswordReset(ctx context.Context, userID uint64, codeHash string, now time.Time) (PasswordResetCode, error) {
	var row PasswordResetCode
	return row, d.db.WithContext(ctx).Where("user_id = ? AND code_hash = ? AND consumed_at IS NULL AND expires_at > ?", userID, codeHash, now).Order("id desc").First(&row).Error
}
func (d *Database) CompletePasswordReset(ctx context.Context, userID, resetID uint64, password string, now time.Time) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).Where("id = ?", userID).Updates(map[string]any{"password": password, "status": "active", "updated_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Model(&PasswordResetCode{}).Where("id = ? AND consumed_at IS NULL", resetID).Update("consumed_at", now).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&UserSession{}).Error
	})
}

func (d *Database) EnsureBuiltinRoles(ctx context.Context, roles []string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for _, name := range roles {
			role := Role{Name: name, CreatedAt: now, UpdatedAt: now}
			if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&role).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (d *Database) BindUserRole(ctx context.Context, userID uint64, roleName string, clusterID *uint64, createdBy *uint64) error {
	var role Role
	if err := d.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}
	scope := "*"
	if clusterID != nil {
		scope = fmtUint(*clusterID)
	}
	row := UserRoleBinding{UserID: userID, RoleID: role.ID, ClusterID: clusterID, ScopeKey: scope, CreatedByUserID: createdBy, CreatedAt: time.Now().UTC()}
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}
func (d *Database) ListClusterRoleBindings(ctx context.Context, clusterID uint64) ([]UserRoleBinding, error) {
	var rows []UserRoleBinding
	err := d.db.WithContext(ctx).Preload("User").Preload("Role").Where("cluster_id = ?", clusterID).Order("id asc").Find(&rows).Error
	return rows, err
}
func (d *Database) DeleteClusterRoleBinding(ctx context.Context, clusterID, bindingID uint64) error {
	result := d.db.WithContext(ctx).Where("id = ? AND cluster_id = ?", bindingID, clusterID).Delete(&UserRoleBinding{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (d *Database) ListRoles(ctx context.Context) ([]Role, error) {
	var rows []Role
	return rows, d.db.WithContext(ctx).Order("name asc").Find(&rows).Error
}
func (d *Database) CreateRole(ctx context.Context, role *Role) error {
	return d.db.WithContext(ctx).Create(role).Error
}
func (d *Database) ListAuditEvents(ctx context.Context, clusterID uint64, limit int) ([]AuditEvent, error) {
	return d.ListAuditEventsFiltered(ctx, clusterID, AuditFilter{Limit: limit})
}
func (d *Database) ListAuditEventsFiltered(ctx context.Context, clusterID uint64, filter AuditFilter) ([]AuditEvent, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	query := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
	for column, value := range map[string]string{"actor_username": filter.ActorUsername, "action": filter.Action, "resource_kind": filter.ResourceKind, "resource_key": filter.ResourceKey} {
		if value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	if filter.ActorUserID != nil {
		query = query.Where("actor_user_id = ?", *filter.ActorUserID)
	}
	var rows []AuditEvent
	return rows, query.Order("occurred_at desc").Limit(limit).Find(&rows).Error
}
func (d *Database) CreateAuditEvent(ctx context.Context, row *AuditEvent) error {
	d.auditMu.Lock()
	defer d.auditMu.Unlock()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return createAuditEvent(tx, row) })
}
func createAuditEvent(tx *gorm.DB, row *AuditEvent) error {
	const headKey = "system.audit_chain_head"
	now := time.Now().UTC()
	head := Setting{Key: headKey, Value: "", CreatedAt: now, UpdatedAt: now}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&head).Error; err != nil {
		return err
	}
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("key = ?", headKey).First(&head).Error; err != nil {
		return err
	}
	if head.Value == "" {
		var previous AuditEvent
		err := tx.Order("id desc").First(&previous).Error
		if err == nil {
			head.Value = previous.EventHash
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	if head.Value != "" {
		previousHash := head.Value
		row.PreviousHash = &previousHash
	}
	payload, err := json.Marshal(struct {
		ID                                                                                                uint64
		OccurredAt                                                                                        time.Time `json:"occurred_at"`
		EventType, RequestID, ActorUsername, Action, Outcome                                              string
		ActorUserID, ClusterID, BeforeGeneration, AfterGeneration                                         *uint64
		HTTPStatus                                                                                        *int
		ClusterName, ResourceKind, ResourceKey, Risk, ErrorCode, SourceIP, UserAgent, PlanID, OperationID *string
		ParametersJSON, DetailsJSON, PreviousHash                                                         *string
	}{row.ID, row.OccurredAt.UTC(), row.EventType, row.RequestID, row.ActorUsername, row.Action, row.Outcome, row.ActorUserID, row.ClusterID, row.BeforeGeneration, row.AfterGeneration, row.HTTPStatus, row.ClusterName, row.ResourceKind, row.ResourceKey, row.Risk, row.ErrorCode, row.SourceIP, row.UserAgent, row.PlanID, row.OperationID, row.ParametersJSON, row.DetailsJSON, row.PreviousHash})
	if err != nil {
		return err
	}
	row.EventHash = SHA256(string(payload))
	if err := tx.Create(row).Error; err != nil {
		return err
	}
	return tx.Model(&Setting{}).Where("key = ?", headKey).Updates(map[string]any{"value": row.EventHash, "updated_at": now}).Error
}
func fmtUint(value uint64) string { return strconv.FormatUint(value, 10) }
