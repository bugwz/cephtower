export type PageKey =
  | 'overview'
  | 'clusterManagement'
  | 'poolManagement'
  | 'hostManagement'
  | 'monManagement'
  | 'mgrManagement'
  | 'osdManagement'
  | 'deviceManagement'
  | 'mdsManagement'
  | 'rbdImages'
  | 'rbdSnapshots'
  | 'rbdNamespaces'
  | 'rbdTrash'
  | 'rbdGroups'
  | 'imageMirroring'
  | 'iscsi'
  | 'nvmeGateway'
  | 'nvmeTcp'
  | 'nvmeNamespaces'
  | 'nvmeListeners'
  | 'nvmeHosts'
  | 'nvmeConnections'
  | 'cephfs'
  | 'cephfsClients'
  | 'subvolumeGroups'
  | 'subvolumes'
  | 'cephfsSnapshots'
  | 'snapshotSchedules'
  | 'cephfsAuthorizations'
  | 'cephfsEntries'
  | 'nfsClusters'
  | 'nfs'
  | 'smbClusters'
  | 'smb'
  | 'rgwOverview'
  | 'rgwUsers'
  | 'rgwAccounts'
  | 'rgwRoles'
  | 'bucketManagement'
  | 'bucketPolicy'
  | 'gatewayManagement'
  | 'multisite'
  | 'rgwZonegroups'
  | 'rgwZones'
  | 'rgwPeriod'
  | 'objectStorageConfig'
  | 'monitorOverview'
  | 'performanceMetrics'
  | 'runtimeLogs'
  | 'alertList'
  | 'alertRules'
  | 'alertSilences'
  | 'systemInfo'
  | 'systemUsers'
  | 'systemRoles'
  | 'systemRoleBindings'
  | 'dataManagement'
  | 'auditEvents'

export type NavIcon =
  | 'overview'
  | 'cluster'
  | 'host'
  | 'mon'
  | 'mgr'
  | 'osd'
  | 'device'
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
  | 'audit'

export type NavChildDefinition = {
  key: PageKey
  label: string
  path: string
  icon: NavIcon
  permission: 'cluster' | 'storage' | 'system'
}

export type NavChildGroupDefinition = {
  key: string
  label: string
  path: string
  icon: NavIcon
  children: NavChildDefinition[]
}

