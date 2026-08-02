package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClusterFilter struct {
	Names           []string
	ClientUsernames []string
}

type ClusterFilterOptions map[string][]string

func (d *Database) ListClusters(ctx context.Context, filter ClusterFilter) ([]CephCluster, error) {
	query := d.db.WithContext(ctx)
	if len(filter.Names) > 0 {
		query = query.Where("name IN ?", filter.Names)
	}
	if len(filter.ClientUsernames) > 0 {
		query = query.Where("client_username IN ?", filter.ClientUsernames)
	}
	var rows []CephCluster
	return rows, query.Order("id asc").Find(&rows).Error
}

func (d *Database) ClusterFilterOptions(ctx context.Context, fields []string) (ClusterFilterOptions, error) {
	options := make(ClusterFilterOptions, len(fields))
	for _, field := range fields {
		var values []string
		switch field {
		case "name":
			if err := d.db.WithContext(ctx).Model(&CephCluster{}).Distinct().Order("name asc").Pluck("name", &values).Error; err != nil {
				return nil, err
			}
		case "client_username":
			if err := d.db.WithContext(ctx).Model(&CephCluster{}).Distinct().Order("client_username asc").Pluck("client_username", &values).Error; err != nil {
				return nil, err
			}
		default:
			continue
		}
		options[field] = values
	}
	return options, nil
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
	FieldValues                               map[string][]string
}

type ResourceFilterOptions map[string][]string

type AuditFilter struct {
	ActorUsername, Action, ResourceKind, ResourceKey string
	ActorUserID                                      *uint64
	Limit                                            int
}

func (d *Database) ListResources(ctx context.Context, clusterID uint64, filter ResourceFilter) ([]CephEntityRecord, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := d.entityRows(ctx, clusterID, filter, len(filter.FieldValues) == 0, limit+1)
	if err != nil {
		return nil, err
	}
	if len(filter.FieldValues) == 0 {
		return rows, nil
	}
	filtered := make([]CephEntityRecord, 0, min(len(rows), limit+1))
	for _, row := range rows {
		if resourceMatchesFieldFilters(row, filter.FieldValues) {
			filtered = append(filtered, row)
			if len(filtered) > limit {
				break
			}
		}
	}
	return filtered, nil
}

func (d *Database) ResourceFilterOptions(ctx context.Context, clusterID uint64, filter ResourceFilter, fields []string) (ResourceFilterOptions, error) {
	options := make(ResourceFilterOptions, len(fields))
	cleanFields := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, exists := options[field]; !exists {
			options[field] = nil
			cleanFields = append(cleanFields, field)
		}
	}
	if len(cleanFields) == 0 {
		return options, nil
	}
	rows, err := d.entityRows(ctx, clusterID, filter, false, 0)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]map[string]struct{}, len(cleanFields))
	for _, field := range cleanFields {
		seen[field] = map[string]struct{}{}
	}
	for _, row := range rows {
		values := resourceFieldValues(row)
		for _, field := range cleanFields {
			for _, value := range values[field] {
				if value != "" {
					seen[field][value] = struct{}{}
				}
			}
		}
	}
	for _, field := range cleanFields {
		for value := range seen[field] {
			options[field] = append(options[field], value)
		}
		sort.Strings(options[field])
	}
	return options, nil
}

