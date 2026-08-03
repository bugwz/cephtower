package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

type refreshResourceRequest struct {
	ClusterID uint64   `json:"cluster_id"`
	Scope     string   `json:"scope,omitempty"`
	Module    string   `json:"module,omitempty"`
	Modules   []string `json:"modules,omitempty"`
	Kind      string   `json:"kind,omitempty"`
	Kinds     []string `json:"kinds,omitempty"`
}

func (h *Handler) RefreshResource(w http.ResponseWriter, r *http.Request) {
	var request refreshResourceRequest
	if !DecodeStrict(w, r, &request) {
		return
	}
	annotateAudit(r, "resource.refresh", "resource", refreshAuditKey(request), string(cephdomain.RiskLow), &request.ClusterID)
	if h.Reconciler == nil {
		WriteError(w, r, http.StatusNotImplemented, "capability_unavailable", "resource refresh is unavailable", false, nil)
		return
	}
	modules := append([]string(nil), request.Modules...)
	if request.Module != "" {
		modules = append(modules, request.Module)
	}
	kinds := append([]string(nil), request.Kinds...)
	if request.Kind != "" {
		kinds = append(kinds, request.Kind)
	}
	if request.Scope == "all" {
		modules = nil
		kinds = nil
	}
	var result any
	var err error
	switch {
	case len(kinds) > 0:
		result, err = h.Reconciler.RefreshKinds(r.Context(), request.ClusterID, kinds)
	case len(modules) > 0:
		result, err = h.Reconciler.Refresh(r.Context(), request.ClusterID, modules)
	default:
		result, err = h.Reconciler.Refresh(r.Context(), request.ClusterID, nil)
	}
	if err != nil {
		writeActionError(w, r, err)
		return
	}
	WriteSuccess(w, http.StatusOK, "success", result)
}

func refreshAuditKey(request refreshResourceRequest) string {
	if request.Scope == "all" {
		return "cluster/all"
	}
	if request.Kind != "" {
		return "kind/" + request.Kind
	}
	if len(request.Kinds) > 0 {
		return "kind/" + strings.Join(request.Kinds, ",")
	}
	if request.Module != "" {
		return "module/" + request.Module
	}
	if len(request.Modules) > 0 {
		return "module/" + strings.Join(request.Modules, ",")
	}
	return "cluster/all"
}

func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("overview", true)(w, r)
}

func (h *Handler) ListHealthChecks(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("health_check", false)(w, r)
}

func (h *Handler) MuteHealthCheck(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("health_check", "health.mute", "low")(w, r)
}

func (h *Handler) UnmuteHealthCheck(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("health_check", "health.unmute", "low")(w, r)
}

func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("host", false)(w, r)
}

func (h *Handler) CreateHost(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("host", "host.create", "medium")(w, r)
}

func (h *Handler) GetHost(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("host", true)(w, r)
}

func (h *Handler) UpdateHost(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("host", "host.update", "medium")(w, r)
}

func (h *Handler) DeleteHost(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("host", "host.delete", "high")(w, r)
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("device", false)(w, r)
}

func (h *Handler) RunHostAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("host", "host.action", "medium")(w, r)
}

func (h *Handler) IdentifyDevice(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("device", "device.identify", "low")(w, r)
}

func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("service", false)(w, r)
}

func (h *Handler) CreateService(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("service", "service.create", "medium")(w, r)
}

func (h *Handler) GetService(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("service", true)(w, r)
}

func (h *Handler) UpdateService(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("service", "service.update", "medium")(w, r)
}

func (h *Handler) DeleteService(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("service", "service.delete", "high")(w, r)
}

func (h *Handler) ListDaemons(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("daemon", false)(w, r)
}

func (h *Handler) GetDaemon(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("daemon", true)(w, r)
}

func (h *Handler) RunDaemonAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("daemon", "daemon.action", "low")(w, r)
}

func (h *Handler) GetUpgrade(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("upgrade", true)(w, r)
}

func (h *Handler) CheckUpgrade(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("upgrade", "upgrade.check", "low")(w, r)
}

func (h *Handler) RunUpgradeAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("upgrade", "upgrade.action", "high")(w, r)
}

