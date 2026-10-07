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
		var mirrorImageMetadata []cephdomain.RBDImage
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
					payload := cephdomain.RBDImage{ImagePath: spec, ImageSpec: imageKey, Pool: pool.PoolName, Namespace: namespace, Name: image.Name, SizeBytes: image.Size, Format: image.Format}
					p.enrichRBDImage(ctx, access, spec, &payload)
					if namespace == "" {
						mirrorImageMetadata = append(mirrorImageMetadata, payload)
					}
					imageRowIndex := len(rows)
					rows = append(rows, observation("rbd_image", imageKey, image.Name, "rbd_cli", payload, now))
					var snapshots []map[string]any
					if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_snapshot", []string{"snap", "ls", spec, "--format", "json"}, &snapshots) {
						checkedSnapshots := 0
						hasChildren := false
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
							if protected, ok := rbdSnapshotProtected(snapshot["protected"]); ok {
								snapshot["protected"] = protected
								snapshot["is_protected"] = protected
							} else {
								markCollectionUnavailable(ctx, "collect.rbd_snapshot")
								continue
							}
							// snap ls without --all returns user snapshots, including on
							// snapshot-mirrored images. Their children must still be read.
							var children []map[string]any
							if !p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_snapshot_children", []string{"children", spec + "@" + name, "--all", "--format", "json"}, &children) {
								continue
							}
							if children == nil {
								markCollectionUnavailable(ctx, "collect.rbd_snapshot_children")
								continue
							}
							snapshot["children"] = children
							checkedSnapshots++
							hasChildren = hasChildren || len(children) > 0
							snapshot["image_features"] = payload.Features
							if used, ok := payload.SnapshotUsage[name]; ok {
								snapshot["used_bytes"] = strconv.FormatUint(used, 10)
								snapshot["disk_usage"] = strconv.FormatUint(used, 10)
							}
							rows = append(rows, Observation{Kind: "rbd_snapshot", NaturalKey: imageKey + "@" + name, ParentKind: "rbd_image", ParentKey: imageKey, Name: name, Status: "available", Source: "rbd_cli", Payload: snapshot, ObservedAt: now})
						}
						if hasChildren || (snapshots != nil && checkedSnapshots == len(snapshots)) {
							payload.HasSnapshotChildren = &hasChildren
						}
					}
					rows[imageRowIndex].Payload = payload
				}
			}
		}
		for _, namespace := range append([]string{""}, namespaces...) {
			scope := pool.PoolName
			args := []string{"trash", "ls", "--all", "--long", "--pool", pool.PoolName, "--format", "json"}
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
					item["deferment_status"] = item["status"]
					item["trash_source"] = item["source"]
					delete(item, "status")
					delete(item, "source")
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
						if validRBDGroupRows(images, false) {
							payload["images"] = images
						} else {
							markCollectionUnavailable(ctx, "collect.rbd_group_images")
						}
					}
					if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_group_snapshots", []string{"group", "snap", "list", spec, "--format", "json"}, &snapshots) {
						if validRBDGroupRows(snapshots, true) {
							payload["snapshots"] = snapshots
						} else {
							markCollectionUnavailable(ctx, "collect.rbd_group_snapshots")
						}
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
					enrichMirrorReplayMetrics(status["images"])
					enrichMirrorImageMetadata(status["images"], mirrorImageMetadata)
					mirroring["images"] = status["images"]
				}
			}
			rows = append(rows, observation("rbd_mirroring", pool.PoolName, pool.PoolName, "rbd_cli", mirroring, now))
		}
	}
	for _, filesystem := range fs.Filesystems {
		name := filesystem.MDSMap.FSName
		rows = append(rows, p.collectCephFSSnapshotSchedules(ctx, access, name, now)...)
		rows = append(rows, p.collectCephFSSubvolumeScope(ctx, access, name, "", now)...)
		var groups []namedWire
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_group", []string{"fs", "subvolumegroup", "ls", name, "--format", "json"}, &groups) {
			for _, group := range groups {
				payload := map[string]any{"filesystem": name, "name": group.Name}
				var details map[string]any
				if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_group", []string{"fs", "subvolumegroup", "info", name, group.Name, "--format", "json"}, &details) {
					mergeCephFSInfo(payload, details)
				}
				rows = append(rows, Observation{Kind: "subvolume_group", NaturalKey: name + "/" + group.Name, ParentKind: "filesystem", ParentKey: name, Name: group.Name, Status: "available", Source: "ceph_cli", Payload: payload, ObservedAt: now})
				rows = append(rows, p.collectCephFSSubvolumeScope(ctx, access, name, group.Name, now)...)
			}
		}
	}
	rows = append(rows, p.collectRGWOptional(ctx, access, now)...)
	rows = append(rows, p.collectGatewayOptional(ctx, access, now)...)
	rows = append(rows, p.collectOSDRemovals(ctx, access, now)...)
	return rows
}

