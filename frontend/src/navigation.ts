export type PageKey =
  | 'overview'
  | 'clusterManagement'
  | 'hostManagement'
  | 'monManagement'
  | 'mgrManagement'
  | 'osdManagement'
  | 'mdsManagement'
  | 'blockPools'
  | 'rbdImages'
  | 'imageMirroring'
  | 'iscsi'
  | 'nvmeTcp'
  | 'filePools'
  | 'cephfs'
  | 'nfs'
  | 'smb'
  | 'rgwOverview'
  | 'rgwUsers'
  | 'bucketManagement'
  | 'gatewayManagement'
  | 'multisite'
  | 'objectStorageConfig'
  | 'monitorOverview'
  | 'performanceMetrics'
  | 'runtimeLogs'
  | 'alertList'
  | 'alertRules'
  | 'alertSilences'
  | 'systemInfo'
  | 'systemUsers'
  | 'dataManagement'

export type NavIcon =
  | 'overview'
  | 'cluster'
  | 'host'
  | 'mon'
  | 'mgr'
  | 'osd'
  | 'mds'
  | 'block'
  | 'pool'
  | 'rbd'
  | 'sync'
  | 'iscsi'
  | 'nvme'
  | 'file'
  | 'cephfs'
  | 'nfs'
  | 'smb'
  | 'object'
  | 'user'
  | 'bucket'
  | 'gateway'
  | 'site'
  | 'config'
  | 'monitor'
  | 'metrics'
  | 'logs'
  | 'alert'
  | 'rule'
  | 'silence'
  | 'system'
  | 'data'

export type NavChildDefinition = {
  key: PageKey
  label: string
  path: string
  icon: NavIcon
  permission: 'cluster' | 'storage' | 'system'
}

export type NavSectionDefinition = {
  key: string
  label: string
  path: string
  icon: NavIcon
  children: NavChildDefinition[]
}

export const NAV_SECTIONS: NavSectionDefinition[] = [
  {
    key: 'overview-section',
    label: '总览',
    path: '/overview',
    icon: 'overview',
    children: [{ key: 'overview', label: '总览', path: '/overview', icon: 'overview', permission: 'cluster' }]
  },
  {
    key: 'cluster-management-section',
    label: '集群管理',
    path: '/cluster-management',
    icon: 'cluster',
    children: [
      { key: 'clusterManagement', label: '集群管理', path: '/cluster-management/clusters', icon: 'cluster', permission: 'cluster' },
      { key: 'hostManagement', label: '主机管理', path: '/cluster-management/hosts', icon: 'host', permission: 'cluster' },
      { key: 'monManagement', label: 'MON管理', path: '/cluster-management/mons', icon: 'mon', permission: 'cluster' },
      { key: 'mgrManagement', label: 'MGR管理', path: '/cluster-management/mgrs', icon: 'mgr', permission: 'cluster' },
      { key: 'osdManagement', label: 'OSD管理', path: '/cluster-management/osds', icon: 'osd', permission: 'cluster' },
      { key: 'mdsManagement', label: 'MDS管理', path: '/cluster-management/mdss', icon: 'mds', permission: 'cluster' }
    ]
  },
  {
    key: 'storage-block-section',
    label: '块存储',
    path: '/storage-block',
    icon: 'block',
    children: [
      { key: 'blockPools', label: '存储池', path: '/storage-block/pools', icon: 'pool', permission: 'storage' },
      { key: 'rbdImages', label: 'RBD镜像', path: '/storage-block/rbd-images', icon: 'rbd', permission: 'storage' },
      { key: 'imageMirroring', label: '镜像同步', path: '/storage-block/mirroring', icon: 'sync', permission: 'storage' },
      { key: 'iscsi', label: 'iSCSI', path: '/storage-block/iscsi', icon: 'iscsi', permission: 'storage' },
      { key: 'nvmeTcp', label: 'NVMe/TCP', path: '/storage-block/nvme-tcp', icon: 'nvme', permission: 'storage' }
    ]
  },
  {
    key: 'storage-file-section',
    label: '文件存储',
    path: '/storage-file',
    icon: 'file',
    children: [
      { key: 'filePools', label: '存储池', path: '/storage-file/pools', icon: 'pool', permission: 'storage' },
      { key: 'cephfs', label: 'CephFS', path: '/storage-file/cephfs', icon: 'cephfs', permission: 'storage' },
      { key: 'nfs', label: 'NFS', path: '/storage-file/nfs', icon: 'nfs', permission: 'storage' },
      { key: 'smb', label: 'SMB', path: '/storage-file/smb', icon: 'smb', permission: 'storage' }
    ]
  },
  {
    key: 'storage-object-section',
    label: '对象存储',
    path: '/storage-object',
    icon: 'object',
    children: [
      { key: 'rgwOverview', label: 'RGW总览', path: '/storage-object/rgw-overview', icon: 'object', permission: 'storage' },
      { key: 'rgwUsers', label: '用户管理', path: '/storage-object/users', icon: 'user', permission: 'storage' },
      { key: 'bucketManagement', label: 'Bucket管理', path: '/storage-object/buckets', icon: 'bucket', permission: 'storage' },
      { key: 'gatewayManagement', label: '网关管理', path: '/storage-object/gateways', icon: 'gateway', permission: 'storage' },
      { key: 'multisite', label: '多站点', path: '/storage-object/multisite', icon: 'site', permission: 'storage' },
      { key: 'objectStorageConfig', label: '对象存储配置', path: '/storage-object/configuration', icon: 'config', permission: 'storage' }
    ]
  },
  {
    key: 'monitoring-alerting-section',
    label: '监控报警',
    path: '/monitoring-alerting',
    icon: 'monitor',
    children: [
      { key: 'monitorOverview', label: '监控总览', path: '/monitoring-alerting/overview', icon: 'monitor', permission: 'system' },
      { key: 'performanceMetrics', label: '性能指标', path: '/monitoring-alerting/metrics', icon: 'metrics', permission: 'system' },
      { key: 'runtimeLogs', label: '运行日志', path: '/monitoring-alerting/logs', icon: 'logs', permission: 'system' },
      { key: 'alertList', label: '告警列表', path: '/monitoring-alerting/alerts', icon: 'alert', permission: 'system' },
      { key: 'alertRules', label: '告警规则', path: '/monitoring-alerting/rules', icon: 'rule', permission: 'system' },
      { key: 'alertSilences', label: '告警静默', path: '/monitoring-alerting/silences', icon: 'silence', permission: 'system' }
    ]
  },
  {
    key: 'system-management-section',
    label: '系统管理',
    path: '/system-management',
    icon: 'system',
    children: [
      { key: 'systemInfo', label: '系统信息', path: '/system-management/info', icon: 'system', permission: 'system' },
      { key: 'systemUsers', label: '用户管理', path: '/system-management/users', icon: 'user', permission: 'system' },
      { key: 'dataManagement', label: '数据管理', path: '/system-management/data', icon: 'data', permission: 'system' }
    ]
  }
]

export const NAV_PAGES = NAV_SECTIONS.flatMap((section) => section.children)

export const pagePaths = NAV_PAGES.reduce(
  (paths, page) => {
    paths[page.key] = page.path
    return paths
  },
  {} as Record<PageKey, string>
)

export function findNavPage(pageKey: PageKey) {
  return NAV_PAGES.find((page) => page.key === pageKey)
}

export function findNavSection(pageKey: PageKey) {
  return NAV_SECTIONS.find((section) => section.children.some((page) => page.key === pageKey))
}