func (h *Handler) ListMonitors(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("mon", false)(w, r)
}

func (h *Handler) RunMonitorAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("mon", "monitor.action", "low")(w, r)
}

func (h *Handler) ListManagers(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("mgr", false)(w, r)
}

func (h *Handler) FailManager(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("mgr", "manager.fail", "medium")(w, r)
}

func (h *Handler) ListManagerModules(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("mgr_module", false)(w, r)
}

func (h *Handler) UpdateManagerModule(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("mgr_module", "manager_module.update", "medium")(w, r)
}

func (h *Handler) ListOSDs(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("osd", false)(w, r)
}

func (h *Handler) GetOSD(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("osd", true)(w, r)
}

func (h *Handler) GetOSDFlag(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("osd_flag", true)(w, r)
}

func (h *Handler) UpdateOSDFlag(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("osd_flag", "osd_flag.update", "medium")(w, r)
}

func (h *Handler) RunOSDAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("osd", "osd.action", "medium")(w, r)
}

func (h *Handler) CheckOSDRemoval(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("osd", "osd.removal_check", "low")(w, r)
}

func (h *Handler) DeleteOSD(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("osd", "osd.delete", "high")(w, r)
}

func (h *Handler) ListOSDRemovals(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("osd_removal", false)(w, r)
}

func (h *Handler) PreviewOSDDeployment(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("osd_deployment", "osd_deployment.preview", "low")(w, r)
}

func (h *Handler) CreateOSDDeployment(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("osd_deployment", "osd_deployment.create", "high")(w, r)
}

func (h *Handler) ZapDevice(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("device", "device.zap", "high")(w, r)
}

func (h *Handler) ListCrushRules(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("crush_rule", false)(w, r)
}

func (h *Handler) CreateCrushRule(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("crush_rule", "crush_rule.create", "medium")(w, r)
}

func (h *Handler) GetCrushRule(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("crush_rule", true)(w, r)
}

func (h *Handler) UpdateCrushRule(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("crush_rule", "crush_rule.update", "medium")(w, r)
}

func (h *Handler) DeleteCrushRule(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("crush_rule", "crush_rule.delete", "high")(w, r)
}

func (h *Handler) ListErasureCodeProfiles(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("erasure_code_profile", false)(w, r)
}

func (h *Handler) CreateErasureCodeProfile(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("erasure_code_profile", "erasure_code_profile.create", "medium")(w, r)
}

func (h *Handler) GetErasureCodeProfile(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("erasure_code_profile", true)(w, r)
}

func (h *Handler) DeleteErasureCodeProfile(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("erasure_code_profile", "erasure_code_profile.delete", "high")(w, r)
}

func (h *Handler) ListPools(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("pool", false)(w, r)
}

func (h *Handler) CreatePool(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("pool", "pool.create", "medium")(w, r)
}

func (h *Handler) GetPool(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("pool", true)(w, r)
}

func (h *Handler) UpdatePool(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("pool", "pool.update", "medium")(w, r)
}

func (h *Handler) DeletePool(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("pool", "pool.delete", "high")(w, r)
}

func (h *Handler) ListRBDImages(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_image", false)(w, r)
}

func (h *Handler) CreateRBDImage(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_image", "rbd_image.create", "medium")(w, r)
}

func (h *Handler) GetRBDImage(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_image", true)(w, r)
}

func (h *Handler) UpdateRBDImage(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_image", "rbd_image.update", "medium")(w, r)
}

func (h *Handler) DeleteRBDImage(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_image", "rbd_image.delete", "high")(w, r)
}

func (h *Handler) RunRBDImageAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_image", "rbd_image.action", "medium")(w, r)
}

func (h *Handler) ListRBDSnapshots(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_snapshot", false)(w, r)
}

func (h *Handler) CreateRBDSnapshot(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_snapshot", "rbd_snapshot.create", "medium")(w, r)
}

func (h *Handler) UpdateRBDSnapshot(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_snapshot", "rbd_snapshot.update", "medium")(w, r)
}

func (h *Handler) DeleteRBDSnapshot(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_snapshot", "rbd_snapshot.delete", "high")(w, r)
}

func (h *Handler) RunRBDSnapshotAction(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_snapshot", "rbd_snapshot.action", "medium")(w, r)
}

