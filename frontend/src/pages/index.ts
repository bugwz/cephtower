import type { ComponentType } from 'react'
import type { PageKey } from '../navigation'
import { BlockPoolsPage, ImageMirroringPage, IscsiPage, NvmeTcpPage, RbdImagesPage } from './storage-block/pages'
import {
  ClusterManagementPage,
  HostManagementPage,
  MdsManagementPage,
  MgrManagementPage,
  MonManagementPage,
  OsdManagementPage
} from './cluster-management/pages'
import { CephfsPage, FilePoolsPage, NfsPage, SmbPage } from './storage-file/pages'
import {
  AlertListPage,
  AlertRulesPage,
  AlertSilencesPage,
  MonitorOverviewPage,
  PerformanceMetricsPage,
  RuntimeLogsPage
} from './monitoring-alerting/pages'
import {
  BucketManagementPage,
  GatewayManagementPage,
  MultisitePage,
  ObjectStorageConfigPage,
  RgwOverviewPage,
  RgwUsersPage
} from './storage-object/pages'
import { OverviewPage } from './overview/OverviewPage'
import { DataManagementPage, SystemInfoPage, SystemUsersPage } from './system-management/pages'

export type { PageKey } from '../navigation'

export { OverviewPage } from './overview/OverviewPage'
export { DemoPage } from './DemoPage'
export { LoginPage } from './LoginPage'
export { InitializationPage } from './InitializationPage'
export { UserManagementPage } from './system-management/UserManagementPage'

export const pageComponents: Record<PageKey, ComponentType> = {
  overview: OverviewPage,
  clusterManagement: ClusterManagementPage,
  hostManagement: HostManagementPage,
  monManagement: MonManagementPage,
  mgrManagement: MgrManagementPage,
  osdManagement: OsdManagementPage,
  mdsManagement: MdsManagementPage,
  blockPools: BlockPoolsPage,
  rbdImages: RbdImagesPage,
  imageMirroring: ImageMirroringPage,
  iscsi: IscsiPage,
  nvmeTcp: NvmeTcpPage,
  filePools: FilePoolsPage,
  cephfs: CephfsPage,
  nfs: NfsPage,
  smb: SmbPage,
  rgwOverview: RgwOverviewPage,
  rgwUsers: RgwUsersPage,
  bucketManagement: BucketManagementPage,
  gatewayManagement: GatewayManagementPage,
  multisite: MultisitePage,
  objectStorageConfig: ObjectStorageConfigPage,
  monitorOverview: MonitorOverviewPage,
  performanceMetrics: PerformanceMetricsPage,
  runtimeLogs: RuntimeLogsPage,
  alertList: AlertListPage,
  alertRules: AlertRulesPage,
  alertSilences: AlertSilencesPage,
  systemInfo: SystemInfoPage,
  systemUsers: SystemUsersPage,
  dataManagement: DataManagementPage
}
