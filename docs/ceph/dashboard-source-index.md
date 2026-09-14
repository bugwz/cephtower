# Ceph Dashboard 前端与 CephTower API 源码索引

由 `python3 tools/dashboard-inventory/generate.py` 生成。

此索引列出参考前端服务的声明方法及本项目已注册的 API，供逐项追踪页面 → 服务 →
Dashboard 控制器 → Ceph 命令 → CephTower API 使用。方法名来自静态扫描，可能包括
界面辅助方法；存在方法、路由或命令包装不代表已经完成端到端实现或实机验证。

实现进展和差异见 [dashboard-implementation.md](dashboard-implementation.md)。

## 参考前端服务

| 服务源码 | 声明方法（包含辅助方法） |
| --- | --- |
| [auth.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/auth.service.ts) | `check`, `login`, `logout` |
| [ceph-service.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/ceph-service.service.ts) | `list`, `getDaemons`, `create`, `update`, `delete`, `getKnownTypes` |
| [ceph-user.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/ceph-user.service.ts) | `export` |
| [cephfs-snapshot-schedule.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/cephfs-snapshot-schedule.service.ts) | `create`, `update`, `activate`, `deactivate`, `delete`, `checkScheduleExists`, `checkRetentionPolicyExists`, `getSnapshotSchedule`, `getSnapshotScheduleList`, `parseScheduleCopy`, `parseRetentionCopy` |
| [cephfs-subvolume-group.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/cephfs-subvolume-group.service.ts) | `get`, `create`, `info`, `exists`, `update`, `remove` |
| [cephfs-subvolume.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/cephfs-subvolume.service.ts) | `get`, `create`, `info`, `remove`, `exists`, `existsInFs`, `update`, `getSnapshotVisibility`, `setSnapshotVisibility`, `getSnapshots`, `getSnapshotInfo`, `snapshotExists`, `createSnapshot`, `deleteSnapshot`, `createSnapshotClone` |
| [cephfs.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/cephfs.service.ts) | `list`, `lsDir`, `getCephfs`, `getTabs`, `getClients`, `evictClient`, `getMdsCounters`, `getFsRootDirectory`, `mkSnapshot`, `rmSnapshot`, `quota`, `create`, `isCephFsPool`, `remove`, `rename`, `setAuth`, `getUsedPools` |
| [cluster.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/cluster.service.ts) | `getStatus`, `updateStatus` |
| [configuration.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/configuration.service.ts) | `findValue`, `getValue`, `getConfigData`, `get`, `filter`, `create`, `delete`, `bulkCreate` |
| [crush-rule.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/crush-rule.service.ts) | `create`, `delete`, `getInfo` |
| [custom-login-banner.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/custom-login-banner.service.ts) | `getBannerText` |
| [daemon.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/daemon.service.ts) | `action`, `list` |
| [directory-store.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/directory-store.service.ts) | `loadDirectories`, `search`, `stopPollingDictories` |
| [erasure-code-profile.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/erasure-code-profile.service.ts) | `list`, `create`, `delete`, `getInfo` |
| [feedback.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/feedback.service.ts) | `isKeyExist`, `createIssue` |
| [hardware.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/hardware.service.ts) | `getSummary` |
| [health.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/health.service.ts) | `getFullHealth`, `getMinimalHealth`, `getHealthSnapshot`, `getClusterFsid`, `getOrchestratorName`, `getTelemetryStatus` |
| [host.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/host.service.ts) | `list`, `create`, `delete`, `getDevices`, `getSmartData`, `getDaemons`, `getLabels`, `update`, `identifyDevice`, `getInventoryParams`, `getInventory`, `inventoryList`, `inventoryDeviceList`, `getAllHosts`, `checkHostsFactsAvailable` |
| [iscsi.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/iscsi.service.ts) | `listTargets`, `getTarget`, `updateTarget`, `status`, `settings`, `version`, `portals`, `createTarget`, `deleteTarget`, `getDiscovery`, `updateDiscovery`, `overview` |
| [logging.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/logging.service.ts) | `jsError` |
| [logs.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/logs.service.ts) | `getLogs`, `validateDashboardUrl` |
| [mgr-module.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/mgr-module.service.ts) | `list`, `getConfig`, `updateConfig`, `enable`, `disable`, `getOptions`, `updateModuleState` |
| [monitor.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/monitor.service.ts) | `getMonitor` |
| [motd.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/motd.service.ts) | `get` |
| [multi-cluster.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/multi-cluster.service.ts) | `startPolling`, `startClusterTokenStatusPolling`, `checkAndStartTimer`, `subscribeClusterTokenStatus`, `refresh`, `refreshTokenStatus`, `subscribeOnce`, `subscribe`, `setCluster`, `getCluster`, `deleteCluster`, `editCluster`, `addCluster`, `reConnectCluster`, `getClusterObserver`, `getClusterTokenStatusObserver`, `checkTokenStatus`, `showPrometheusDelayMessage`, `isClusterAdded`, `managePrometheusConnectionError`, `refreshMultiCluster` |
| [nfs.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/nfs.service.ts) | `list`, `get`, `create`, `update`, `delete`, `listClusters`, `lsDir`, `fsals`, `filesystems`, `nfsClusterList` |
| [nvmeof.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/nvmeof.service.ts) | `getAvailableHosts`, `fetchHostsAndGroups`, `getHostsForGroup`, `formatGwGroupsList`, `listGatewayGroups`, `listGateways`, `listSubsystems`, `getSubsystem`, `createSubsystem`, `deleteSubsystem`, `isSubsystemPresent`, `getInitiators`, `addSubsystemInitiators`, `addNamespaceInitiators`, `updateHostKey`, `removeInitiators`, `listListeners`, `createListener`, `createListeners`, `deleteListener`, `listNamespaces`, `getNamespace`, `createNamespace`, `updateNamespace`, `deleteNamespace`, `exists` |
| [orchestrator.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/orchestrator.service.ts) | `status`, `hasFeature`, `getTableActionDisableDesc`, `getName` |
| [osd.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/osd.service.ts) | `create`, `getList`, `getOsdSettings`, `getDetails`, `getSmartData`, `scrub`, `getDeploymentOptions`, `getFlags`, `updateFlags`, `updateIndividualFlags`, `markOut`, `markIn`, `markDown`, `reweight`, `update`, `markLost`, `purge`, `destroy`, `delete`, `safeToDestroy`, `safeToDelete`, `getDevices` |
| [performance-card.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/performance-card.service.ts) | `getChartData`, `buildQueriesForStorageType`, `convertPerformanceData`, `toSeries`, `mergeSeries` |
| [performance-counter.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/performance-counter.service.ts) | `list`, `get` |
| [pool.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/pool.service.ts) | `create`, `update`, `delete`, `get`, `getList`, `getConfiguration`, `getInfo`, `list` |
| [prometheus.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/prometheus.service.ts) | `unsubscribe`, `getPrometheusData`, `getPrometheusQueryData`, `ifAlertmanagerConfigured`, `disableAlertmanagerConfig`, `ifPrometheusConfigured`, `disablePrometheusConfig`, `getAlerts`, `getGroupedAlerts`, `getSilences`, `getRules`, `setSilence`, `expireSilence`, `getNotifications`, `ifSettingConfigured`, `disableSetting`, `getSettingsValue`, `getGaugeQueryData`, `formatGuageMetric`, `updateTimeStamp`, `getMultiClusterData`, `getMultiClusterQueryRangeData`, `getMultiClusterQueriesData`, `getRangeQueriesData` |
| [rbd-mirroring.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rbd-mirroring.service.ts) | `startPolling`, `refresh`, `retrieveSummaryObservable`, `retrieveSummaryObserver`, `subscribeSummary`, `getPool`, `updatePool`, `getSiteName`, `setSiteName`, `createBootstrapToken`, `importBootstrapToken`, `getPeer`, `getPeerForPool`, `addPeer`, `updatePeer`, `deletePeer` |
| [rbd.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rbd.service.ts) | `isRBDPool`, `create`, `delete`, `update`, `get`, `list`, `copy`, `flatten`, `defaultFeatures`, `cloneFormatVersion`, `createSnapshot`, `renameSnapshot`, `protectSnapshot`, `rollbackSnapshot`, `cloneSnapshot`, `deleteSnapshot`, `listTrash`, `createNamespace`, `listNamespaces`, `deleteNamespace`, `moveTrash`, `purgeTrash`, `restoreTrash`, `removeTrash` |
| [rgw-bucket.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-bucket.service.ts) | `fetchAndTransformBuckets`, `transformBucket`, `calculateSizeUsage`, `calculateObjectUsage`, `calculateAverageObjectSize`, `list`, `get`, `getTotalBucketsAndUsersLength`, `create`, `update`, `delete`, `exists`, `getLockDays`, `setEncryptionConfig`, `getEncryption`, `deleteEncryption`, `getEncryptionConfig`, `setLifecycle`, `getLifecycle`, `updateBucketRateLimit`, `getBucketRateLimit`, `getGlobalBucketRateLimit`, `listNotification`, `setNotification`, `deleteNotification` |
| [rgw-daemon.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-daemon.service.ts) | `list`, `get`, `selectDaemon`, `selectDefaultDaemon`, `request`, `setMultisiteConfig` |
| [rgw-multisite.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-multisite.service.ts) | `migrate`, `getSyncStatus`, `status`, `getSyncPolicy`, `getSyncPolicyGroup`, `createSyncPolicyGroup`, `modifySyncPolicyGroup`, `removeSyncPolicyGroup`, `setUpMultisiteReplication`, `createEditSyncFlow`, `removeSyncFlow`, `createEditSyncPipe`, `removeSyncPipe`, `setRestartGatewayMessage`, `getRgwModuleStatus` |
| [rgw-realm.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-realm.service.ts) | `create`, `update`, `list`, `get`, `getAllRealmsInfo`, `delete`, `getRealmTree`, `importRealmToken`, `getRealmTokens` |
| [rgw-site.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-site.service.ts) | `get`, `isDefaultRealm` |
| [rgw-storage-class.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-storage-class.service.ts) | `removeStorageClass`, `createStorageClass`, `editStorageClass`, `getPlacement_target`, `createStorageClassZone`, `editStorageClassZone` |
| [rgw-topic.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-topic.service.ts) | `listTopic`, `getTopic`, `create`, `delete`, `exists` |
| [rgw-user-accounts.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-user-accounts.service.ts) | `list`, `get`, `create`, `modify`, `remove`, `setQuota` |
| [rgw-user.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-user.service.ts) | `list`, `enumerate`, `enumerateEmail`, `get`, `getQuota`, `create`, `update`, `updateQuota`, `delete`, `createSubuser`, `deleteSubuser`, `addCapability`, `deleteCapability`, `addS3Key`, `deleteS3Key`, `exists`, `emailExists`, `updateUserRateLimit`, `getUserRateLimit`, `getGlobalUserRateLimit` |
| [rgw-zone.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-zone.service.ts) | `create`, `list`, `get`, `getAllZonesInfo`, `delete`, `update`, `getZoneTree`, `getPoolNames`, `createSystemUser`, `getUserList` |
| [rgw-zonegroup.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/rgw-zonegroup.service.ts) | `create`, `update`, `list`, `get`, `getAllZonegroupsInfo`, `delete`, `getZonegroupTree` |
| [role.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/role.service.ts) | `list`, `delete`, `get`, `create`, `clone`, `update`, `exists` |
| [scope.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/scope.service.ts) | `list` |
| [settings.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/settings.service.ts) | `getValues`, `ifSettingConfigured`, `disableSetting`, `getSettingsValue`, `validateGrafanaDashboardUrl`, `getStandardSettings` |
| [smb.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/smb.service.ts) | `passData`, `setDataUploaded`, `uploadData`, `listClusters`, `createCluster`, `removeCluster`, `listShares`, `listJoinAuths`, `listUsersGroups`, `createShare`, `getShare`, `deleteShare`, `getJoinAuth`, `getUsersGroups`, `createJoinAuth`, `createUsersGroups`, `deleteJoinAuth`, `deleteUsersgroups`, `getCluster` |
| [storage-overview.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/storage-overview.service.ts) | `getTrendData`, `getAverageConsumption`, `getTimeUntilFull`, `getTopPools`, `getCount`, `getObjectCounts`, `getRawCapacityThresholds`, `getStorageBreakdown`, `getThresholdStatus`, `convertBytesToUnit`, `formatBytesForChart`, `normalizeGroup`, `isAStorage`, `mapStorageChartData` |
| [telemetry.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/telemetry.service.ts) | `getReport`, `enable` |
| [upgrade.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/upgrade.service.ts) | `list`, `versionAvailableForUpgrades`, `start`, `pause`, `resume`, `stop`, `status`, `listCached`, `startUpgradeModal` |
| [user.service](../../docs/references/ceph/src/pybind/mgr/dashboard/frontend/src/app/shared/api/user.service.ts) | `list`, `delete`, `get`, `create`, `update`, `changePassword`, `validateUserName`, `validatePassword` |