func (h *Handler) ListRBDNamespaces(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_namespace", false)(w, r)
}

func (h *Handler) CreateRBDNamespace(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_namespace", "rbd_namespace.create", "medium")(w, r)
}

func (h *Handler) DeleteRBDNamespace(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_namespace", "rbd_namespace.delete", "high")(w, r)
}

func (h *Handler) ListRBDTrash(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_trash", false)(w, r)
}

func (h *Handler) RestoreRBDTrash(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_trash", "rbd_trash.restore", "medium")(w, r)
}

func (h *Handler) DeleteRBDTrash(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_trash", "rbd_trash.delete", "high")(w, r)
}

func (h *Handler) PurgeRBDTrash(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_trash", "rbd_trash.purge", "high")(w, r)
}

func (h *Handler) ListRBDGroups(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_group", false)(w, r)
}

func (h *Handler) CreateRBDGroup(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_group", "rbd_group.create", "medium")(w, r)
}

func (h *Handler) GetRBDMirroring(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rbd_mirroring", true)(w, r)
}

func (h *Handler) UpdateRBDMirroring(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rbd_mirroring", "rbd_mirroring.update", "medium")(w, r)
}

func (h *Handler) ListFilesystems(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("filesystem", false)(w, r)
}

func (h *Handler) CreateFilesystem(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("filesystem", "filesystem.create", "medium")(w, r)
}

func (h *Handler) GetFilesystem(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("filesystem", true)(w, r)
}

func (h *Handler) UpdateFilesystem(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("filesystem", "filesystem.update", "medium")(w, r)
}

func (h *Handler) DeleteFilesystem(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("filesystem", "filesystem.delete", "high")(w, r)
}

func (h *Handler) ListFilesystemClients(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("cephfs_client", false)(w, r)
}

func (h *Handler) EvictFilesystemClient(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("cephfs_client", "cephfs_client.evict", "medium")(w, r)
}

func (h *Handler) ListSubvolumeGroups(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("subvolume_group", false)(w, r)
}

func (h *Handler) CreateSubvolumeGroup(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume_group", "subvolume_group.create", "medium")(w, r)
}

func (h *Handler) GetSubvolumeGroup(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("subvolume_group", true)(w, r)
}

func (h *Handler) UpdateSubvolumeGroup(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume_group", "subvolume_group.update", "medium")(w, r)
}

func (h *Handler) DeleteSubvolumeGroup(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume_group", "subvolume_group.delete", "high")(w, r)
}

func (h *Handler) ListSubvolumes(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("subvolume", false)(w, r)
}

func (h *Handler) CreateSubvolume(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume", "subvolume.create", "medium")(w, r)
}

func (h *Handler) GetSubvolume(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("subvolume", true)(w, r)
}

func (h *Handler) UpdateSubvolume(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume", "subvolume.update", "medium")(w, r)
}

func (h *Handler) DeleteSubvolume(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("subvolume", "subvolume.delete", "high")(w, r)
}

func (h *Handler) ListCephFSSnapshots(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("cephfs_snapshot", false)(w, r)
}

func (h *Handler) CreateCephFSSnapshot(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("cephfs_snapshot", "cephfs_snapshot.create", "medium")(w, r)
}

func (h *Handler) CloneCephFSSnapshot(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("cephfs_snapshot", "cephfs_snapshot.clone", "medium")(w, r)
}

func (h *Handler) ListSnapshotSchedules(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("snapshot_schedule", false)(w, r)
}

func (h *Handler) CreateSnapshotSchedule(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("snapshot_schedule", "snapshot_schedule.create", "medium")(w, r)
}

func (h *Handler) ListCephFSAuthorizations(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("cephfs_authorization", false)(w, r)
}

func (h *Handler) CreateCephFSAuthorization(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("cephfs_authorization", "cephfs_authorization.create", "medium")(w, r)
}

func (h *Handler) ListCephFSEntries(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("cephfs_entry", false)(w, r)
}

func (h *Handler) UpdateCephFSEntryQuota(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("cephfs_entry", "cephfs_entry.quota", "medium")(w, r)
}

