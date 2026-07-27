package ceph

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func (p *NativeProvider) collectStorageOptional(ctx context.Context, access ClusterAccess, pools []poolWire, fs fsDumpWire, now time.Time) []Observation {
	var rows []Observation
	var mirroringPools []any
	for _, pool := range pools {
		var namespaces []string
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_namespace", []string{"namespace", "list", pool.PoolName, "--format", "json"}, &namespaces) {
			for _, name := range namespaces {
				rows = append(rows, observation("rbd_namespace", pool.PoolName+"/"+name, name, "rbd_cli", map[string]any{"pool": pool.PoolName, "name": name}, now))
			}
		}
		var images []rbdImageWire
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_image_detail", []string{"ls", "--long", pool.PoolName, "--format", "json"}, &images) {
			for _, image := range images {
				spec := pool.PoolName + "/" + image.Name
				imageKey := base64.RawURLEncoding.EncodeToString([]byte(spec))
				var snapshots []map[string]any
				if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_snapshot", []string{"snap", "ls", spec, "--format", "json"}, &snapshots) {
					for _, snapshot := range snapshots {
						name := textField(snapshot, "name")
						if name == "" {
							continue
						}
						rows = append(rows, Observation{Kind: "rbd_snapshot", NaturalKey: imageKey + "@" + name, ParentKind: "rbd_image", ParentKey: imageKey, Name: name, Status: "available", Source: "rbd_cli", Payload: snapshot, ObservedAt: now})
					}
				}
			}
		}
		var trash []map[string]any
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_trash", []string{"trash", "ls", pool.PoolName, "--format", "json"}, &trash) {
			for _, item := range trash {
				id := textField(item, "id")
				if id == "" {
					continue
				}
				key := opaquePair(pool.PoolName, id)
				rows = append(rows, observation("rbd_trash", key, textField(item, "name"), "rbd_cli", item, now))
			}
		}
		var groups []string
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_group", []string{"group", "list", pool.PoolName, "--format", "json"}, &groups) {
			for _, name := range groups {
				rows = append(rows, observation("rbd_group", pool.PoolName+"/"+name, name, "rbd_cli", map[string]any{"pool": pool.PoolName, "name": name}, now))
			}
		}
		var mirroring map[string]any
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_mirroring", []string{"mirror", "pool", "info", pool.PoolName, "--format", "json"}, &mirroring) {
			mirroring["pool"] = pool.PoolName
			mirroringPools = append(mirroringPools, mirroring)
		}
	}
	if len(mirroringPools) > 0 {
		rows = append(rows, observation("rbd_mirroring", "mirroring", "mirroring", "rbd_cli", map[string]any{"pools": mirroringPools}, now))
	}
	for _, filesystem := range fs.Filesystems {
		name := filesystem.MDSMap.FSName
		var groups []namedWire
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_group", []string{"fs", "subvolumegroup", "ls", name, "--format", "json"}, &groups) {
			for _, group := range groups {
				rows = append(rows, Observation{Kind: "subvolume_group", NaturalKey: name + "/" + group.Name, ParentKind: "filesystem", ParentKey: name, Name: group.Name, Status: "available", Source: "ceph_cli", Payload: map[string]any{"filesystem": name, "name": group.Name}, ObservedAt: now})
			}
		}
		var subvolumes []namedWire
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume_detail", []string{"fs", "subvolume", "ls", name, "--format", "json"}, &subvolumes) {
			for _, subvolume := range subvolumes {
				var snapshots []namedWire
				if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_snapshot", []string{"fs", "subvolume", "snapshot", "ls", name, subvolume.Name, "--format", "json"}, &snapshots) {
					for _, snapshot := range snapshots {
						parent := name + "/" + subvolume.Name
						rows = append(rows, Observation{Kind: "cephfs_snapshot", NaturalKey: parent + "/" + snapshot.Name, ParentKind: "subvolume", ParentKey: parent, Name: snapshot.Name, Status: "available", Source: "ceph_cli", Payload: map[string]any{"filesystem": name, "subvolume": subvolume.Name, "name": snapshot.Name}, ObservedAt: now})
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

func (p *NativeProvider) collectRGWOptional(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rows []Observation
	for _, resource := range []struct {
		kind, noun, listKey, idFlag string
	}{{"rgw_user", "user", "", "--uid"}, {"rgw_account", "account", "accounts", "--account-id"}, {"rgw_role", "role", "roles", "--role-name"}} {
		var list any
		if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind, []string{resource.noun, "list", "--format", "json"}, &list) {
			continue
		}
		for _, id := range stringList(list, resource.listKey) {
			var details map[string]any
			verb := "info"
			if resource.noun == "account" || resource.noun == "role" {
				verb = "get"
			}
			if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind+"_detail", []string{resource.noun, verb, resource.idFlag, id, "--format", "json"}, &details) {
				details = map[string]any{"id": id}
			}
			rows = append(rows, observation(resource.kind, id, id, "rgw_admin", details, now))
		}
	}
	var buckets any
	if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket", []string{"bucket", "list", "--format", "json"}, &buckets) {
		for _, bucket := range stringList(buckets, "buckets") {
			var details map[string]any
			if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket_detail", []string{"bucket", "stats", "--bucket", bucket, "--format", "json"}, &details) {
				details = map[string]any{"bucket": bucket}
			}
			tenant := textField(details, "tenant")
			key := opaquePair(tenant, bucket)
			rows = append(rows, observation("rgw_bucket", key, bucket, "rgw_admin", details, now))
		}
	}
	for _, resource := range []struct{ kind, noun, key string }{{"rgw_realm", "realm", "realms"}, {"rgw_zonegroup", "zonegroup", "zonegroups"}, {"rgw_zone", "zone", "zones"}} {
		var list any
		if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind, []string{resource.noun, "list", "--format", "json"}, &list) {
			continue
		}
		for _, name := range stringList(list, resource.key) {
			rows = append(rows, observation(resource.kind, name, name, "rgw_admin", map[string]any{"name": name}, now))
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
	var modules map[string]any
	if p.optional(ctx, access, executor.BinaryCeph, "collect.mgr_module", []string{"mgr", "module", "ls", "--format", "json"}, &modules) {
		for _, state := range []string{"enabled_modules", "disabled_modules", "always_on_modules"} {
			for _, name := range stringList(modules[state], "") {
				rows = append(rows, observation("mgr_module", name, name, "ceph_cli", map[string]any{"name": name, "state": state}, now))
			}
		}
	}
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
