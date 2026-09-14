package ceph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func (p *NativeProvider) collectStorageOptional(ctx context.Context, access ClusterAccess, pools []poolWire, fs fsDumpWire, now time.Time) []Observation {
	var rows []Observation
	for _, pool := range pools {
		var namespaces []string
		var namespaceRows []namedWire
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_namespace", []string{"namespace", "list", pool.PoolName, "--format", "json"}, &namespaceRows) {
			if namespaceRows == nil {
				markCollectionUnavailable(ctx, "collect.rbd_namespace")
			}
			for _, namespace := range namespaceRows {
				name := namespace.Name
				if strings.TrimSpace(name) == "" || strings.Contains(name, "/") {
					markCollectionUnavailable(ctx, "collect.rbd_namespace")
					continue
				}
				namespaces = append(namespaces, name)
				rows = append(rows, observation("rbd_namespace", pool.PoolName+"/"+name, name, "rbd_cli", map[string]any{"pool": pool.PoolName, "namespace": name, "name": name}, now))
			}
		}
		for _, namespace := range append([]string{""}, namespaces...) {
			scope := pool.PoolName
			imageArgs := []string{"ls", "--long", "--pool", pool.PoolName, "--format", "json"}
			if namespace != "" {
				scope += "/" + namespace
				imageArgs = append(imageArgs, "--namespace", namespace)
			}
			var images []rbdImageWire
			if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_image_detail", imageArgs, &images) {
				if images == nil {
					markCollectionUnavailable(ctx, "collect.rbd_image_detail")
				}
				for _, image := range images {
					if image.Snapshot != nil {
						continue
					}
					if strings.TrimSpace(image.Name) == "" {
						markCollectionUnavailable(ctx, "collect.rbd_image_detail")
						continue
					}
					spec := scope + "/" + image.Name
					imageKey := base64.RawURLEncoding.EncodeToString([]byte(spec))
					if namespace != "" {
						payload := cephdomain.RBDImage{ImagePath: spec, ImageSpec: imageKey, Pool: pool.PoolName, Namespace: namespace, Name: image.Name, SizeBytes: image.Size, Format: image.Format}
						p.enrichRBDImage(ctx, access, spec, &payload)
						rows = append(rows, observation("rbd_image", imageKey, image.Name, "rbd_cli", payload, now))
					}
					var snapshots []map[string]any
					if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_snapshot", []string{"snap", "ls", spec, "--format", "json"}, &snapshots) {
						if snapshots == nil {
							markCollectionUnavailable(ctx, "collect.rbd_snapshot")
						}
						for _, snapshot := range snapshots {
							name := textField(snapshot, "name")
							if name == "" {
								markCollectionUnavailable(ctx, "collect.rbd_snapshot")
								continue
							}
							snapshot["image_spec"] = imageKey
							snapshot["image_path"] = spec
							snapshot["pool_name"] = pool.PoolName
							snapshot["namespace"] = namespace
							rows = append(rows, Observation{Kind: "rbd_snapshot", NaturalKey: imageKey + "@" + name, ParentKind: "rbd_image", ParentKey: imageKey, Name: name, Status: "available", Source: "rbd_cli", Payload: snapshot, ObservedAt: now})
						}
					}
				}
			}
		}
		for _, namespace := range append([]string{""}, namespaces...) {
			scope := pool.PoolName
			args := []string{"trash", "ls", "--long", "--pool", pool.PoolName, "--format", "json"}
			if namespace != "" {
				scope += "/" + namespace
				args = append(args, "--namespace", namespace)
			}
			var trash []map[string]any
			if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_trash", args, &trash) {
				if trash == nil {
					markCollectionUnavailable(ctx, "collect.rbd_trash")
				}
				for _, item := range trash {
					id := textField(item, "id")
					if id == "" {
						markCollectionUnavailable(ctx, "collect.rbd_trash")
						continue
					}
					item["pool"] = pool.PoolName
					item["namespace"] = namespace
					item["image_id"] = id
					rows = append(rows, observation("rbd_trash", opaquePair(scope, id), textField(item, "name"), "rbd_cli", item, now))
				}
			}
		}
		for _, namespace := range append([]string{""}, namespaces...) {
			scope := pool.PoolName
			args := []string{"group", "list", "--pool", pool.PoolName, "--format", "json"}
			if namespace != "" {
				scope += "/" + namespace
				args = append(args, "--namespace", namespace)
			}
			var groups []string
			if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_group", args, &groups) {
				if groups == nil {
					markCollectionUnavailable(ctx, "collect.rbd_group")
				}
				for _, name := range groups {
					spec := scope + "/" + name
					payload := map[string]any{"pool": pool.PoolName, "namespace": namespace, "name": name, "group_spec": spec}
					var info map[string]any
					if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_group_info", []string{"group", "info", spec, "--format", "json"}, &info) {
						if id := textField(info, "group_id"); id != "" {
							payload["group_id"] = id
						} else {
							markCollectionUnavailable(ctx, "collect.rbd_group_info")
						}
					}
					var images, snapshots any
					if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_group_images", []string{"group", "image", "list", spec, "--format", "json"}, &images) {
						payload["images"] = images
					}
					if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_group_snapshots", []string{"group", "snap", "list", spec, "--format", "json"}, &snapshots) {
						payload["snapshots"] = snapshots
					}
					rows = append(rows, observation("rbd_group", spec, name, "rbd_cli", payload, now))
				}
			}
		}
		mirroring, ok := p.collectPoolMirroring(ctx, access, pool.PoolName)
		if ok {
			mirroring["pool"] = pool.PoolName
			if mirroring["mode"] != "disabled" {
				var status map[string]any
				if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirroring_status", []string{"mirror", "pool", "status", pool.PoolName, "--verbose", "--format", "json"}, &status) {
					mirroring["summary"] = status["summary"]
					mirroring["daemons"] = status["daemons"]
					mirroring["images"] = status["images"]
				}
			}
			rows = append(rows, observation("rbd_mirroring", pool.PoolName, pool.PoolName, "rbd_cli", mirroring, now))
		}
	}
	for _, filesystem := range fs.Filesystems {
		name := filesystem.MDSMap.FSName
		var groups []namedWire
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_group", []string{"fs", "subvolumegroup", "ls", name, "--format", "json"}, &groups) {
			for _, group := range groups {
				payload := map[string]any{"filesystem": name, "name": group.Name}
				var details map[string]any
				if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_group", []string{"fs", "subvolumegroup", "info", name, group.Name, "--format", "json"}, &details) {
					for key, value := range details {
						payload[key] = value
					}
				}
				rows = append(rows, Observation{Kind: "subvolume_group", NaturalKey: name + "/" + group.Name, ParentKind: "filesystem", ParentKey: name, Name: group.Name, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
			}
		}
		var subvolumes []namedWire
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume_detail", []string{"fs", "subvolume", "ls", name, "--format", "json"}, &subvolumes) {
			for _, subvolume := range subvolumes {
				var snapshots []map[string]any
				if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_snapshot", []string{"fs", "subvolume", "snapshot", "ls", name, subvolume.Name, "--format", "json"}, &snapshots) {
					for _, snapshot := range snapshots {
						snapshotName := textField(snapshot, "name")
						if snapshotName == "" {
							continue
						}
						parent := name + "/" + subvolume.Name
						payload := map[string]any{"filesystem": name, "subvolume": subvolume.Name, "name": snapshotName}
						for key, value := range snapshot {
							payload[key] = value
						}
						rows = append(rows, Observation{Kind: "cephfs_snapshot", NaturalKey: parent + "/" + snapshotName, ParentKind: "subvolume", ParentKey: parent, Name: snapshotName, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
					}
				}
			}
		}
	}
	rows = append(rows, p.collectRGWOptional(ctx, access, now)...)
	rows = append(rows, p.collectGatewayOptional(ctx, access, now)...)
	var removals any
	if p.optional(ctx, access, executor.BinaryCeph, "collect.osd_removal", []string{"orch", "osd", "rm", "status", "--format", "json"}, &removals) {
		for index, item := range objectList(removals) {
			key := textField(item, "osd_id", "osd", "id")
			if key == "" {
				key = strconv.Itoa(index)
			}
			rows = append(rows, observation("osd_removal", key, key, "ceph_cli", item, now))
		}
	}
	return rows
}