func (h *Handler) GetRGWStatus(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_status", true)(w, r)
}

func (h *Handler) ListRGWUsers(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_user", false)(w, r)
}

func (h *Handler) CreateRGWUser(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_user", "rgw_user.create", "medium")(w, r)
}

func (h *Handler) GetRGWUser(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_user", true)(w, r)
}

func (h *Handler) UpdateRGWUser(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_user", "rgw_user.update", "medium")(w, r)
}

func (h *Handler) DeleteRGWUser(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_user", "rgw_user.delete", "high")(w, r)
}

func (h *Handler) CreateRGWUserKey(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_key", "rgw_key.create", "medium")(w, r)
}

func (h *Handler) DeleteRGWUserKey(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_key", "rgw_key.delete", "high")(w, r)
}

func (h *Handler) ListRGWAccounts(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_account", false)(w, r)
}

func (h *Handler) CreateRGWAccount(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_account", "rgw_account.create", "medium")(w, r)
}

func (h *Handler) ListRGWRoles(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_role", false)(w, r)
}

func (h *Handler) CreateRGWRole(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_role", "rgw_role.create", "medium")(w, r)
}

func (h *Handler) ListRGWBuckets(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_bucket", false)(w, r)
}

func (h *Handler) CreateRGWBucket(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_bucket", "rgw_bucket.create", "medium")(w, r)
}

func (h *Handler) GetRGWBucket(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_bucket", true)(w, r)
}

func (h *Handler) UpdateRGWBucket(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_bucket", "rgw_bucket.update", "medium")(w, r)
}

func (h *Handler) DeleteRGWBucket(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_bucket", "rgw_bucket.delete", "high")(w, r)
}

func (h *Handler) GetRGWBucketPolicy(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("rgw_bucket_policy")(w, r)
}

func (h *Handler) UpdateRGWBucketPolicy(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_bucket_policy", "rgw_bucket_policy.update", "medium")(w, r)
}

func (h *Handler) ListRGWRealms(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_realm", false)(w, r)
}

func (h *Handler) CreateRGWRealm(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_realm", "rgw_realm.create", "medium")(w, r)
}

func (h *Handler) ListRGWZonegroups(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_zonegroup", false)(w, r)
}

func (h *Handler) CreateRGWZonegroup(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_zonegroup", "rgw_zonegroup.create", "medium")(w, r)
}

func (h *Handler) ListRGWZones(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("rgw_zone", false)(w, r)
}

func (h *Handler) CreateRGWZone(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_zone", "rgw_zone.create", "medium")(w, r)
}

func (h *Handler) CommitRGWPeriod(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("rgw_period", "rgw_period.commit", "high")(w, r)
}

func (h *Handler) ListNFSClusters(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("nfs_cluster", false)(w, r)
}

func (h *Handler) CreateNFSCluster(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nfs_cluster", "nfs_cluster.create", "medium")(w, r)
}

func (h *Handler) GetNFSCluster(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("nfs_cluster", true)(w, r)
}

func (h *Handler) DeleteNFSCluster(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nfs_cluster", "nfs_cluster.delete", "high")(w, r)
}

func (h *Handler) ListNFSExports(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("nfs_export", false)(w, r)
}

func (h *Handler) CreateNFSExport(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nfs_export", "nfs_export.create", "medium")(w, r)
}

func (h *Handler) GetNFSExport(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("nfs_export", true)(w, r)
}

func (h *Handler) UpdateNFSExport(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nfs_export", "nfs_export.update", "medium")(w, r)
}

func (h *Handler) DeleteNFSExport(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nfs_export", "nfs_export.delete", "high")(w, r)
}

func (h *Handler) ListSMBClusters(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("smb_cluster", false)(w, r)
}

func (h *Handler) CreateSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("smb_cluster", "smb_cluster.create", "medium")(w, r)
}

func (h *Handler) GetSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("smb_cluster", true)(w, r)
}

func (h *Handler) UpdateSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("smb_cluster", "smb_cluster.update", "medium")(w, r)
}

func (h *Handler) DeleteSMBCluster(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("smb_cluster", "smb_cluster.delete", "high")(w, r)
}

func (h *Handler) ListSMBShares(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("smb_share", false)(w, r)
}