func (p *NativeProvider) collectCephFSSubvolumeScope(ctx context.Context, access ClusterAccess, filesystem, group string, now time.Time) []Observation {
	groupKey := group
	if groupKey == "" {
		groupKey = "_nogroup"
	}
	listArgs := []string{"fs", "subvolume", "ls", filesystem}
	if group != "" {
		listArgs = append(listArgs, group)
	}
	listArgs = append(listArgs, "--format", "json")
	var subvolumes []namedWire
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume", listArgs, &subvolumes) {
		return nil
	}
	rows := make([]Observation, 0, len(subvolumes)*2)
	for _, subvolume := range subvolumes {
		if strings.TrimSpace(subvolume.Name) == "" {
			continue
		}
		parent := filesystem + "/" + groupKey + "/" + subvolume.Name
		payload := map[string]any{"fs": filesystem, "filesystem": filesystem, "group": groupKey, "name": subvolume.Name}
		infoArgs := []string{"fs", "subvolume", "info", filesystem, subvolume.Name}
		if group != "" {
			infoArgs = append(infoArgs, group)
		}
		infoArgs = append(infoArgs, "--format", "json")
		var details map[string]any
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_subvolume_detail", infoArgs, &details) {
			mergeCephFSInfo(payload, details)
			if details["type"] == "clone" && details["state"] != "snapshot-retained" {
				statusArgs := []string{"fs", "clone", "status", filesystem, subvolume.Name}
				if group != "" {
					statusArgs = append(statusArgs, group)
				}
				statusArgs = append(statusArgs, "--format", "json")
				var clone map[string]any
				if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_clone_status", statusArgs, &clone) {
					if status, ok := clone["status"].(map[string]any); ok {
						payload["clone_status"] = status
						payload["clone_state"] = status["state"]
						payload["clone_source"] = status["source"]
						payload["clone_progress"] = status["progress_report"]
						payload["clone_failure"] = status["failure"]
					}
				}
			}
		}
		state := textField(payload, "state")
		if state == "" {
			state = "unknown"
		}
		rows = append(rows, Observation{Kind: "subvolume", NaturalKey: parent, ParentKind: "filesystem", ParentKey: filesystem, Name: subvolume.Name, Status: state, Source: "ceph_cli", Payload: payload, ObservedAt: now})

		snapshotArgs := []string{"fs", "subvolume", "snapshot", "ls", filesystem, subvolume.Name}
		if group != "" {
			snapshotArgs = append(snapshotArgs, group)
		}
		snapshotArgs = append(snapshotArgs, "--format", "json")
		var snapshots []map[string]any
		if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_snapshot", snapshotArgs, &snapshots) {
			for _, snapshot := range snapshots {
				snapshotName := textField(snapshot, "name")
				if snapshotName == "" {
					continue
				}
				snapshotPayload := map[string]any{"fs": filesystem, "filesystem": filesystem, "group": groupKey, "subvolume": subvolume.Name, "name": snapshotName}
				mergeCephFSInfo(snapshotPayload, snapshot)
				infoArgs := []string{"fs", "subvolume", "snapshot", "info", filesystem, subvolume.Name, snapshotName}
				if group != "" {
					infoArgs = append(infoArgs, group)
				}
				infoArgs = append(infoArgs, "--format", "json")
				var snapshotInfo map[string]any
				if p.optional(ctx, access, executor.BinaryCeph, "collect.cephfs_snapshot_info", infoArgs, &snapshotInfo) {
					mergeCephFSInfo(snapshotPayload, snapshotInfo)
				}
				rows = append(rows, Observation{Kind: "cephfs_snapshot", NaturalKey: parent + "/" + snapshotName, ParentKind: "subvolume", ParentKey: parent, Name: snapshotName, Status: "available", Source: "ceph_cli", Payload: snapshotPayload, ObservedAt: now})
			}
		}
	}
	return rows
}

