package store

import (
	"context"
	"reflect"
	"time"
)

type CephDataFetchRun struct {
	ID              uint        `gorm:"primaryKey"`
	ClusterID       uint        `gorm:"not null;index"`
	Cluster         CephCluster `gorm:"constraint:OnDelete:CASCADE"`
	Module          string      `gorm:"size:64;not null;index"`
	Status          string      `gorm:"size:32;not null;index"`
	Source          string      `gorm:"size:32;not null"`
	StartedAt       time.Time   `gorm:"not null;index"`
	FinishedAt      *time.Time
	DurationMS      int
	RecordsUpserted int    `gorm:"not null;default:0"`
	RecordsDeleted  int    `gorm:"not null;default:0"`
	Error           string `gorm:"type:text;not null;default:''"`
	CreatedAt       time.Time
}

func (CephDataFetchRun) TableName() string {
	return "ceph_data_fetch_run"
}

type DataFetchRunFilter struct {
	ClusterID string
	Module    string
	Limit     int
}

func (d *Database) ListDataFetchRuns(ctx context.Context, filter DataFetchRunFilter) ([]CephDataFetchRun, error) {
	query := d.db.WithContext(ctx).Order("started_at desc").Limit(filter.Limit)
	if filter.ClusterID != "" {
		query = query.Where("cluster_id = ?", filter.ClusterID)
	}
	if filter.Module != "" {
		query = query.Where("module = ?", filter.Module)
	}
	var runs []CephDataFetchRun
	return runs, query.Find(&runs).Error
}

func (d *Database) LatestDataFetchRun(ctx context.Context, clusterID uint, module string) (CephDataFetchRun, error) {
	var run CephDataFetchRun
	err := d.db.WithContext(ctx).Where("cluster_id = ? AND module = ?", clusterID, module).Order("started_at desc").First(&run).Error
	return run, err
}

func (d *Database) CreateDataFetchRun(ctx context.Context, run *CephDataFetchRun) error {
	return d.db.WithContext(ctx).Create(run).Error
}
func (d *Database) FinishDataFetchRun(ctx context.Context, id uint, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&CephDataFetchRun{}).Where("id = ?", id).Updates(updates).Error
}

func (d *Database) CountClusterRecords(ctx context.Context, clusterID uint, model any) (int, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(model).Where("cluster_id = ?", clusterID).Count(&count).Error
	return int(count), err
}

func (d *Database) ReplaceClusterRecords(ctx context.Context, clusterID uint, model, records any) error {
	if err := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Delete(model).Error; err != nil {
		return err
	}
	reflectValue := reflect.ValueOf(records)
	if reflectValue.Kind() == reflect.Pointer {
		reflectValue = reflectValue.Elem()
	}
	if (reflectValue.Kind() == reflect.Slice || reflectValue.Kind() == reflect.Array) && reflectValue.Len() == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Create(records).Error
}

func (d *Database) ListClusterRecords(ctx context.Context, clusterID uint, order string, dest any) error {
	return d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order(order).Find(dest).Error
}

func (d *Database) ListClusterRecordsFiltered(ctx context.Context, clusterID uint, filters map[string]any, order string, dest any) error {
	query := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
	for field, value := range filters {
		query = query.Where(field+" = ?", value)
	}
	return query.Order(order).Find(dest).Error
}

func (d *Database) FindClusterRecord(ctx context.Context, clusterID uint, filters map[string]any, dest any) error {
	query := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
	for field, value := range filters {
		query = query.Where(field+" = ?", value)
	}
	return query.First(dest).Error
}

func (d *Database) UpsertClusterRecord(ctx context.Context, clusterID uint, identity map[string]any, value any) error {
	query := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
	for field, item := range identity {
		query = query.Where(field+" = ?", item)
	}
	return query.Assign(value).FirstOrCreate(value).Error
}