func (h *Handler) CreateSMBShare(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("smb_share", "smb_share.create", "medium")(w, r)
}

func (h *Handler) GetSMBShare(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("smb_share", true)(w, r)
}

func (h *Handler) UpdateSMBShare(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("smb_share", "smb_share.update", "medium")(w, r)
}

func (h *Handler) DeleteSMBShare(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("smb_share", "smb_share.delete", "high")(w, r)
}

func (h *Handler) GetNVMeOFGateway(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_gateway")(w, r)
}

func (h *Handler) ListNVMeOFSubsystems(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_subsystem")(w, r)
}

func (h *Handler) CreateNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_subsystem", "nvmeof_subsystem.create", "medium")(w, r)
}

func (h *Handler) GetNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_subsystem")(w, r)
}

func (h *Handler) UpdateNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_subsystem", "nvmeof_subsystem.update", "medium")(w, r)
}

func (h *Handler) DeleteNVMeOFSubsystem(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_subsystem", "nvmeof_subsystem.delete", "high")(w, r)
}

func (h *Handler) ListNVMeOFNamespaces(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_namespace")(w, r)
}

func (h *Handler) CreateNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_namespace", "nvmeof_namespace.create", "medium")(w, r)
}

func (h *Handler) GetNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_namespace")(w, r)
}

func (h *Handler) UpdateNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_namespace", "nvmeof_namespace.update", "medium")(w, r)
}

func (h *Handler) DeleteNVMeOFNamespace(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_namespace", "nvmeof_namespace.delete", "high")(w, r)
}

func (h *Handler) ListNVMeOFListeners(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_listener")(w, r)
}

func (h *Handler) CreateNVMeOFListener(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_listener", "nvmeof_listener.create", "medium")(w, r)
}

func (h *Handler) DeleteNVMeOFListener(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_listener", "nvmeof_listener.delete", "high")(w, r)
}

func (h *Handler) ListNVMeOFHosts(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_host")(w, r)
}

func (h *Handler) CreateNVMeOFHost(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_host", "nvmeof_host.create", "medium")(w, r)
}

func (h *Handler) DeleteNVMeOFHost(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("nvmeof_host", "nvmeof_host.delete", "high")(w, r)
}

func (h *Handler) ListNVMeOFConnections(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("nvmeof_connection")(w, r)
}

func (h *Handler) GetISCSIGateway(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("iscsi_gateway")(w, r)
}

func (h *Handler) ListISCSITargets(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("iscsi_target")(w, r)
}

func (h *Handler) CreateISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("iscsi_target", "iscsi_target.create", "medium")(w, r)
}

func (h *Handler) GetISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("iscsi_target")(w, r)
}

func (h *Handler) UpdateISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("iscsi_target", "iscsi_target.update", "medium")(w, r)
}

func (h *Handler) DeleteISCSITarget(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("iscsi_target", "iscsi_target.delete", "high")(w, r)
}

func (h *Handler) ListConfigurationOptions(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("config_option", false)(w, r)
}

func (h *Handler) ListConfigurationValues(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("config_value", false)(w, r)
}

func (h *Handler) SetConfigurationValue(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("config_value", "config_value.set", "medium")(w, r)
}

func (h *Handler) DeleteConfigurationValue(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("config_value", "config_value.delete", "medium")(w, r)
}

func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	h.ReadResource("log", false)(w, r)
}

func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, r, http.StatusInternalServerError, "stream_unavailable", "streaming is unavailable", false, nil)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) QueryMetric(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("metric")(w, r)
}

func (h *Handler) QueryMetricRange(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("metric")(w, r)
}

func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("alert")(w, r)
}

func (h *Handler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("alert_rule")(w, r)
}

func (h *Handler) ListSilences(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("silence")(w, r)
}

func (h *Handler) CreateSilence(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("silence", "silence.create", "medium")(w, r)
}

func (h *Handler) DeleteSilence(w http.ResponseWriter, r *http.Request) {
	h.MutateResource("silence", "silence.delete", "medium")(w, r)
}

func (h *Handler) GetGrafana(w http.ResponseWriter, r *http.Request) {
	h.ReadExternal("grafana")(w, r)
}
