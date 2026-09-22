package store

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ObservationHistoryFilter struct {
	ClusterID  uint64
	Kind       string
	NaturalKey string
	Since      *time.Time
	Limit      int
}

func (d *Database) AppendObservationHistoryIfDue(ctx context.Context, row *CephObservationHistory, minimumInterval time.Duration) (bool, error) {
	if row.ClusterID == 0 || row.Kind == "" || row.NaturalKey == "" || row.ObservedAt.IsZero() {
		return false, errors.New("cluster, kind, natural key, and observation time are required")
	}
	inserted := false
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var latest CephObservationHistory
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("cluster_id = ? AND kind = ? AND natural_key = ?", row.ClusterID, row.Kind, row.NaturalKey).
			Order("observed_at desc").First(&latest).Error
		if err == nil {
			if !row.ObservedAt.After(latest.ObservedAt) || row.ObservedAt.Sub(latest.ObservedAt) < minimumInterval {
				return nil
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if row.CreatedAt.IsZero() {
			row.CreatedAt = time.Now().UTC()
		}
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		inserted = true
		return nil
	})
	return inserted, err
}

func (d *Database) ListObservationHistory(ctx context.Context, filter ObservationHistoryFilter) ([]CephObservationHistory, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	query := d.db.WithContext(ctx).Where("cluster_id = ? AND kind = ?", filter.ClusterID, filter.Kind)
	if filter.NaturalKey != "" {
		query = query.Where("natural_key = ?", filter.NaturalKey)
	}
	if filter.Since != nil {
		query = query.Where("observed_at >= ?", *filter.Since)
	}
	var rows []CephObservationHistory
	return rows, query.Order("observed_at desc, id desc").Limit(limit).Find(&rows).Error
}

func (d *Database) PruneObservationHistory(ctx context.Context, kind string, before time.Time) (int64, error) {
	result := d.db.WithContext(ctx).Where("kind = ? AND observed_at < ?", kind, before).Delete(&CephObservationHistory{})
	return result.RowsAffected, result.Error
}