// Native creation time must not collide with the resource envelope's cache timestamp.
func mergeCephFSInfo(payload, info map[string]any) {
	for key, value := range info {
		if key == "bytes_quota" || key == "bytes_used" {
			if number, ok := value.(json.Number); ok {
				value = number.String()
			}
		}
		if key == "created_at" {
			key = "ceph_created_at"
		}
		payload[key] = value
	}
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
		kind, noun, idFlag string
	}{{"rgw_user", "user", "--uid"}, {"rgw_account", "account", "--account-id"}, {"rgw_role", "role", "--role-name"}} {
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

		ids, valid := rgwMetadataKeys(list)
		if !valid {
			markCollectionUnavailable(ctx, "collect."+resource.kind)
			continue
		}
		for _, id := range ids {
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
				if observedID, ok := details["id"].(string); !ok || observedID != id {
					markCollectionUnavailable(ctx, "collect.rgw_account_detail")
					continue
				}
				details["account_id"] = id
				details["account_name"] = textField(details, "name")
				var stats map[string]any
				if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_account_stats", []string{"account", "stats", "--account-id", id, "--format", "json"}, &stats) {
					if counters, ok := stats["stats"].(map[string]any); ok && counters != nil {
						preserveRGWStorageCounters(counters)
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
				if observedID, ok := details["full_user_id"].(string); !ok || observedID != id {
					markCollectionUnavailable(ctx, "collect.rgw_user_detail")
					continue
				}
				details["uid"] = id
				if textField(details, "account_id") != "" && textField(details, "type") != "root" {
					var policies []string
					if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_user_policies", []string{"user", "policy", "list", "attached", "--uid", id, "--format", "json"}, &policies) {
						valid := policies != nil
						seen := make(map[string]bool, len(policies))
						for _, arn := range policies {
							if arn == "" || seen[arn] {
								valid = false
								break
							}
							seen[arn] = true
						}
						if valid {
							details["managed_user_policies"] = policies
						} else {
							markCollectionUnavailable(ctx, "collect.rgw_user_policies")
						}
					}
				}
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
					if counters, ok := stats["stats"].(map[string]any); ok && counters != nil {
						preserveRGWStorageCounters(counters)
						details["storage_stats"] = stats
					} else {
						markCollectionUnavailable(ctx, "collect.rgw_user_stats")
					}
				}
				details["stats_scope"] = nil
				if accountID, ok := details["account_id"].(string); ok {
					details["stats_scope"] = "user"
					if accountID != "" {
						details["stats_scope"] = "account"
					}
				}
			}
			rows = append(rows, observation(resource.kind, id, id, "rgw_admin", details, now))
		}
	}
	var buckets any
	if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket", []string{"metadata", "list", "bucket", "--format", "json"}, &buckets) {
		bucketNames, valid := rgwMetadataKeys(buckets)
		if !valid {
			markCollectionUnavailable(ctx, "collect.rgw_bucket")
			bucketNames = nil
		}
		var progress any
		progressAvailable := p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_lifecycle_progress", []string{"lc", "list", "--format", "json"}, &progress)
		progressEntries, progressValid := rgwLifecycleProgress(progress)
		if progressAvailable && !progressValid {
			markCollectionUnavailable(ctx, "collect.rgw_lifecycle_progress")
		}
		for _, entry := range bucketNames {
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
			observedBucket, bucketPresent := details["bucket"].(string)
			observedTenant, tenantPresent := details["tenant"].(string)
			if !bucketPresent || !tenantPresent || observedBucket != bucket || observedTenant != tenant {
				markCollectionUnavailable(ctx, "collect.rgw_bucket_detail")
				continue
			}
			var limits map[string]any
			args = []string{"ratelimit", "get", "--bucket", bucket, "--ratelimit-scope", "bucket"}
			preserveRGWBucketUsage(details)
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
			var syncPolicy any
			details["bucket_sync_policy"] = nil
			if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_bucket_sync_policy",
				[]string{"sync", "policy", "get", "--bucket", bucket, "--tenant", tenant, "--format", "json"}, &syncPolicy) {
				if policy, valid := rgwBucketSyncPolicy(syncPolicy); valid {
					details["bucket_sync_policy"] = policy
				} else {
					markCollectionUnavailable(ctx, "collect.rgw_bucket_sync_policy")
				}
			}
			details["lifecycle_progress"] = rgwBucketLifecycleProgress(details, progressEntries, progressAvailable && progressValid)
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
			if resource.kind == "rgw_realm" || resource.kind == "rgw_zonegroup" || resource.kind == "rgw_zone" {
				if !p.optional(ctx, access, executor.BinaryRGWAdmin, "collect."+resource.kind+"_detail", []string{resource.noun, "get", "--rgw-" + resource.noun, name, "--format", "json"}, &details) {
					continue
				}
				if textField(details, "name") != name || textField(details, "id") == "" {
					markCollectionUnavailable(ctx, "collect."+resource.kind+"_detail")
					continue
				}
			}
			if listed, ok := list.(map[string]any); ok {
				if defaultID, valid := listed["default_info"].(string); valid {
					details["is_default"] = defaultID == textField(details, "id")
				}
			}
			if resource.kind == "rgw_realm" {
				details["current_period_details"] = nil
				realmID, periodID := textField(details, "id"), textField(details, "current_period")
				if periodID != "" {
					var period map[string]any
					if p.optional(ctx, access, executor.BinaryRGWAdmin, "collect.rgw_current_period", []string{"period", "get", "--realm-id", realmID, "--period", periodID, "--format", "json"}, &period) {
						if period != nil && textField(period, "id") == periodID && textField(period, "realm_id") == realmID {
							details["current_period_details"] = period
						} else {
							markCollectionUnavailable(ctx, "collect.rgw_current_period")
						}
					}
				}
			}
			if resource.kind == "rgw_zone" {
				attachRGWZonePoolReferences(details)
			}
			rows = append(rows, observation(resource.kind, name, name, "rgw_admin", details, now))
		}
	}
	attachZoneMemberships(rows)
	rows = append(rows, p.collectRGWTopics(ctx, access, now)...)
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
			if gateway.prefix == "smb" {
				rows = append(rows, p.collectSMBShares(ctx, access, cluster, now)...)
				continue
			}
			var list any
			args := []string{"nfs", "export", "ls", cluster, "--detailed", "--format", "json"}
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
		valid := options != nil
		seen := make(map[string]bool, len(options))
		for _, name := range options {
			if name == "" || strings.TrimSpace(name) != name || seen[name] {
				valid = false
				break
			}
			seen[name] = true
		}
		if !valid {
			markCollectionUnavailable(ctx, "collect.config_option")
		} else {
			for _, name := range options {
				rows = append(rows, observation("config_option", name, name, "ceph_cli", map[string]any{"name": name}, now))
			}
		}
	}
	rows = append(rows, p.collectManagerModules(ctx, access, now)...)
	rows = append(rows, p.collectCrushRules(ctx, access, now)...)
	return append(rows, p.collectErasureProfiles(ctx, access, now)...)
}