## 参考控制器

| 源码 | API 路由声明 |
| --- | --- |
| [auth](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/auth.py) | `/auth` |
| [ceph_users](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/ceph_users.py) | `/cluster/user` |
| [cephfs](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/cephfs.py) | `/cephfs`, `/cephfs/subvolume`, `/cephfs/subvolume/group`, `/cephfs/subvolume/snapshot`, `/cephfs/subvolume/snapshot/clone`, `/cephfs/snapshot/schedule` |
| [cluster](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/cluster.py) | `/cluster`, `/cluster/upgrade` |
| [cluster_configuration](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/cluster_configuration.py) | `/cluster_conf` |
| [crush_rule](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/crush_rule.py) | `/crush_rule` |
| [daemon](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/daemon.py) | `/daemon` |
| [docs](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/docs.py) |  |
| [erasure_code_profile](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/erasure_code_profile.py) | `/erasure_code_profile` |
| [feedback](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/feedback.py) | `/feedback`, `/feedback/api_key` |
| [frontend_logging](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/frontend_logging.py) | `/logging` |
| [grafana](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/grafana.py) | `/grafana` |
| [hardware](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/hardware.py) | `/hardware` |
| [health](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/health.py) | `/health` |
| [home](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/home.py) | `/langs`, `/login` |
| [host](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/host.py) | `/host` |
| [iscsi](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/iscsi.py) | `/iscsi`, `/iscsi/target` |
| [logs](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/logs.py) | `/logs` |
| [mgr_modules](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/mgr_modules.py) | `/mgr/module` |
| [monitor](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/monitor.py) | `/monitor` |
| [multi_cluster](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/multi_cluster.py) | `/multi-cluster` |
| [nfs](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/nfs.py) | `/nfs-ganesha/cluster`, `/nfs-ganesha/export`, `/nfs-ganesha` |
| [nvmeof](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/nvmeof.py) | `/nvmeof/gateway`, `/nvmeof/spdk`, `/nvmeof/subsystem`, `/nvmeof/subsystem/{nqn}/listener`, `/nvmeof/subsystem/{nqn}/namespace`, `/nvmeof/subsystem/{nqn}/host`, `/nvmeof/subsystem/{nqn}/connection`, `/nvmeof` |
| [oauth2](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/oauth2.py) |  |
| [orchestrator](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/orchestrator.py) | `/orchestrator` |
| [osd](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/osd.py) | `/osd`, `/osd/flags` |
| [perf_counters](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/perf_counters.py) | `perf_counters/mds`, `perf_counters/mon`, `perf_counters/osd`, `perf_counters/rgw`, `perf_counters/rbd-mirror`, `perf_counters/mgr`, `perf_counters/tcmu-runner`, `perf_counters` |
| [pool](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/pool.py) | `/pool` |
| [prometheus](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/prometheus.py) | `/prometheus`, `/prometheus/notifications` |
| [rbd](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/rbd.py) | `/block/image`, `/block/rbd`, `/block/image/{image_spec}/snap`, `/block/image/trash`, `/block/pool/{pool_name}/namespace`, `/block/pool/{pool_name}/group`, `/block/pool/{pool_name}/group/{group_name}/snap` |
| [rbd_mirroring](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/rbd_mirroring.py) | `/block/mirroring`, `/block/mirroring/summary`, `/block/mirroring/{pool_name}/{image_name}/summary`, `/block/mirroring/pool`, `/block/mirroring/pool/{pool_name}/bootstrap`, `/block/mirroring/pool/{pool_name}/peer` |
| [rgw](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/rgw.py) | `/rgw`, `/rgw/multisite`, `rgw/multisite`, `/rgw/daemon`, `/rgw/site`, `/rgw/bucket`, `/rgw/user`, `/rgw/roles`, `/rgw/realm`, `/rgw/zonegroup`, `/rgw/zone`, `/rgw/topic` |
| [rgw_iam](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/rgw_iam.py) | `rgw/accounts` |
| [role](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/role.py) | `/role`, `/scope` |
| [saml2](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/saml2.py) |  |
| [service](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/service.py) | `/service` |
| [settings](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/settings.py) | `/settings`, `/standard_settings` |
| [smb](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/smb.py) | `/smb/cluster`, `/smb/share`, `/smb/joinauth`, `/smb/usersgroups`, `/smb` |
| [summary](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/summary.py) | `/summary` |
| [task](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/task.py) | `/task` |
| [telemetry](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/telemetry.py) | `/telemetry` |
| [user](../../docs/references/ceph/src/pybind/mgr/dashboard/controllers/user.py) | `/user`, `/user/{username}` |