func (p *NativeProvider) collectPoolMirroringMode(ctx context.Context, access ClusterAccess, pool string) *string {
	mirroring, ok := p.collectPoolMirroring(ctx, access, pool)
	if !ok {
		return nil
	}
	mode := firstNonEmpty(textField(mirroring, "mirror_mode"), textField(mirroring, "mode"))
	if mode == "" {
		return nil
	}
	return &mode
}

func (p *NativeProvider) collectPoolMirroring(ctx context.Context, access ClusterAccess, pool string) (map[string]any, bool) {
	if strings.TrimSpace(pool) == "" {
		return nil, false
	}
	var mirroring map[string]any
	if !p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirroring", []string{"mirror", "pool", "info", pool, "--format", "json"}, &mirroring) {
		return nil, false
	}
	mode, ok := mirroring["mode"].(string)
	if !ok || (mode != "disabled" && mode != "image" && mode != "pool" && mode != "init-only") {
		markCollectionUnavailable(ctx, "collect.rbd_mirroring")
		return nil, false
	}
	return mirroring, true
}

func (p *NativeProvider) collectRGWOptional(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rows []Observation
	for _, resource := range []struct {
		kind, noun, listKey, idFlag string
	}{{"rgw_user", "user", "", "--uid"}, {"rgw_account", "account", "accounts", "--account-id"}, {"rgw_role", "role", "roles", "--role-name"}} {
		var list any
		if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind, []string{resource.noun, "list", "--format", "json"}, &list) {
			continue
		}
		if list == nil {
			markCollectionUnavailable(ctx, "collect."+resource.kind)
			continue
		}
		if resource.kind == "rgw_role" {
			rows = append(rows, rgwRoleObservations(ctx, list, "", now)...)
			continue
		}

		for _, id := range stringList(list, resource.listKey) {
			var details map[string]any
			verb := "info"
			if resource.noun == "account" || resource.noun == "role" {
				verb = "get"
			}
			if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind+"_detail", []string{resource.noun, verb, resource.idFlag, id, "--format", "json"}, &details) {
				continue
			}
			if len(details) == 0 {
				markCollectionUnavailable(ctx, "collect."+resource.kind+"_detail")
				continue
			}
			if resource.kind == "rgw_account" {
				details["account_id"] = id
				details["account_name"] = textField(details, "name")
				var stats map[string]any
				if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_account_stats", []string{"account", "stats", "--account-id", id, "--format", "json"}, &stats) {
					if stats != nil && stats["stats"] != nil {
						details["storage_stats"] = stats
					} else {
						markCollectionUnavailable(ctx, "collect.rgw_account_stats")
					}
				}
				var roles any
				if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_role", []string{"role", "list", "--account-id", id, "--format", "json"}, &roles) {
					rows = append(rows, rgwRoleObservations(ctx, roles, id, now)...)
				}

			}
			if resource.kind == "rgw_user" {
				details["uid"] = id
				var limits map[string]any
				if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_user_ratelimit", []string{"ratelimit", "get", "--uid", id, "--ratelimit-scope", "user", "--format", "json"}, &limits) {
					if value, ok := limits["user_ratelimit"].(map[string]any); ok {
						details["rate_limit"] = value
					} else {
						markCollectionUnavailable(ctx, "collect.rgw_user_ratelimit")
					}
				}
				var stats map[string]any
				if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_user_stats", []string{"user", "stats", "--uid", id, "--format", "json"}, &stats) {
					if stats != nil {
						details["storage_stats"] = stats
					} else {
						markCollectionUnavailable(ctx, "collect.rgw_user_stats")
					}
				}
				details["stats_scope"] = "user"
				if textField(details, "account_id") != "" {
					details["stats_scope"] = "account"
				}
			}
			rows = append(rows, observation(resource.kind, id, id, "rgw_admin", details, now))
		}
	}
	var buckets any
	if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket", []string{"bucket", "list", "--format", "json"}, &buckets) {
		for _, entry := range stringList(buckets, "buckets") {
			tenant, bucket := "", entry
			if prefix, name, found := strings.Cut(entry, "/"); found {
				tenant, bucket = prefix, name
			}
			if bucket == "" || strings.Contains(bucket, "/") {
				markCollectionUnavailable(ctx, "collect.rgw_bucket_detail")
				continue
			}
			args := []string{"bucket", "stats", "--bucket", bucket, "--format", "json"}
			if tenant != "" {
				args = append(args, "--tenant", tenant)
			}
			var details map[string]any
			if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket_detail", args, &details) {
				continue
			}
			if textField(details, "bucket") != bucket || textField(details, "tenant") != tenant {
				markCollectionUnavailable(ctx, "collect.rgw_bucket_detail")
				continue
			}
			var limits map[string]any
			args = []string{"ratelimit", "get", "--bucket", bucket, "--ratelimit-scope", "bucket"}
			if tenant != "" {
				args = append(args, "--tenant", tenant)
			}
			args = append(args, "--format", "json")
			if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket_ratelimit", args, &limits) {
				if value, ok := limits["bucket_ratelimit"].(map[string]any); ok {
					details["rate_limit"] = value
				} else {
					markCollectionUnavailable(ctx, "collect.rgw_bucket_ratelimit")
				}
			}
			key := opaquePair(tenant, bucket)
			rows = append(rows, observation("rgw_bucket", key, bucket, "rgw_admin", details, now))
		}
	}
	for _, resource := range []struct{ kind, noun, key string }{{"rgw_realm", "realm", "realms"}, {"rgw_zonegroup", "zonegroup", "zonegroups"}, {"rgw_zone", "zone", "zones"}} {
		var list any
		if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind, []string{resource.noun, "list", "--format", "json"}, &list) {
			continue
		}
		if list == nil {
			markCollectionUnavailable(ctx, "collect."+resource.kind)
			continue
		}
		for _, name := range stringList(list, resource.key) {
			details := map[string]any{"name": name}
			if resource.kind == "rgw_realm" || resource.kind == "rgw_zonegroup" {
				if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind+"_detail", []string{resource.noun, "get", "--rgw-" + resource.noun, name, "--format", "json"}, &details) {
					continue
				}
				if textField(details, "name") != name || textField(details, "id") == "" {
					markCollectionUnavailable(ctx, "collect."+resource.kind+"_detail")
					continue
				}
			}
			rows = append(rows, observation(resource.kind, name, name, "rgw_admin", details, now))
		}
	}
	return rows
}