func resourceQuery(query *gorm.DB, clusterID uint64, filter ResourceFilter) *gorm.DB {
	query = query.Where("cluster_id = ?", clusterID)
	for column, value := range map[string]string{"parent_kind": filter.ParentKind, "parent_key": filter.ParentKey, "name": filter.Name, "status": filter.Status} {
		if value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	return query
}

func resourceMatchesFieldFilters(row CephEntityRecord, filters map[string][]string) bool {
	values := resourceFieldValues(row)
	for field, selected := range filters {
		if len(selected) == 0 {
			continue
		}
		available := values[field]
		if len(available) == 0 {
			return false
		}
		selectedSet := make(map[string]struct{}, len(selected))
		for _, value := range selected {
			if value != "" {
				selectedSet[value] = struct{}{}
			}
		}
		if len(selectedSet) == 0 {
			continue
		}
		matched := false
		for _, value := range available {
			if _, ok := selectedSet[value]; ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func resourceFieldValues(row CephEntityRecord) map[string][]string {
	values := map[string][]string{
		"id":               {strconv.FormatUint(row.ID, 10)},
		"kind":             {row.Kind},
		"natural_key":      {row.NaturalKey},
		"generation":       {strconv.FormatUint(row.Generation, 10)},
		"resource_version": {strconv.FormatUint(row.ResourceVersion, 10)},
		"source":           {row.Source},
		"observed_at":      {row.ObservedAt.Format(time.RFC3339Nano)},
		"created_at":       {row.CreatedAt.Format(time.RFC3339Nano)},
		"updated_at":       {row.UpdatedAt.Format(time.RFC3339Nano)},
		"stale":            {strconv.FormatBool(row.StaleAt != nil)},
	}
	addOptionalString(values, "parent_kind", row.ParentKind)
	addOptionalString(values, "parent_key", row.ParentKey)
	addOptionalString(values, "name", row.Name)
	addOptionalString(values, "status", row.Status)
	addOptionalString(values, "source_version", row.SourceVersion)
	if row.StaleAt != nil {
		values["stale_at"] = []string{row.StaleAt.Format(time.RFC3339Nano)}
	}
	dataSets := []string{row.DiscoveredData}
	if row.ConfiguredData != nil {
		dataSets = append([]string{*row.ConfiguredData}, dataSets...)
	}
	for _, data := range dataSets {
		var payload map[string]any
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			continue
		}
		for field, value := range payload {
			if _, exists := values[field]; exists {
				continue
			}
			values[field] = valueTexts(value)
		}
	}
	return values
}

func addOptionalString(values map[string][]string, field string, value *string) {
	if value != nil && *value != "" {
		values[field] = []string{*value}
	}
}

func valueTexts(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		if typed == "" {
			return nil
		}
		return []string{typed}
	case bool:
		return []string{strconv.FormatBool(typed)}
	case float64:
		return []string{strconv.FormatFloat(typed, 'f', -1, 64)}
	case []any:
		var values []string
		for _, item := range typed {
			values = append(values, valueTexts(item)...)
		}
		return values
	default:
		data, err := json.Marshal(typed)
		if err != nil || len(data) == 0 || string(data) == "null" {
			return nil
		}
		return []string{string(data)}
	}
}

func (d *Database) entityRows(ctx context.Context, clusterID uint64, filter ResourceFilter, paginate bool, limit int) ([]CephEntityRecord, error) {
	if filter.Kind == "host" {
		query := d.db.WithContext(ctx).Where("cluster_id = ?", clusterID)
		if filter.Name != "" {
			query = query.Where("hostname = ?", filter.Name)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.AfterID > 0 {
			query = query.Where("id > ?", filter.AfterID)
		}
		if paginate {
			query = query.Limit(limit)
		}
		var hosts []CephHost
		if err := query.Order("id asc").Find(&hosts).Error; err != nil {
			return nil, err
		}
		rows := make([]CephEntityRecord, 0, len(hosts))
		for _, host := range hosts {
			rows = append(rows, hostEntityRecord(host))
		}
		return rows, nil
	}
	table, ok := EntityTableName(filter.Kind)
	if !ok {
		return nil, fmt.Errorf("unsupported entity kind %q", filter.Kind)
	}
	query := resourceQuery(d.db.WithContext(ctx).Table(table), clusterID, filter)
	if filter.AfterID > 0 {
		query = query.Where("id > ?", filter.AfterID)
	}
	if paginate {
		query = query.Limit(limit)
	}
	var rows []CephEntityRecord
	if err := query.Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	for index := range rows {
		rows[index].Kind = filter.Kind
	}
	return rows, nil
}

func (d *Database) FindResource(ctx context.Context, clusterID uint64, kind, key string) (CephEntityRecord, error) {
	if kind == "overview" {
		cluster, err := d.FindCluster(ctx, clusterID)
		if err != nil {
			return CephEntityRecord{}, err
		}
		return clusterEntityRecord(cluster), nil
	}
	if kind == "host" {
		host, err := d.FindCephHost(ctx, clusterID, key)
		if err != nil {
			return CephEntityRecord{}, err
		}
		return hostEntityRecord(host), nil
	}
	table, ok := EntityTableName(kind)
	if !ok {
		return CephEntityRecord{}, fmt.Errorf("unsupported entity kind %q", kind)
	}
	var row CephEntityRecord
	err := d.db.WithContext(ctx).Table(table).Where("cluster_id = ? AND natural_key = ?", clusterID, key).First(&row).Error
	row.Kind = kind
	return row, err
}

func clusterEntityRecord(cluster CephCluster) CephEntityRecord {
	status := cluster.Status
	observed := cluster.UpdatedAt
	if cluster.ObservedAt != nil {
		observed = *cluster.ObservedAt
	}
	return CephEntityRecord{
		ID: cluster.ID, ClusterID: cluster.ID, Kind: "overview", NaturalKey: "overview",
		Name: &cluster.Name, Status: &status, Generation: cluster.Generation, ResourceVersion: max(cluster.Generation, 1),
		Source: "ceph_probe", ObservedAt: observed, DiscoveredData: cluster.DiscoveredData,
		CreatedAt: cluster.CreatedAt, UpdatedAt: cluster.UpdatedAt,
	}
}

func hostEntityRecord(host CephHost) CephEntityRecord {
	name := host.Hostname
	observed := host.UpdatedAt
	if host.ObservedAt != nil {
		observed = *host.ObservedAt
	}
	configuredValues := map[string]any{}
	if host.ConfiguredData != nil {
		_ = json.Unmarshal([]byte(*host.ConfiguredData), &configuredValues)
	}
	configuredValues["hostname"], configuredValues["address"] = host.Hostname, host.Address
	configuredValues["ssh_address"], configuredValues["ssh_port"] = host.SSHAddress, host.SSHPort
	configuredValues["ssh_user"], configuredValues["ssh_auth_method"] = host.SSHUser, host.SSHAuthMethod
	configuredValues["host_ssh_configured"], configuredValues["notes"] = host.SSHAddress != "" && host.SSHUser != "", host.Notes
	configured, _ := json.Marshal(configuredValues)
	configuredData := string(configured)
	return CephEntityRecord{
		ID: host.ID, ClusterID: host.ClusterID, Kind: "host", NaturalKey: host.Hostname,
		Name: &name, Status: host.Status, Generation: host.Generation, ResourceVersion: host.ResourceVersion,
		Source: host.Source, SourceVersion: host.SourceVersion, ObservedAt: observed, StaleAt: host.StaleAt,
		ConfiguredData: &configuredData, DiscoveredData: host.DiscoveredData, CreatedAt: host.CreatedAt, UpdatedAt: host.UpdatedAt,
	}
}
func (d *Database) UpsertCapabilities(ctx context.Context, rows []CephClusterCapability) error {
	if len(rows) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "cluster_id"}, {Name: "name"}}, DoUpdates: clause.AssignmentColumns([]string{"supported", "reason", "version", "details_json", "observed_at", "updated_at"})}).Create(&rows).Error
}
func (d *Database) ListCephHosts(ctx context.Context, clusterID uint64) ([]CephHost, error) {
	var rows []CephHost
	return rows, d.db.WithContext(ctx).Where("cluster_id = ?", clusterID).Order("hostname asc").Find(&rows).Error
}

