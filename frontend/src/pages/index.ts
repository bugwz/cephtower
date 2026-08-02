import type { ComponentType } from 'react'
import type { PageKey } from '../navigation'
import { AuditPage } from './audit/AuditPage'
import {
  BlockPoolsPage,
  ImageMirroringPage,
  IscsiPage,
  NvmeConnectionsPage,
  NvmeGatewayPage,
  NvmeHostsPage,
  NvmeListenersPage,
  NvmeNamespacesPage,
  NvmeTcpPage,
  RbdGroupsPage,
  RbdImagesPage,
  RbdNamespacesPage,
  RbdSnapshotsPage,
  RbdTrashPage
} from './block/pages'
import {
  ClusterDetailPage,
  ClusterPage,
  DeviceDetailPage,
  DeviceManagementPage,
  HostDetailPage,
  HostPage,
  MdsManagementPage,
  MgrManagementPage,
  MonDetailPage,
  MonManagementPage,
  OsdManagementPage
} from './cluster/pages'
import {
  CephfsAuthorizationsPage,
  CephfsClientsPage,
  CephfsEntriesPage,
  CephfsPage,
  CephfsSnapshotsPage,
  FilePoolsPage,
  NfsClustersPage,
  NfsPage,
  SmbClustersPage,
  SmbPage,
  SnapshotSchedulesPage,
  SubvolumeGroupsPage,
  SubvolumesPage
} from './file/pages'
import {
  AlertListPage,
  AlertRulesPage,
  AlertSilencesPage,
  MonitorOverviewPage,
  PerformanceMetricsPage,
  RuntimeLogsPage
} from './monitoring/pages'
import {
  BucketManagementPage,
  BucketPolicyPage,
  GatewayManagementPage,
  MultisitePage,
  ObjectStorageConfigPage,
  RgwAccountsPage,
  RgwOverviewPage,
  RgwPeriodPage,
  RgwRolesPage,
  RgwUsersPage,
  RgwZonegroupsPage,
  RgwZonesPage
} from './object/pages'
import { OverviewPage } from './overview/OverviewPage'
import { DataManagementPage, SystemInfoPage, SystemRoleBindingsPage, SystemRolesPage, SystemUsersPage } from './system/pages'

export type { PageKey } from '../navigation'

export { OverviewPage } from './overview/OverviewPage'
export { LoginPage } from './LoginPage'
export { InitializationPage } from './InitializationPage'
export { UserPage } from './system/UserPage'
export { ClusterDetailPage } from './cluster/ClusterDetailPage'
export { HostDetailPage } from './cluster/HostDetailPage'
export { DeviceDetailPage } from './cluster/pages'
export { MonDetailPage } from './cluster/pages'

export const pageComponents: Record<PageKey, ComponentType> = {
  overview: OverviewPage,
  clusterManagement: ClusterPage,
  hostManagement: HostPage,
  monManagement: MonManagementPage,
  mgrManagement: MgrManagementPage,
  osdManagement: OsdManagementPage,
  deviceManagement: DeviceManagementPage,
  mdsManagement: MdsManagementPage,
  blockPools: BlockPoolsPage,
  rbdImages: RbdImagesPage,
  rbdSnapshots: RbdSnapshotsPage,
  rbdNamespaces: RbdNamespacesPage,
  rbdTrash: RbdTrashPage,
  rbdGroups: RbdGroupsPage,
  imageMirroring: ImageMirroringPage,
  iscsi: IscsiPage,
  nvmeGateway: NvmeGatewayPage,
  nvmeTcp: NvmeTcpPage,
  nvmeNamespaces: NvmeNamespacesPage,
  nvmeListeners: NvmeListenersPage,
  nvmeHosts: NvmeHostsPage,
  nvmeConnections: NvmeConnectionsPage,
  filePools: FilePoolsPage,
  cephfs: CephfsPage,
  cephfsClients: CephfsClientsPage,
  subvolumeGroups: SubvolumeGroupsPage,
  subvolumes: SubvolumesPage,
  cephfsSnapshots: CephfsSnapshotsPage,
  snapshotSchedules: SnapshotSchedulesPage,
  cephfsAuthorizations: CephfsAuthorizationsPage,
  cephfsEntries: CephfsEntriesPage,
  nfsClusters: NfsClustersPage,
  nfs: NfsPage,
  smbClusters: SmbClustersPage,
  smb: SmbPage,
  rgwOverview: RgwOverviewPage,
  rgwUsers: RgwUsersPage,
  rgwAccounts: RgwAccountsPage,
  rgwRoles: RgwRolesPage,
  bucketManagement: BucketManagementPage,
  bucketPolicy: BucketPolicyPage,
  gatewayManagement: GatewayManagementPage,
  multisite: MultisitePage,
  rgwZonegroups: RgwZonegroupsPage,
  rgwZones: RgwZonesPage,
  rgwPeriod: RgwPeriodPage,
  objectStorageConfig: ObjectStorageConfigPage,
  monitorOverview: MonitorOverviewPage,
  performanceMetrics: PerformanceMetricsPage,
  runtimeLogs: RuntimeLogsPage,
  alertList: AlertListPage,
  alertRules: AlertRulesPage,
  alertSilences: AlertSilencesPage,
  systemInfo: SystemInfoPage,
  systemUsers: SystemUsersPage,
  systemRoles: SystemRolesPage,
  systemRoleBindings: SystemRoleBindingsPage,
  dataManagement: DataManagementPage,
  auditEvents: AuditPage
}
