import type { ComponentType } from 'react'
import type { PageKey } from '../navigation'
import { BlockPoolsPage, ImageMirroringPage, IscsiPage, NvmeTcpPage, RbdImagesPage } from './block/pages'
import {
  ClusterDetailPage,
  ClusterPage,
  HostPage,
  MdsManagementPage,
  MgrManagementPage,
  MonManagementPage,
  OsdManagementPage
} from './cluster/pages'
import { CephfsPage, FilePoolsPage, NfsPage, SmbPage } from './file/pages'
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
  GatewayManagementPage,
  MultisitePage,
  ObjectStorageConfigPage,
  RgwOverviewPage,
  RgwUsersPage
} from './object/pages'
import { OverviewPage } from './overview/OverviewPage'
import { DataManagementPage, SystemInfoPage, SystemUsersPage } from './system/pages'

export type { PageKey } from '../navigation'

export { OverviewPage } from './overview/OverviewPage'
export { DemoPage } from './DemoPage'
export { LoginPage } from './LoginPage'
export { InitializationPage } from './InitializationPage'
export { UserPage } from './system/UserPage'
export { ClusterDetailPage } from './cluster/ClusterDetailPage'

export const pageComponents: Record<PageKey, ComponentType> = {
  overview: OverviewPage,
  clusterManagement: ClusterPage,
  hostManagement: HostPage,
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