func (d *Database) FindCephHost(ctx context.Context, clusterID uint64, hostname string) (CephHost, error) {
	var row CephHost
	return row, d.db.WithContext(ctx).Where("cluster_id = ? AND hostname = ?", clusterID, hostname).First(&row).Error
}

func (d *Database) UpsertCephHost(ctx context.Context, row *CephHost) error {
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cluster_id"}, {Name: "hostname"}},
		DoUpdates: clause.AssignmentColumns([]string{"ssh_address", "ssh_port", "ssh_user", "ssh_auth_method", "ssh_password_secret", "ssh_private_key_secret", "ssh_key_passphrase_secret", "notes", "updated_at"}),
	}).Create(row).Error
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
func (d *Database) ReconcileResources(ctx context.Context, clusterID, generation uint64, rows []CephEntityRecord, authoritativeKinds []string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		for i := range rows {
			row := &rows[i]
			row.ClusterID, row.Generation, row.UpdatedAt = clusterID, generation, now
			if row.Kind == "overview" {
				continue
			}
			if row.Kind == "host" {
				if err := reconcileHostEntity(tx, *row, now); err != nil {
					return err
				}
				continue
			}
			table, ok := EntityTableName(row.Kind)
			if !ok {
				return fmt.Errorf("unsupported entity kind %q", row.Kind)
			}
			var existing CephEntityRecord
			err := tx.Table(table).Where("cluster_id = ? AND natural_key = ?", clusterID, row.NaturalKey).First(&existing).Error
			if err == nil {
				row.ID, row.CreatedAt, row.ResourceVersion, row.ConfiguredData = existing.ID, existing.CreatedAt, existing.ResourceVersion, existing.ConfiguredData
				if existing.DiscoveredData != row.DiscoveredData || !equalStringPointer(existing.Status, row.Status) {
					row.ResourceVersion++
				}
				if err := tx.Table(table).Save(row).Error; err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				row.ResourceVersion, row.CreatedAt = 1, now
				if err := tx.Table(table).Create(row).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		for _, kind := range authoritativeKinds {
			if kind == "overview" {
				continue
			}
			if kind == "host" {
				if err := tx.Model(&CephHost{}).Where("cluster_id = ? AND generation <> ? AND stale_at IS NULL", clusterID, generation).Update("stale_at", now).Error; err != nil {
					return err
				}
				continue
			}
			table, ok := EntityTableName(kind)
			if !ok {
				return fmt.Errorf("unsupported entity kind %q", kind)
			}
			if err := tx.Table(table).Where("cluster_id = ? AND generation <> ? AND stale_at IS NULL", clusterID, generation).Update("stale_at", now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (d *Database) MarkModuleResourcesStale(ctx context.Context, clusterID uint64, kinds []string, now time.Time) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, kind := range kinds {
			if kind == "overview" {
				continue
			}
			if kind == "host" {
				if err := tx.Model(&CephHost{}).Where("cluster_id = ? AND stale_at IS NULL", clusterID).Update("stale_at", now).Error; err != nil {
					return err
				}
				continue
			}
			table, ok := EntityTableName(kind)
			if !ok {
				return fmt.Errorf("unsupported entity kind %q", kind)
			}
			if err := tx.Table(table).Where("cluster_id = ? AND stale_at IS NULL", clusterID).Update("stale_at", now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func reconcileHostEntity(tx *gorm.DB, row CephEntityRecord, now time.Time) error {
	var existing CephHost
	err := tx.Where("cluster_id = ? AND hostname = ?", row.ClusterID, row.NaturalKey).First(&existing).Error
	if err == nil {
		version := existing.ResourceVersion
		if existing.DiscoveredData != row.DiscoveredData || !equalStringPointer(existing.Status, row.Status) {
			version++
		}
		return tx.Model(&CephHost{}).Where("id = ?", existing.ID).Updates(map[string]any{
			"status": row.Status, "discovered_data": row.DiscoveredData, "generation": row.Generation,
			"resource_version": version, "source": row.Source, "source_version": row.SourceVersion,
			"observed_at": row.ObservedAt, "stale_at": nil, "updated_at": now,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	host := CephHost{
		ClusterID: row.ClusterID, Hostname: row.NaturalKey, SSHPort: 22, SSHUser: "root", SSHAuthMethod: "password",
		Status: row.Status, DiscoveredData: row.DiscoveredData, Generation: row.Generation, ResourceVersion: 1,
		Source: row.Source, SourceVersion: row.SourceVersion, ObservedAt: &row.ObservedAt, CreatedAt: now, UpdatedAt: now,
	}
	return tx.Create(&host).Error
}

func (d *Database) UpdateClusterDiscovery(ctx context.Context, row CephCluster) error {
	return d.db.WithContext(ctx).Model(&CephCluster{}).Where("id = ?", row.ID).Updates(map[string]any{
		"discovered_data": row.DiscoveredData, "fsid": row.FSID, "ceph_version": row.CephVersion,
		"status": row.Status, "enabled": row.Enabled, "generation": row.Generation,
		"last_seen_at": row.LastSeenAt, "last_error_code": row.LastErrorCode,
		"last_error_message": row.LastErrorMessage, "observed_at": row.ObservedAt, "updated_at": row.UpdatedAt,
	}).Error
}

func (d *Database) SaveResourceConfiguration(ctx context.Context, clusterID uint64, kind, key, configuredData string) error {
	if key == "" || configuredData == "" {
		return nil
	}
	if kind == "host" {
		var values map[string]any
		_ = json.Unmarshal([]byte(configuredData), &values)
		var address *string
		if value, ok := values["address"].(string); ok && strings.TrimSpace(value) != "" {
			cleaned := strings.TrimSpace(value)
			address = &cleaned
		}
		now := time.Now().UTC()
		var existing CephHost
		err := d.db.WithContext(ctx).Where("cluster_id = ? AND hostname = ?", clusterID, key).First(&existing).Error
		if err == nil {
			return d.db.WithContext(ctx).Model(&CephHost{}).Where("id = ?", existing.ID).Updates(map[string]any{
				"address": address, "configured_data": configuredData, "updated_at": now,
			}).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return d.db.WithContext(ctx).Create(&CephHost{
			ClusterID: clusterID, Hostname: key, Address: address, SSHPort: 22, SSHUser: "root", SSHAuthMethod: "password",
			ConfiguredData: &configuredData, DiscoveredData: "{}", ResourceVersion: 1, Source: "user", CreatedAt: now, UpdatedAt: now,
		}).Error
	}
	table, ok := EntityTableName(kind)
	if !ok {
		return nil
	}
	now := time.Now().UTC()
	var existing CephEntityRecord
	err := d.db.WithContext(ctx).Table(table).Where("cluster_id = ? AND natural_key = ?", clusterID, key).First(&existing).Error
	if err == nil {
		return d.db.WithContext(ctx).Table(table).Where("id = ?", existing.ID).Updates(map[string]any{
			"configured_data": configuredData, "updated_at": now,
		}).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	name := key
	return d.db.WithContext(ctx).Table(table).Create(&CephEntityRecord{
		ClusterID: clusterID, NaturalKey: key, Name: &name, ResourceVersion: 1, Source: "user",
		ObservedAt: now, ConfiguredData: &configuredData, DiscoveredData: "{}", CreatedAt: now, UpdatedAt: now,
	}).Error
}

func (d *Database) DeleteResourceState(ctx context.Context, clusterID uint64, kind, key string) error {
	if key == "" {
		return nil
	}
	if kind == "host" {
		return d.db.WithContext(ctx).Where("cluster_id = ? AND hostname = ?", clusterID, key).Delete(&CephHost{}).Error
	}
	table, ok := EntityTableName(kind)
	if !ok {
		return nil
	}
	return d.db.WithContext(ctx).Table(table).Where("cluster_id = ? AND natural_key = ?", clusterID, key).Delete(&CephEntityRecord{}).Error
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
		ID                                                                           uint64
		OccurredAt                                                                   time.Time `json:"occurred_at"`
		EventType, RequestID, ActorUsername, Action, Outcome                         string
		ActorUserID, ClusterID, BeforeGeneration, AfterGeneration                    *uint64
		HTTPStatus                                                                   *int
		ClusterName, ResourceKind, ResourceKey, Risk, ErrorCode, SourceIP, UserAgent *string
		ParametersJSON, DetailsJSON, PreviousHash                                    *string
	}{row.ID, row.OccurredAt.UTC(), row.EventType, row.RequestID, row.ActorUsername, row.Action, row.Outcome, row.ActorUserID, row.ClusterID, row.BeforeGeneration, row.AfterGeneration, row.HTTPStatus, row.ClusterName, row.ResourceKind, row.ResourceKey, row.Risk, row.ErrorCode, row.SourceIP, row.UserAgent, row.ParametersJSON, row.DetailsJSON, row.PreviousHash})
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
