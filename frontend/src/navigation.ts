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
    key: 'cluster-section',
    label: '集群管理',
    path: '/cluster',
    icon: 'cluster',
    children: [
      { key: 'clusterManagement', label: '集群管理', path: '/cluster/cluster', icon: 'cluster', permission: 'cluster' },
      { key: 'hostManagement', label: '主机管理', path: '/cluster/host', icon: 'host', permission: 'cluster' },
      { key: 'monManagement', label: 'MON管理', path: '/cluster/mon', icon: 'mon', permission: 'cluster' },
      { key: 'mgrManagement', label: 'MGR管理', path: '/cluster/mgr', icon: 'mgr', permission: 'cluster' },
      { key: 'osdManagement', label: 'OSD管理', path: '/cluster/osd', icon: 'osd', permission: 'cluster' },
      { key: 'mdsManagement', label: 'MDS管理', path: '/cluster/mds', icon: 'mds', permission: 'cluster' }
    ]
  },
  {
    key: 'block-section',
    label: '块存储',
    path: '/block',
    icon: 'block',
    children: [
      { key: 'blockPools', label: '存储池', path: '/block/pool', icon: 'pool', permission: 'storage' },
      { key: 'rbdImages', label: 'RBD镜像', path: '/block/rbd-image', icon: 'rbd', permission: 'storage' },
      { key: 'imageMirroring', label: '镜像同步', path: '/block/mirroring', icon: 'sync', permission: 'storage' },
      { key: 'iscsi', label: 'iSCSI', path: '/block/iscsi', icon: 'iscsi', permission: 'storage' },
      { key: 'nvmeTcp', label: 'NVMe/TCP', path: '/block/nvme-tcp', icon: 'nvme', permission: 'storage' }
    ]
  },
  {
    key: 'file-section',
    label: '文件存储',
    path: '/file',
    icon: 'file',
    children: [
      { key: 'filePools', label: '存储池', path: '/file/pool', icon: 'pool', permission: 'storage' },
      { key: 'cephfs', label: 'CephFS', path: '/file/cephfs', icon: 'cephfs', permission: 'storage' },
      { key: 'nfs', label: 'NFS', path: '/file/nfs', icon: 'nfs', permission: 'storage' },
      { key: 'smb', label: 'SMB', path: '/file/smb', icon: 'smb', permission: 'storage' }
    ]
  },
  {
    key: 'object-section',
    label: '对象存储',
    path: '/object',
    icon: 'object',
    children: [
      { key: 'rgwOverview', label: 'RGW总览', path: '/object/rgw-overview', icon: 'object', permission: 'storage' },
      { key: 'rgwUsers', label: '用户管理', path: '/object/user', icon: 'user', permission: 'storage' },
      { key: 'bucketManagement', label: 'Bucket管理', path: '/object/bucket', icon: 'bucket', permission: 'storage' },
      { key: 'gatewayManagement', label: '网关管理', path: '/object/gateway', icon: 'gateway', permission: 'storage' },
      { key: 'multisite', label: '多站点', path: '/object/multisite', icon: 'site', permission: 'storage' },
      { key: 'objectStorageConfig', label: '对象存储配置', path: '/object/configuration', icon: 'config', permission: 'storage' }
    ]
  },
  {
    key: 'monitoring-section',
    label: '监控报警',
    path: '/monitoring',
    icon: 'monitor',
    children: [
      { key: 'monitorOverview', label: '监控总览', path: '/monitoring/overview', icon: 'monitor', permission: 'system' },
      { key: 'performanceMetrics', label: '性能指标', path: '/monitoring/metric', icon: 'metrics', permission: 'system' },
      { key: 'runtimeLogs', label: '运行日志', path: '/monitoring/log', icon: 'logs', permission: 'system' },
      { key: 'alertList', label: '告警列表', path: '/monitoring/alert', icon: 'alert', permission: 'system' },
      { key: 'alertRules', label: '告警规则', path: '/monitoring/rule', icon: 'rule', permission: 'system' },
      { key: 'alertSilences', label: '告警静默', path: '/monitoring/silence', icon: 'silence', permission: 'system' }
    ]
  },
  {
    key: 'system-section',
    label: '系统管理',
    path: '/system',
    icon: 'system',
    children: [
      { key: 'systemInfo', label: '系统信息', path: '/system/info', icon: 'system', permission: 'system' },
      { key: 'systemUsers', label: '用户管理', path: '/system/user', icon: 'user', permission: 'system' },
      { key: 'dataManagement', label: '配置管理', path: '/system/data', icon: 'config', permission: 'system' }
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