func (p *NativeProvider) collectCrushRules(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rules []map[string]any
	if !p.optional(ctx, access, executor.BinaryCeph, "collect.crush_rule", []string{"osd", "crush", "rule", "dump", "--format", "json"}, &rules) {
		return nil
	}
	var rows []Observation
	names, ids := map[string]bool{}, map[int64]bool{}
	if rules == nil {
		markCollectionUnavailable(ctx, "collect.crush_rule")
		return nil
	}
	for _, item := range rules {
		name, named := item["rule_name"].(string)
		idNumber, numbered := item["rule_id"].(json.Number)
		id, err := idNumber.Int64()
		if !named || strings.TrimSpace(name) == "" || !numbered || err != nil || id < 0 || names[name] || ids[id] {
			markCollectionUnavailable(ctx, "collect.crush_rule")
			return nil
		}
		names[name], ids[id] = true, true
		rows = append(rows, observation("crush_rule", name, name, "ceph_cli", item, now))
	}
	return rows
}

func (p *NativeProvider) collectErasureProfiles(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	var rows []Observation
	var profiles []string
	if p.optional(ctx, access, executor.BinaryCeph, "collect.erasure_code_profile", []string{"osd", "erasure-code-profile", "ls", "--format", "json"}, &profiles) {
		if profiles == nil {
			markCollectionUnavailable(ctx, "collect.erasure_code_profile")
			return nil
		}
		seen := map[string]bool{}
		for _, name := range profiles {
			if strings.TrimSpace(name) == "" || seen[name] {
				markCollectionUnavailable(ctx, "collect.erasure_code_profile")
				return nil
			}
			seen[name] = true
			var native erasureProfileWire
			if !p.optional(ctx, access, executor.BinaryCeph, "collect.erasure_code_profile_detail", []string{"osd", "erasure-code-profile", "get", name, "--format", "json"}, &native) || strings.TrimSpace(native["plugin"]) == "" {
				markCollectionUnavailable(ctx, "collect.erasure_code_profile")
				return nil
			}
			details := map[string]string(native)
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
			if cache, ok := status["persistent_cache"].(map[string]any); ok {
				for field, value := range cache {
					if number, ok := value.(json.Number); ok {
						cache[field] = number.String()
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
	image.Features = rbdImageFeatures(info["features"])
	image.Parent, _ = info["parent"].(map[string]any)
	if size := firstUintPointer(info, "size"); size != nil {
		image.SizeBytes = size
	}
	image.ObjectCount = firstUintPointer(info, "objects")
	if rawLimit, ok := info["snapshot_limit"].(json.Number); ok {
		if limit, err := strconv.ParseUint(rawLimit.String(), 10, 64); err == nil && limit < ^uint64(0) {
			image.SnapshotLimit = &limit
		}
	}
	image.ObjectSize = firstUintPointer(info, "object_size")
	image.StripeUnit = firstUintPointer(info, "stripe_unit")
	image.StripeCount = firstUintPointer(info, "stripe_count")
	image.Order = firstUintPointer(info, "order")
	image.CreatedAt = textField(info, "create_timestamp")
	image.DataPool = textField(info, "data_pool")
	image.BlockPrefix = textField(info, "block_name_prefix")
	image.MirrorState = "disabled"
	if mirroring, ok := info["mirroring"].(map[string]any); ok {
		image.MirrorMode = textField(mirroring, "mode")
		if state := textField(mirroring, "state"); state != "" {
			image.MirrorState = state
		}
		image.MirrorGlobalID = textField(mirroring, "global_id")
		if primary, ok := mirroring["primary"].(bool); ok {
			image.Primary = &primary
		}
	}
	if hasString(image.Features, "fast-diff") {
		var usage map[string]any
		if p.optional(ctx, access, executor.BinaryRBD, "collect.rbd_image_usage", []string{"du", spec, "--format", "json"}, &usage) {
			if !applyRBDImageUsage(image, usage) {
				markCollectionUnavailable(ctx, "collect.rbd_image_usage")
			}
		}
	}
}

func validRBDGroupRows(value any, snapshots bool) bool {
	rows, ok := value.([]any)
	if !ok || rows == nil {
		return false
	}
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			return false
		}
		fields := []string{"pool", "namespace", "image"}
		if snapshots {
			fields = []string{"id", "snapshot", "state"}
		}
		for _, field := range fields {
			text, ok := row[field].(string)
			if !ok || (text == "" && field != "namespace") {
				return false
			}
		}
		if !snapshots {
			state, ok := row["state"].(json.Number)
			if !ok {
				return false
			}
			if _, err := strconv.ParseInt(state.String(), 10, 32); err != nil {
				return false
			}
		}
	}
	return true
}

func rbdImageFeatures(value any) []string {
	items, ok := value.([]any)
	if !ok || items == nil {
		return nil
	}
	features := make([]string, 0, len(items))
	for _, item := range items {
		feature, ok := item.(string)
		if !ok || strings.TrimSpace(feature) == "" {
			return nil
		}
		features = append(features, feature)
	}
	return features
}

func hasString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func applyRBDImageUsage(image *cephdomain.RBDImage, usage map[string]any) bool {
	rows, ok := usage["images"].([]any)
	if !ok || len(rows) == 0 {
		return false
	}
	var total uint64
	var current *uint64
	snapshots := map[string]uint64{}
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok || row == nil {
			return false
		}
		rawUsed, ok := row["used_size"].(json.Number)
		if !ok {
			return false
		}
		used, err := strconv.ParseUint(rawUsed.String(), 10, 64)
		if err != nil {
			return false
		}
		if used > ^uint64(0)-total {
			return false
		}
		total += used
		if rawSnapshot, isSnapshot := row["snapshot"]; isSnapshot {
			snapshot, ok := rawSnapshot.(string)
			if !ok || snapshot == "" {
				return false
			}
			if _, duplicate := snapshots[snapshot]; duplicate {
				return false
			}
			snapshots[snapshot] = used
		} else {
			if current != nil {
				return false
			}
			value := used
			current = &value
		}
	}
	if current == nil {
		return false
	}
	image.UsedBytes = current
	image.TotalUsedBytes = &total
	image.SnapshotUsage = snapshots
	return true
}

func rbdSnapshotProtected(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "yes":
			return true, true
		case "false", "no":
			return false, true
		}
	}
	return false, false
}