## CephTower 已注册 API

| 路由源码 | 方法和路径 | Handler |
| --- | --- | --- |
| [alert.go](../../backend/internal/api/v1/router/alert.go) | `GET /api/v1/alert/alerts` | `ListAlerts` |
| [alert.go](../../backend/internal/api/v1/router/alert.go) | `GET /api/v1/alert/rules` | `ListAlertRules` |
| [alert.go](../../backend/internal/api/v1/router/alert.go) | `GET /api/v1/alert/silences` | `ListSilences` |
| [alert.go](../../backend/internal/api/v1/router/alert.go) | `POST /api/v1/alert/silence` | `CreateSilence` |
| [alert.go](../../backend/internal/api/v1/router/alert.go) | `DELETE /api/v1/alert/silence` | `DeleteSilence` |
| [audit.go](../../backend/internal/api/v1/router/audit.go) | `GET /api/v1/audit/events` | `ListAuditEvents` |
| [auth.go](../../backend/internal/api/v1/router/auth.go) | `GET /api/v1/bootstrap` | `BootstrapStatus` |
| [auth.go](../../backend/internal/api/v1/router/auth.go) | `POST /api/v1/bootstrap/dbtest` | `TestBootstrapDatabase` |
| [auth.go](../../backend/internal/api/v1/router/auth.go) | `POST /api/v1/bootstrap/run` | `BootstrapRun` |
| [auth.go](../../backend/internal/api/v1/router/auth.go) | `POST /api/v1/auth/login` | `Login` |
| [auth.go](../../backend/internal/api/v1/router/auth.go) | `GET /api/v1/user` | `ListUsers` |
| [auth.go](../../backend/internal/api/v1/router/auth.go) | `POST /api/v1/user` | `CreateUser` |
| [ceph_user.go](../../backend/internal/api/v1/router/ceph_user.go) | `GET /api/v1/ceph/users` | `ListCephUsers` |
| [ceph_user.go](../../backend/internal/api/v1/router/ceph_user.go) | `POST /api/v1/ceph/user` | `CreateCephUser` |
| [ceph_user.go](../../backend/internal/api/v1/router/ceph_user.go) | `PATCH /api/v1/ceph/user` | `UpdateCephUser` |
| [ceph_user.go](../../backend/internal/api/v1/router/ceph_user.go) | `DELETE /api/v1/ceph/user` | `DeleteCephUser` |
| [ceph_user.go](../../backend/internal/api/v1/router/ceph_user.go) | `POST /api/v1/ceph/users/import` | `ImportCephUsers` |
| [ceph_user.go](../../backend/internal/api/v1/router/ceph_user.go) | `GET /api/v1/ceph/users/export` | `ExportCephUsers` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystems` | `ListFilesystems` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem` | `CreateFilesystem` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem` | `GetFilesystem` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `PATCH /api/v1/filesystem` | `UpdateFilesystem` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `DELETE /api/v1/filesystem` | `DeleteFilesystem` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/clients` | `ListFilesystemClients` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `DELETE /api/v1/filesystem/client` | `EvictFilesystemClient` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/subvolume/groups` | `ListSubvolumeGroups` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/subvolume/group` | `CreateSubvolumeGroup` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/subvolume/group` | `GetSubvolumeGroup` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `PATCH /api/v1/filesystem/subvolume/group` | `UpdateSubvolumeGroup` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `DELETE /api/v1/filesystem/subvolume/group` | `DeleteSubvolumeGroup` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/subvolumes` | `ListSubvolumes` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/subvolume` | `CreateSubvolume` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/subvolume` | `GetSubvolume` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `PATCH /api/v1/filesystem/subvolume` | `UpdateSubvolume` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `DELETE /api/v1/filesystem/subvolume` | `DeleteSubvolume` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/subvolume/snapshots` | `ListCephFSSnapshots` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/subvolume/snapshot` | `CreateCephFSSnapshot` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `DELETE /api/v1/filesystem/subvolume/snapshot` | `DeleteCephFSSnapshot` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/subvolume/snapshot/clone` | `CloneCephFSSnapshot` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/snapshot/schedule/status` | `GetSnapshotScheduleStatus` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/snapshot/schedule` | `CreateSnapshotSchedule` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/snapshot/schedule/action` | `RunSnapshotScheduleAction` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/snapshot/schedule/retention` | `UpdateSnapshotRetention` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/authorizations` | `ListCephFSAuthorizations` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `POST /api/v1/filesystem/authorization` | `CreateCephFSAuthorization` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `GET /api/v1/filesystem/entries` | `ListCephFSEntries` |
| [cephfs.go](../../backend/internal/api/v1/router/cephfs.go) | `PATCH /api/v1/filesystem/entry/quota` | `UpdateCephFSEntryQuota` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `GET /api/v1/clusters` | `ListClusters` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `POST /api/v1/cluster` | `CreateCluster` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `GET /api/v1/cluster` | `GetCluster` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `PATCH /api/v1/cluster` | `UpdateCluster` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `DELETE /api/v1/cluster` | `DeleteCluster` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `POST /api/v1/cluster/probe` | `ProbeCluster` |
| [cluster.go](../../backend/internal/api/v1/router/cluster.go) | `GET /api/v1/cluster/capabilities` | `Capabilities` |
| [configuration.go](../../backend/internal/api/v1/router/configuration.go) | `GET /api/v1/configuration/option` | `GetConfigurationOption` |
| [configuration.go](../../backend/internal/api/v1/router/configuration.go) | `GET /api/v1/configuration/options` | `ListConfigurationOptions` |
| [configuration.go](../../backend/internal/api/v1/router/configuration.go) | `GET /api/v1/configuration/values` | `ListConfigurationValues` |
| [configuration.go](../../backend/internal/api/v1/router/configuration.go) | `PUT /api/v1/configuration/value` | `SetConfigurationValue` |
| [configuration.go](../../backend/internal/api/v1/router/configuration.go) | `DELETE /api/v1/configuration/value` | `DeleteConfigurationValue` |
| [credential.go](../../backend/internal/api/v1/router/credential.go) | `GET /api/v1/credentials` | `ListCredentials` |
| [credential.go](../../backend/internal/api/v1/router/credential.go) | `PUT /api/v1/credential` | `PutCredential` |
| [credential.go](../../backend/internal/api/v1/router/credential.go) | `DELETE /api/v1/credential` | `DeleteCredential` |
| [crush.go](../../backend/internal/api/v1/router/crush.go) | `GET /api/v1/crush/rules` | `ListCrushRules` |
| [crush.go](../../backend/internal/api/v1/router/crush.go) | `POST /api/v1/crush/rule` | `CreateCrushRule` |
| [crush.go](../../backend/internal/api/v1/router/crush.go) | `GET /api/v1/crush/rule` | `GetCrushRule` |
| [crush.go](../../backend/internal/api/v1/router/crush.go) | `PATCH /api/v1/crush/rule` | `UpdateCrushRule` |
| [crush.go](../../backend/internal/api/v1/router/crush.go) | `DELETE /api/v1/crush/rule` | `DeleteCrushRule` |
| [daemon.go](../../backend/internal/api/v1/router/daemon.go) | `GET /api/v1/daemons` | `ListDaemons` |
| [daemon.go](../../backend/internal/api/v1/router/daemon.go) | `GET /api/v1/daemon` | `GetDaemon` |
| [daemon.go](../../backend/internal/api/v1/router/daemon.go) | `POST /api/v1/daemon/action` | `RunDaemonAction` |
| [device.go](../../backend/internal/api/v1/router/device.go) | `GET /api/v1/devices` | `ListDevices` |
| [device.go](../../backend/internal/api/v1/router/device.go) | `POST /api/v1/device/identify` | `IdentifyDevice` |
| [device.go](../../backend/internal/api/v1/router/device.go) | `POST /api/v1/device/zap` | `ZapDevice` |
| [endpoint.go](../../backend/internal/api/v1/router/endpoint.go) | `GET /api/v1/endpoints` | `ListEndpoints` |
| [endpoint.go](../../backend/internal/api/v1/router/endpoint.go) | `POST /api/v1/endpoint` | `CreateEndpoint` |
| [endpoint.go](../../backend/internal/api/v1/router/endpoint.go) | `PATCH /api/v1/endpoint` | `UpdateEndpoint` |
| [endpoint.go](../../backend/internal/api/v1/router/endpoint.go) | `DELETE /api/v1/endpoint` | `DeleteEndpoint` |
| [erasure_code_profile.go](../../backend/internal/api/v1/router/erasure_code_profile.go) | `GET /api/v1/erasure/code/profiles` | `ListErasureCodeProfiles` |
| [erasure_code_profile.go](../../backend/internal/api/v1/router/erasure_code_profile.go) | `POST /api/v1/erasure/code/profile` | `CreateErasureCodeProfile` |
| [erasure_code_profile.go](../../backend/internal/api/v1/router/erasure_code_profile.go) | `GET /api/v1/erasure/code/profile` | `GetErasureCodeProfile` |
| [erasure_code_profile.go](../../backend/internal/api/v1/router/erasure_code_profile.go) | `DELETE /api/v1/erasure/code/profile` | `DeleteErasureCodeProfile` |
| [grafana.go](../../backend/internal/api/v1/router/grafana.go) | `GET /api/v1/grafana` | `GetGrafana` |
| [health.go](../../backend/internal/api/v1/router/health.go) | `GET /api/v1/healthz` | `Health` |
| [health.go](../../backend/internal/api/v1/router/health.go) | `GET /api/v1/readyz` | `Ready` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `GET /api/v1/hosts` | `ListHosts` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `POST /api/v1/host` | `CreateHost` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `GET /api/v1/host` | `GetHost` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `GET /api/v1/host/devices` | `GetHostDevices` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `GET /api/v1/host/smart` | `GetHostSMART` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `PATCH /api/v1/host` | `UpdateHost` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `DELETE /api/v1/host` | `DeleteHost` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `GET /api/v1/host/ssh` | `GetHostSSH` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `PATCH /api/v1/host/ssh` | `SaveHostSSH` |
| [host.go](../../backend/internal/api/v1/router/host.go) | `POST /api/v1/host/action` | `RunHostAction` |
| [iscsi.go](../../backend/internal/api/v1/router/iscsi.go) | `GET /api/v1/iscsi/gateway` | `GetISCSIGateway` |
| [iscsi.go](../../backend/internal/api/v1/router/iscsi.go) | `GET /api/v1/iscsi/targets` | `ListISCSITargets` |
| [iscsi.go](../../backend/internal/api/v1/router/iscsi.go) | `POST /api/v1/iscsi/target` | `CreateISCSITarget` |
| [iscsi.go](../../backend/internal/api/v1/router/iscsi.go) | `GET /api/v1/iscsi/target` | `GetISCSITarget` |
| [iscsi.go](../../backend/internal/api/v1/router/iscsi.go) | `PATCH /api/v1/iscsi/target` | `UpdateISCSITarget` |
| [iscsi.go](../../backend/internal/api/v1/router/iscsi.go) | `DELETE /api/v1/iscsi/target` | `DeleteISCSITarget` |
| [logs.go](../../backend/internal/api/v1/router/logs.go) | `GET /api/v1/logs` | `ListLogs` |
| [manager.go](../../backend/internal/api/v1/router/manager.go) | `GET /api/v1/managers` | `ListManagers` |
| [manager.go](../../backend/internal/api/v1/router/manager.go) | `POST /api/v1/manager/fail` | `FailManager` |
| [manager_module.go](../../backend/internal/api/v1/router/manager_module.go) | `GET /api/v1/manager/modules` | `ListManagerModules` |
| [manager_module.go](../../backend/internal/api/v1/router/manager_module.go) | `PATCH /api/v1/manager/module` | `UpdateManagerModule` |
| [metric.go](../../backend/internal/api/v1/router/metric.go) | `GET /api/v1/metric/query` | `QueryMetric` |
| [metric.go](../../backend/internal/api/v1/router/metric.go) | `GET /api/v1/metric/range` | `QueryMetricRange` |
| [monitor.go](../../backend/internal/api/v1/router/monitor.go) | `GET /api/v1/monitors` | `ListMonitors` |
| [monitor.go](../../backend/internal/api/v1/router/monitor.go) | `GET /api/v1/monitor/status` | `GetMonitorStatus` |
| [monitor.go](../../backend/internal/api/v1/router/monitor.go) | `GET /api/v1/monitor/perf/counters` | `ListMonitorPerfCounters` |
| [monitor.go](../../backend/internal/api/v1/router/monitor.go) | `POST /api/v1/monitor/action` | `RunMonitorAction` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `GET /api/v1/nfs/clusters` | `ListNFSClusters` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `POST /api/v1/nfs/cluster` | `CreateNFSCluster` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `GET /api/v1/nfs/cluster` | `GetNFSCluster` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `DELETE /api/v1/nfs/cluster` | `DeleteNFSCluster` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `GET /api/v1/nfs/exports` | `ListNFSExports` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `POST /api/v1/nfs/export` | `CreateNFSExport` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `GET /api/v1/nfs/export` | `GetNFSExport` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `PATCH /api/v1/nfs/export` | `UpdateNFSExport` |
| [nfs.go](../../backend/internal/api/v1/router/nfs.go) | `DELETE /api/v1/nfs/export` | `DeleteNFSExport` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/gateway` | `GetNVMeOFGateway` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystems` | `ListNVMeOFSubsystems` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `POST /api/v1/nvmeof/subsystem` | `CreateNVMeOFSubsystem` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystem` | `GetNVMeOFSubsystem` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `PATCH /api/v1/nvmeof/subsystem` | `UpdateNVMeOFSubsystem` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `DELETE /api/v1/nvmeof/subsystem` | `DeleteNVMeOFSubsystem` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystem/namespaces` | `ListNVMeOFNamespaces` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `POST /api/v1/nvmeof/subsystem/namespace` | `CreateNVMeOFNamespace` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystem/namespace` | `GetNVMeOFNamespace` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `PATCH /api/v1/nvmeof/subsystem/namespace` | `UpdateNVMeOFNamespace` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `DELETE /api/v1/nvmeof/subsystem/namespace` | `DeleteNVMeOFNamespace` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystem/listeners` | `ListNVMeOFListeners` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `POST /api/v1/nvmeof/subsystem/listener` | `CreateNVMeOFListener` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `DELETE /api/v1/nvmeof/subsystem/listener` | `DeleteNVMeOFListener` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystem/hosts` | `ListNVMeOFHosts` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `POST /api/v1/nvmeof/subsystem/host` | `CreateNVMeOFHost` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `DELETE /api/v1/nvmeof/subsystem/host` | `DeleteNVMeOFHost` |
| [nvmeof.go](../../backend/internal/api/v1/router/nvmeof.go) | `GET /api/v1/nvmeof/subsystem/connections` | `ListNVMeOFConnections` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `GET /api/v1/osds` | `ListOSDs` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `GET /api/v1/osd` | `GetOSD` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `GET /api/v1/osd/inspection` | `GetOSDInspection` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `GET /api/v1/osd/flag` | `GetOSDFlag` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `PATCH /api/v1/osd/flag` | `UpdateOSDFlag` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `POST /api/v1/osd/action` | `RunOSDAction` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `POST /api/v1/osd/removal/check` | `CheckOSDRemoval` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `DELETE /api/v1/osd` | `DeleteOSD` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `GET /api/v1/osd/removals` | `ListOSDRemovals` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `POST /api/v1/osd/deployment/preview` | `PreviewOSDDeployment` |
| [osd.go](../../backend/internal/api/v1/router/osd.go) | `POST /api/v1/osd/deployment` | `CreateOSDDeployment` |
| [overview.go](../../backend/internal/api/v1/router/overview.go) | `GET /api/v1/overview` | `GetOverview` |
| [overview.go](../../backend/internal/api/v1/router/overview.go) | `GET /api/v1/health` | `ListHealthChecks` |
| [overview.go](../../backend/internal/api/v1/router/overview.go) | `POST /api/v1/health/mute` | `MuteHealthCheck` |
| [overview.go](../../backend/internal/api/v1/router/overview.go) | `DELETE /api/v1/health/mute` | `UnmuteHealthCheck` |
| [pool.go](../../backend/internal/api/v1/router/pool.go) | `GET /api/v1/pools` | `ListPools` |
| [pool.go](../../backend/internal/api/v1/router/pool.go) | `POST /api/v1/pool` | `CreatePool` |
| [pool.go](../../backend/internal/api/v1/router/pool.go) | `GET /api/v1/pool` | `GetPool` |
| [pool.go](../../backend/internal/api/v1/router/pool.go) | `PATCH /api/v1/pool` | `UpdatePool` |
| [pool.go](../../backend/internal/api/v1/router/pool.go) | `DELETE /api/v1/pool` | `DeletePool` |
| [rbac.go](../../backend/internal/api/v1/router/rbac.go) | `GET /api/v1/role` | `ListRoles` |
| [rbac.go](../../backend/internal/api/v1/router/rbac.go) | `POST /api/v1/role` | `CreateRole` |
| [rbac.go](../../backend/internal/api/v1/router/rbac.go) | `GET /api/v1/role/bindings` | `ListRoleBindings` |
| [rbac.go](../../backend/internal/api/v1/router/rbac.go) | `POST /api/v1/role/binding` | `CreateRoleBinding` |
| [rbac.go](../../backend/internal/api/v1/router/rbac.go) | `DELETE /api/v1/role/binding` | `DeleteRoleBinding` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/images` | `ListRBDImages` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/image` | `CreateRBDImage` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/image` | `GetRBDImage` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `PATCH /api/v1/rbd/image` | `UpdateRBDImage` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `DELETE /api/v1/rbd/image` | `DeleteRBDImage` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/image/action` | `RunRBDImageAction` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/image/snapshots` | `ListRBDSnapshots` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/image/snapshot` | `CreateRBDSnapshot` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `PATCH /api/v1/rbd/image/snapshot` | `UpdateRBDSnapshot` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `DELETE /api/v1/rbd/image/snapshot` | `DeleteRBDSnapshot` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/image/snapshot/action` | `RunRBDSnapshotAction` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/namespaces` | `ListRBDNamespaces` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/namespace` | `CreateRBDNamespace` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `DELETE /api/v1/rbd/namespace` | `DeleteRBDNamespace` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/trash` | `ListRBDTrash` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/trash/restore` | `RestoreRBDTrash` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `DELETE /api/v1/rbd/trash` | `DeleteRBDTrash` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/trash/purge` | `PurgeRBDTrash` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/groups` | `ListRBDGroups` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/group` | `CreateRBDGroup` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/group/action` | `RunRBDGroupAction` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/group/member` | `UpdateRBDGroupMember` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/group/snapshot` | `CreateRBDGroupSnapshot` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `GET /api/v1/rbd/mirroring` | `GetRBDMirroring` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `PATCH /api/v1/rbd/mirroring` | `UpdateRBDMirroring` |
| [rbd.go](../../backend/internal/api/v1/router/rbd.go) | `POST /api/v1/rbd/mirroring/peer` | `UpdateRBDMirroringPeer` |
| [resource.go](../../backend/internal/api/v1/router/resource.go) | `POST /api/v1/resource/refresh` | `RefreshResource` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/status` | `GetRGWStatus` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/users` | `ListRGWUsers` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/user` | `CreateRGWUser` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/user` | `GetRGWUser` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/user` | `UpdateRGWUser` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PUT /api/v1/rgw/user/quota` | `UpdateRGWUserQuota` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/user/caps` | `MutateRGWUserCaps` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PUT /api/v1/rgw/user/ratelimit` | `UpdateRGWUserRateLimit` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `DELETE /api/v1/rgw/user` | `DeleteRGWUser` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/user/key` | `CreateRGWUserKey` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `DELETE /api/v1/rgw/user/key` | `DeleteRGWUserKey` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/accounts` | `ListRGWAccounts` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/account` | `CreateRGWAccount` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/account` | `UpdateRGWAccount` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PUT /api/v1/rgw/account/quota` | `UpdateRGWAccountQuota` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `DELETE /api/v1/rgw/account` | `DeleteRGWAccount` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/roles` | `ListRGWRoles` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/role` | `CreateRGWRole` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/role` | `UpdateRGWRole` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/role/policy` | `MutateRGWRolePolicy` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `DELETE /api/v1/rgw/role` | `DeleteRGWRole` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/buckets` | `ListRGWBuckets` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PUT /api/v1/rgw/bucket/ratelimit` | `UpdateRGWBucketRateLimit` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PUT /api/v1/rgw/bucket/quota` | `UpdateRGWBucketQuota` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/bucket` | `CreateRGWBucket` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/bucket` | `GetRGWBucket` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/bucket` | `UpdateRGWBucket` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `DELETE /api/v1/rgw/bucket` | `DeleteRGWBucket` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/bucket/policy` | `GetRGWBucketPolicy` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/bucket/policy` | `UpdateRGWBucketPolicy` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/realms` | `ListRGWRealms` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/realm` | `CreateRGWRealm` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/realm` | `UpdateRGWRealm` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/zonegroups` | `ListRGWZonegroups` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/zonegroup` | `CreateRGWZonegroup` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `PATCH /api/v1/rgw/zonegroup` | `UpdateRGWZonegroup` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `GET /api/v1/rgw/zones` | `ListRGWZones` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/zone` | `CreateRGWZone` |
| [rgw.go](../../backend/internal/api/v1/router/rgw.go) | `POST /api/v1/rgw/period/commit` | `CommitRGWPeriod` |
| [service.go](../../backend/internal/api/v1/router/service.go) | `GET /api/v1/services` | `ListServices` |
| [service.go](../../backend/internal/api/v1/router/service.go) | `POST /api/v1/service` | `CreateService` |
| [service.go](../../backend/internal/api/v1/router/service.go) | `GET /api/v1/service` | `GetService` |
| [service.go](../../backend/internal/api/v1/router/service.go) | `PATCH /api/v1/service` | `UpdateService` |
| [service.go](../../backend/internal/api/v1/router/service.go) | `DELETE /api/v1/service` | `DeleteService` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `GET /api/v1/smb/clusters` | `ListSMBClusters` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `POST /api/v1/smb/cluster` | `CreateSMBCluster` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `GET /api/v1/smb/cluster` | `GetSMBCluster` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `PATCH /api/v1/smb/cluster` | `UpdateSMBCluster` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `DELETE /api/v1/smb/cluster` | `DeleteSMBCluster` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `GET /api/v1/smb/shares` | `ListSMBShares` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `POST /api/v1/smb/share` | `CreateSMBShare` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `GET /api/v1/smb/share` | `GetSMBShare` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `PATCH /api/v1/smb/share` | `UpdateSMBShare` |
| [smb.go](../../backend/internal/api/v1/router/smb.go) | `DELETE /api/v1/smb/share` | `DeleteSMBShare` |
| [upgrade.go](../../backend/internal/api/v1/router/upgrade.go) | `GET /api/v1/upgrade` | `GetUpgrade` |
| [upgrade.go](../../backend/internal/api/v1/router/upgrade.go) | `POST /api/v1/upgrade/check` | `CheckUpgrade` |
| [upgrade.go](../../backend/internal/api/v1/router/upgrade.go) | `POST /api/v1/upgrade/action` | `RunUpgradeAction` |