export type NavSectionDefinition = {
  key: string
  label: string
  path: string
  icon: NavIcon
  children: Array<NavChildDefinition | NavChildGroupDefinition>
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
      { key: 'clusterManagement', label: '集群列表', path: '/cluster/cluster', icon: 'cluster', permission: 'cluster' },
      { key: 'poolManagement', label: '存储池管理', path: '/cluster/pool', icon: 'pool', permission: 'cluster' },
      {
        key: 'host-management-section',
        label: '主机管理',
        path: '/cluster/host',
        icon: 'host',
        children: [
          { key: 'hostManagement', label: '主机列表', path: '/cluster/host', icon: 'host', permission: 'cluster' },
          { key: 'deviceManagement', label: '设备列表', path: '/cluster/host/device', icon: 'device', permission: 'cluster' }
        ]
      },
      { key: 'monManagement', label: 'MON 管理', path: '/cluster/mon', icon: 'mon', permission: 'cluster' },
      { key: 'mgrManagement', label: 'MGR 管理', path: '/cluster/mgr', icon: 'mgr', permission: 'cluster' },
      { key: 'osdManagement', label: 'OSD 管理', path: '/cluster/osd', icon: 'osd', permission: 'cluster' },
      { key: 'mdsManagement', label: 'MDS 管理', path: '/cluster/mds', icon: 'mds', permission: 'cluster' }
    ]
  },
  {
    key: 'block-section',
    label: '块存储',
    path: '/block',
    icon: 'block',
    children: [
      {
        key: 'rbd-image-section',
        label: 'RBD 镜像',
        path: '/block/rbd-image',
        icon: 'rbd',
        children: [
          { key: 'rbdImages', label: '镜像', path: '/block/rbd-image', icon: 'rbd', permission: 'storage' },
          { key: 'rbdSnapshots', label: '快照', path: '/block/rbd-snapshot', icon: 'logs', permission: 'storage' },
          { key: 'rbdNamespaces', label: '命名空间', path: '/block/rbd-namespace', icon: 'cephfs', permission: 'storage' },
          { key: 'rbdTrash', label: '回收站', path: '/block/rbd-trash', icon: 'silence', permission: 'storage' },
          { key: 'rbdGroups', label: '镜像组', path: '/block/rbd-group', icon: 'site', permission: 'storage' }
        ]
      },
      { key: 'imageMirroring', label: '镜像同步', path: '/block/mirroring', icon: 'sync', permission: 'storage' },
      { key: 'iscsi', label: 'iSCSI', path: '/block/iscsi', icon: 'iscsi', permission: 'storage' },
      {
        key: 'nvme-tcp-section',
        label: 'NVMe/TCP',
        path: '/block/nvme-tcp',
        icon: 'nvme',
        children: [
          { key: 'nvmeGateway', label: 'Gateway', path: '/block/nvme-tcp/gateway', icon: 'gateway', permission: 'storage' },
          { key: 'nvmeTcp', label: 'Subsystems', path: '/block/nvme-tcp', icon: 'nvme', permission: 'storage' },
          { key: 'nvmeNamespaces', label: 'Namespaces', path: '/block/nvme-tcp/namespaces', icon: 'cephfs', permission: 'storage' },
          { key: 'nvmeListeners', label: 'Listeners', path: '/block/nvme-tcp/listeners', icon: 'logs', permission: 'storage' },
          { key: 'nvmeHosts', label: 'Hosts', path: '/block/nvme-tcp/hosts', icon: 'host', permission: 'storage' },
          { key: 'nvmeConnections', label: 'Connections', path: '/block/nvme-tcp/connections', icon: 'iscsi', permission: 'storage' }
        ]
      }
    ]
  },
  {
    key: 'file-section',
    label: '文件存储',
    path: '/file',
    icon: 'file',
    children: [
      {
        key: 'cephfs-section',
        label: 'CephFS',
        path: '/file/cephfs',
        icon: 'cephfs',
        children: [
          { key: 'cephfs', label: '文件系统列表', path: '/file/cephfs', icon: 'cephfs', permission: 'storage' },
          { key: 'cephfsClients', label: '客户端', path: '/file/cephfs/clients', icon: 'user', permission: 'storage' },
          { key: 'subvolumeGroups', label: '子卷组', path: '/file/cephfs/subvolume-groups', icon: 'site', permission: 'storage' },
          { key: 'subvolumes', label: '子卷', path: '/file/cephfs/subvolumes', icon: 'file', permission: 'storage' },
          { key: 'cephfsSnapshots', label: '快照', path: '/file/cephfs/snapshots', icon: 'logs', permission: 'storage' },
          { key: 'snapshotSchedules', label: '快照计划', path: '/file/cephfs/schedules', icon: 'monitor', permission: 'storage' },
          { key: 'cephfsAuthorizations', label: '访问授权', path: '/file/cephfs/authorizations', icon: 'audit', permission: 'storage' },
          { key: 'cephfsEntries', label: '目录配额', path: '/file/cephfs/entries', icon: 'config', permission: 'storage' }
        ]
      },
      {
        key: 'nfs-section',
        label: 'NFS',
        path: '/file/nfs',
        icon: 'nfs',
        children: [
          { key: 'nfsClusters', label: '集群', path: '/file/nfs/clusters', icon: 'cluster', permission: 'storage' },
          { key: 'nfs', label: '导出', path: '/file/nfs', icon: 'nfs', permission: 'storage' }
        ]
      },
      {
        key: 'smb-section',
        label: 'SMB',
        path: '/file/smb',
        icon: 'smb',
        children: [
          { key: 'smbClusters', label: '集群', path: '/file/smb/clusters', icon: 'cluster', permission: 'storage' },
          { key: 'smb', label: '共享', path: '/file/smb', icon: 'smb', permission: 'storage' }
        ]
      }
    ]
  },
  {
    key: 'object-section',
    label: '对象存储',
    path: '/object',
    icon: 'object',
    children: [
      { key: 'rgwOverview', label: 'RGW 总览', path: '/object/rgw-overview', icon: 'object', permission: 'storage' },
      {
        key: 'rgw-user-section',
        label: '用户管理',
        path: '/object/user',
        icon: 'user',
        children: [
          { key: 'rgwUsers', label: '用户', path: '/object/user', icon: 'user', permission: 'storage' },
          { key: 'rgwAccounts', label: 'Accounts', path: '/object/user/accounts', icon: 'audit', permission: 'storage' },
          { key: 'rgwRoles', label: 'Roles', path: '/object/user/roles', icon: 'system', permission: 'storage' }
        ]
      },
      {
        key: 'bucket-management-section',
        label: 'Bucket 管理',
        path: '/object/bucket',
        icon: 'bucket',
        children: [
          { key: 'bucketManagement', label: 'Buckets', path: '/object/bucket', icon: 'bucket', permission: 'storage' },
          { key: 'bucketPolicy', label: 'Policy', path: '/object/bucket/policy', icon: 'config', permission: 'storage' }
        ]
      },
      { key: 'gatewayManagement', label: '网关管理', path: '/object/gateway', icon: 'gateway', permission: 'storage' },
      {
        key: 'multisite-section',
        label: '多站点',
        path: '/object/multisite',
        icon: 'site',
        children: [
          { key: 'multisite', label: 'Realms', path: '/object/multisite', icon: 'site', permission: 'storage' },
          { key: 'rgwZonegroups', label: 'ZoneGroups', path: '/object/multisite/zonegroups', icon: 'cluster', permission: 'storage' },
          { key: 'rgwZones', label: 'Zones', path: '/object/multisite/zones', icon: 'gateway', permission: 'storage' },
          { key: 'rgwPeriod', label: 'Period', path: '/object/multisite/period', icon: 'sync', permission: 'storage' }
        ]
      },
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
      {
        key: 'system-user-section',
        label: '用户管理',
        path: '/system/user',
        icon: 'user',
        children: [
          { key: 'systemUsers', label: '用户', path: '/system/user', icon: 'user', permission: 'system' },
          { key: 'systemRoles', label: '角色', path: '/system/user/role', icon: 'system', permission: 'system' },
          { key: 'systemRoleBindings', label: '集群授权', path: '/system/user/binding', icon: 'audit', permission: 'system' }
        ]
      },
      { key: 'dataManagement', label: '配置管理', path: '/system/data', icon: 'config', permission: 'system' }
    ]
  },
  {
    key: 'audit-section',
    label: '审计',
    path: '/audit',
    icon: 'audit',
    children: [
      { key: 'auditEvents', label: '审计事件', path: '/audit/events', icon: 'audit', permission: 'system' }
    ]
  }
]

export const NAV_PAGES = NAV_SECTIONS.flatMap((section) => flattenNavChildren(section.children))

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
  return NAV_SECTIONS.find((section) => flattenNavChildren(section.children).some((page) => page.key === pageKey))
}

function flattenNavChildren(children: Array<NavChildDefinition | NavChildGroupDefinition>): NavChildDefinition[] {
  return children.flatMap((item) => ('children' in item ? item.children : [item]))
}