func rgwRoleObservations(ctx context.Context, list any, account string, now time.Time) []Observation {
	items, ok := list.([]any)
	if !ok || items == nil {
		markCollectionUnavailable(ctx, "collect.rgw_role")
		return nil
	}
	var rows []Observation
	for _, item := range items {
		role, ok := item.(map[string]any)
		name, validName := role["RoleName"].(string)
		roleAccount, validAccount := role["AccountId"].(string)
		if !ok || !validName || name == "" || !validAccount || roleAccount != account {
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

// ZoneParams contains pool configuration; topology lives in ZoneGroup member records.
func attachZoneMemberships(rows []Observation) {
	members := map[string][]any{}
	for _, row := range rows {
		if row.Kind != "rgw_zonegroup" {
			continue
		}
		group, ok := row.Payload.(map[string]any)
		if !ok {
			continue
		}
		for _, zone := range objectList(group["zones"]) {
			id := textField(zone, "id")
			if id == "" {
				continue
			}
			member := map[string]any{"zonegroup_id": textField(group, "id"), "zonegroup_name": row.Name, "realm_id": textField(group, "realm_id"), "is_master": textField(group, "master_zone") == id}
			for _, key := range []string{"endpoints", "log_meta", "log_data", "bucket_index_max_shards", "read_only", "tier_type", "sync_from_all", "sync_from", "redirect_zone", "supported_features"} {
				if value, present := zone[key]; present {
					member[key] = value
				}
			}
			members[id] = append(members[id], member)
		}
	}
	for _, row := range rows {
		if row.Kind != "rgw_zone" {
			continue
		}
		zone, ok := row.Payload.(map[string]any)
		if !ok {
			continue
		}
		if matches := members[textField(zone, "id")]; len(matches) > 0 {
			zone["zonegroup_memberships"] = matches
		}
	}
}