func (p *NativeProvider) collectGatewayOptional(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rows []Observation
	for _, gateway := range []struct{ prefix, clusterID, kind string }{{"nfs", "collect.nfs_export_cluster", "nfs_export"}, {"smb", "collect.smb_share_cluster", "smb_share"}} {
		var clusters []string
		if !p.optional(ctx, access, executor.BinaryCeph, gateway.clusterID, []string{gateway.prefix, "cluster", "ls", "--format", "json"}, &clusters) {
			continue
		}
		for _, cluster := range clusters {
			var list any
			args := []string{gateway.prefix}
			if gateway.prefix == "nfs" {
				args = append(args, "export", "ls", cluster, "--detailed", "--format", "json")
			} else {
				args = append(args, "share", "ls", cluster, "--format", "json")
			}
			if !p.optional(ctx, access, executor.BinaryCeph, "collect."+gateway.kind, args, &list) {
				continue
			}
			for index, item := range objectList(list) {
				id := textField(item, "export_id", "share_id", "name", "pseudo")
				if id == "" {
					id = strconv.Itoa(index)
				}
				rows = append(rows, Observation{Kind: gateway.kind, NaturalKey: opaquePair(cluster, id), ParentKind: gateway.prefix + "_cluster", ParentKey: cluster, Name: id, Status: "available", Source: "ceph_cli", Payload: item, ObservedAt: now})
			}
		}
	}
	return rows
}

