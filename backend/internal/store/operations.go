package store

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	OperationQueued    = "queued"
	OperationRunning   = "running"
	OperationSucceeded = "succeeded"
	OperationFailed    = "failed"
)

type OperationFilter struct {
	ClusterID uint64
	Status    string
	Limit     int
}

func (d *Database) CreateOperation(ctx context.Context, row *CephOperation) error {
	return d.db.WithContext(ctx).Create(row).Error
}

func (d *Database) FindOperation(ctx context.Context, id uint64) (CephOperation, error) {
	var row CephOperation
	return row, d.db.WithContext(ctx).First(&row, id).Error
}

func (d *Database) FindOperationByIdempotencyKey(ctx context.Context, clusterID uint64, key string) (CephOperation, error) {
	var row CephOperation
	return row, d.db.WithContext(ctx).Where("cluster_id = ? AND idempotency_key = ?", clusterID, key).First(&row).Error
}

func (d *Database) ListOperations(ctx context.Context, filter OperationFilter) ([]CephOperation, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	query := d.db.WithContext(ctx)
	if filter.ClusterID != 0 {
		query = query.Where("cluster_id = ?", filter.ClusterID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var rows []CephOperation
	return rows, query.Order("id desc").Limit(limit).Find(&rows).Error
}

func (d *Database) ClaimNextOperation(ctx context.Context, now time.Time) (CephOperation, error) {
	var claimed CephOperation
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ? AND attempts < max_attempts AND (next_attempt_at IS NULL OR next_attempt_at <= ?)", OperationQueued, now).
			Order("id asc")
		var row CephOperation
		if err := query.First(&row).Error; err != nil {
			return err
		}
		result := tx.Model(&CephOperation{}).
			Where("id = ? AND status = ?", row.ID, OperationQueued).
			Updates(map[string]any{
				"status":          OperationRunning,
				"attempts":        gorm.Expr("attempts + 1"),
				"started_at":      now,
				"next_attempt_at": nil,
				"updated_at":      now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.First(&claimed, row.ID).Error
	})
	return claimed, err
}

func (d *Database) CompleteOperation(ctx context.Context, id uint64, resultJSON string, finishedAt time.Time) error {
	result := d.db.WithContext(ctx).Model(&CephOperation{}).
		Where("id = ? AND status = ?", id, OperationRunning).
		Updates(map[string]any{
			"status":        OperationSucceeded,
			"result_json":   nullableText(resultJSON),
			"error_code":    nil,
			"error_message": nil,
			"retryable":     false,
			"finished_at":   finishedAt,
			"updated_at":    finishedAt,
		})
	return operationUpdateError(result)
}

func (d *Database) FailOperation(ctx context.Context, id uint64, code, message string, retryable bool, finishedAt time.Time) error {
	result := d.db.WithContext(ctx).Model(&CephOperation{}).
		Where("id = ? AND status = ?", id, OperationRunning).
		Updates(map[string]any{
			"status":        OperationFailed,
			"error_code":    nullableText(code),
			"error_message": nullableText(message),
			"retryable":     retryable,
			"finished_at":   finishedAt,
			"updated_at":    finishedAt,
		})
	return operationUpdateError(result)
}

func (d *Database) RequeueOperation(ctx context.Context, id uint64, code, message string, nextAttemptAt, updatedAt time.Time) error {
	result := d.db.WithContext(ctx).Model(&CephOperation{}).
		Where("id = ? AND status = ? AND attempts < max_attempts", id, OperationRunning).
		Updates(map[string]any{
			"status":          OperationQueued,
			"error_code":      nullableText(code),
			"error_message":   nullableText(message),
			"retryable":       true,
			"next_attempt_at": nextAttemptAt,
			"started_at":      nil,
			"updated_at":      updatedAt,
		})
	return operationUpdateError(result)
}

func (d *Database) RecoverRunningOperations(ctx context.Context, now time.Time) (int64, error) {
	result := d.db.WithContext(ctx).Model(&CephOperation{}).
		Where("status = ?", OperationRunning).
		Updates(map[string]any{
			"status":          OperationFailed,
			"error_code":      "operation_interrupted",
			"error_message":   "operation execution was interrupted before completion",
			"retryable":       true,
			"next_attempt_at": nil,
			"finished_at":     now,
			"updated_at":      now,
		})
	return result.RowsAffected, result.Error
}

func operationUpdateError(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