func (p *NativeProvider) collectConfigurationOptional(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rows []Observation
	var options []string
	if p.optional(ctx, access, executor.BinaryCeph, "collect.config_option", []string{"config", "ls", "--format", "json"}, &options) {
		for _, name := range options {
			rows = append(rows, observation("config_option", name, name, "ceph_cli", map[string]any{"name": name}, now))
		}
	}
	rows = append(rows, p.collectManagerModules(ctx, access, now)...)
	var rules any
	if p.optional(ctx, access, executor.BinaryCeph, "collect.crush_rule", []string{"osd", "crush", "rule", "dump", "--format", "json"}, &rules) {
		for index, item := range objectList(rules) {
			name := textField(item, "rule_name", "name")
			if name == "" {
				name = strconv.Itoa(index)
			}
			rows = append(rows, observation("crush_rule", name, name, "ceph_cli", item, now))
		}
	}
	var profiles []string
	if p.optional(ctx, access, executor.BinaryCeph, "collect.erasure_code_profile", []string{"osd", "erasure-code-profile", "ls", "--format", "json"}, &profiles) {
		for _, name := range profiles {
			var details map[string]any
			if !p.optional(ctx, access, executor.BinaryCeph, "collect.erasure_code_profile_detail", []string{"osd", "erasure-code-profile", "get", name, "--format", "json"}, &details) {
				details = map[string]any{"name": name}
			}
			rows = append(rows, observation("erasure_code_profile", name, name, "ceph_cli", details, now))
		}
	}
	return rows
}

func (p *NativeProvider) optional(ctx context.Context, access ClusterAccess, binary executor.Binary, id string, args []string, out any) bool {
	return p.runBinaryInto(ctx, access, binary, id, args, out) == nil
}
func observation(kind, key, name, source string, payload any, now time.Time) Observation {
	return Observation{Kind: kind, NaturalKey: key, Name: name, Status: "available", Source: source, Payload: payload, ObservedAt: now}
}
func opaquePair(left, right string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(left + "\x00" + right))
}
func textField(value map[string]any, keys ...string) string {
	for _, key := range keys {
		switch item := value[key].(type) {
		case string:
			if item != "" {
				return item
			}
		case json.Number:
			return item.String()
		case float64:
			return strconv.FormatFloat(item, 'f', -1, 64)
		}
	}
	return ""
}
func stringList(value any, preferredKey string) []string {
	if object, ok := value.(map[string]any); ok {
		if preferredKey != "" {
			value = object[preferredKey]
		} else {
			for _, item := range object {
				if _, ok := item.([]any); ok {
					value = item
					break
				}
			}
		}
	}
	items, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]string); ok {
			return typed
		}
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
			result = append(result, text)
		}
	}
	return result
}
func objectList(value any) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]map[string]any); ok {
			return typed
		}
		if object, ok := value.(map[string]any); ok {
			for _, item := range object {
				if list := objectList(item); len(list) > 0 {
					return list
				}
			}
		}
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			result = append(result, object)
		}
	}
	return result
}

func (p *NativeProvider) enrichRBDImage(ctx context.Context, access ClusterAccess, spec string, image *cephdomain.RBDImage) {
	var configuration []cephdomain.PoolConfig
	if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_image_config", []string{"config", "image", "list", spec, "--format", "json"}, &configuration) {
		if configuration == nil {
			markCollectionUnavailable(ctx, "collect.rbd_image_config")
		} else {
			image.Configuration = configuration
		}
	}

	var status map[string]any
	if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_image_status", []string{"status", spec, "--format", "json"}, &status) {
		if status == nil {
			markCollectionUnavailable(ctx, "collect.rbd_image_status")
		} else {
			for _, watcher := range objectList(status["watchers"]) {
				for _, field := range []string{"client", "cookie"} {
					if number, ok := watcher[field].(json.Number); ok {
						watcher[field] = number.String()
					}
				}
			}
			image.RuntimeStatus = status
		}
	}
	var info map[string]any
	if !p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_image_info", []string{"info", spec, "--format", "json"}, &info) {
		return
	}
	if info == nil {
		markCollectionUnavailable(ctx, "collect.rbd_image_info")
		return
	}
	image.Details = info
	image.Features = stringList(info["features"], "")
	image.Parent, _ = info["parent"].(map[string]any)
}

func rgwRoleObservations(ctx context.Context, list any, account string, now time.Time) []Observation {
	items, ok := list.([]any)
	if !ok {
		markCollectionUnavailable(ctx, "collect.rgw_role")
		return nil
	}
	var rows []Observation
	for _, item := range items {
		role, ok := item.(map[string]any)
		if !ok || textField(role, "RoleName") == "" {
			markCollectionUnavailable(ctx, "collect.rgw_role")
			continue
		}
		name := textField(role, "RoleName")
		if textField(role, "AccountId") != account {
			markCollectionUnavailable(ctx, "collect.rgw_role")
			continue
		}
		key := name
		if account != "" {
			key = account + "/" + name
		}
		rows = append(rows, observation("rgw_role", key, name, "rgw_admin", role, now))
	}
	return rows
}
