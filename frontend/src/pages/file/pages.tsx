import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'

export function FilePoolsPage() {
  return <ResourceListPage definition={definitions.filePools} />
}

export function CephfsPage() {
  return <ResourceListPage definition={definitions.cephfs} />
}

export function CephfsClientsPage() {
  return <ResourceListPage definition={definitions.cephfsClients} />
}

export function SubvolumeGroupsPage() {
  return <ResourceListPage definition={definitions.subvolumeGroups} />
}

export function SubvolumesPage() {
  return <ResourceListPage definition={definitions.subvolumes} />
}

export function CephfsSnapshotsPage() {
  return <ResourceListPage definition={definitions.cephfsSnapshots} />
}

export function SnapshotSchedulesPage() {
  return <ResourceListPage definition={definitions.snapshotSchedules} />
}

export function CephfsAuthorizationsPage() {
  return <ResourceListPage definition={definitions.cephfsAuthorizations} />
}

export function CephfsEntriesPage() {
  return <ResourceListPage definition={definitions.cephfsEntries} />
}

export function NfsClustersPage() {
  return <ResourceListPage definition={definitions.nfsClusters} />
}

export function NfsPage() {
  return <ResourceListPage definition={definitions.nfs} />
}

export function SmbClustersPage() {
  return <ResourceListPage definition={definitions.smbClusters} />
}

export function SmbPage() {
  return <ResourceListPage definition={definitions.smb} />
}

const definitions: Record<
  | 'filePools'
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
  | 'smb',
  ResourceListPageDefinition
> = {
  filePools: {
    title: '文件存储池',
    path: '/pools',
    body: { application: 'cephfs' },
    columns: [
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'type', title: '类型' },
      { key: 'size', title: '副本/大小' },
      { key: 'pg_num', title: 'PG' },
      { key: 'application_metadata', title: '应用' }
    ]
  },
  cephfs: {
    title: 'CephFS 文件系统',
    path: '/filesystems',
    requiredCapabilities: ['cephfs_volume'],
    createAction: {
      title: '新建 CephFS 文件系统',
      buttonLabel: '新建文件系统',
      path: '/filesystem',
      method: 'POST',
      successMessage: 'CephFS 文件系统创建执行成功',
      fields: [
        { name: 'name', label: '文件系统名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新 CephFS 文件系统',
      path: '/filesystem',
      method: 'PATCH',
      successMessage: 'CephFS 文件系统更新执行成功',
      fields: [
        { name: 'max_mds', label: '最大 MDS 数', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        fs: resourceName(row),
        max_mds: Number(values.max_mds)
      })
    },
    deleteAction: {
      title: '删除 CephFS 文件系统',
      path: '/filesystem',
      action: 'filesystem.delete',
      resourceKind: 'filesystem',
      successMessage: 'CephFS 文件系统删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: resourceName(row) }),
      resourceKey: (row) => `filesystem/${resourceName(row)}`
    },
    columns: [
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'metadata_pool', title: '元数据池' },
      { key: 'data_pools', title: '数据池' },
      { key: 'mdsmap', title: 'MDS' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsClients: {
    title: 'CephFS 客户端',
    path: '/filesystem/clients',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'client_id', 'id'],
    deleteAction: {
      title: '驱逐 CephFS 客户端',
      path: '/filesystem/client',
      action: 'cephfs_client.evict',
      resourceKind: 'cephfs_client',
      successMessage: 'CephFS 客户端驱逐执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), client_id: clientId(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/client/${clientId(row)}`
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'filesystem', title: '文件系统名称' },
      { key: 'client_id', title: '客户端 ID' },
      { key: 'hostname', title: '主机' },
      { key: 'root', title: '根路径' },
      { key: 'state', title: '状态' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  subvolumeGroups: {
    title: '子卷组',
    path: '/filesystem/subvolume/groups',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建子卷组',
      buttonLabel: '新建子卷组',
      path: '/filesystem/subvolume/group',
      method: 'POST',
      successMessage: '子卷组创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', required: true },
        { name: 'name', label: '子卷组名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, fs: String(values.fs ?? ''), name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新子卷组',
      path: '/filesystem/subvolume/group',
      method: 'PATCH',
      successMessage: '子卷组更新执行成功',
      fields: [
        { name: 'size', label: '大小（字节）', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, fs: fsName(row), group: groupName(row), size: Number(values.size) })
    },
    deleteAction: {
      title: '删除子卷组',
      path: '/filesystem/subvolume/group',
      action: 'subvolume_group.delete',
      resourceKind: 'subvolume_group',
      successMessage: '子卷组删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), group: groupName(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/subvolume-group/${groupName(row)}`
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'filesystem', title: '文件系统名称' },
      { key: 'name', title: '名称' },
      { key: 'size', title: '大小' },
      { key: 'bytes_used', title: '已用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  subvolumes: {
    title: '子卷',
    path: '/filesystem/subvolumes',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建子卷',
      buttonLabel: '新建子卷',
      path: '/filesystem/subvolume',
      method: 'POST',
      successMessage: '子卷创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', required: true },
        { name: 'name', label: '子卷名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, fs: String(values.fs ?? ''), name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新子卷',
      path: '/filesystem/subvolume',
      method: 'PATCH',
      successMessage: '子卷更新执行成功',
      fields: [
        { name: 'size', label: '大小（字节）', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, fs: fsName(row), subvolume: subvolumeName(row), size: Number(values.size) })
    },
    deleteAction: {
      title: '删除子卷',
      path: '/filesystem/subvolume',
      action: 'subvolume.delete',
      resourceKind: 'subvolume',
      successMessage: '子卷删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), subvolume: subvolumeName(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/subvolume/${subvolumeName(row)}`
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'group', title: '子卷组' },
      { key: 'name', title: '名称' },
      { key: 'path', title: '路径' },
      { key: 'size', title: '大小' },
      { key: 'bytes_used', title: '已用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsSnapshots: {
    title: 'CephFS 快照',
    path: '/filesystem/subvolume/snapshots',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建 CephFS 快照',
      buttonLabel: '新建快照',
      path: '/filesystem/subvolume/snapshot',
      method: 'POST',
      successMessage: 'CephFS 快照创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', required: true },
        { name: 'subvolume', label: '子卷', required: true },
        { name: 'name', label: '快照名称', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        subvolume: String(values.subvolume ?? ''),
        name: String(values.name ?? '')
      })
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'subvolume', title: '子卷' },
      { key: 'name', title: '快照' },
      { key: 'created_at', title: '创建时间' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  snapshotSchedules: {
    title: '快照计划',
    path: '/filesystem/snapshot/schedules',
    requiredCapabilities: ['cephfs_volume'],
    createAction: {
      title: '新建快照计划',
      buttonLabel: '新建快照计划',
      path: '/filesystem/snapshot/schedule',
      method: 'POST',
      successMessage: '快照计划创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', required: true },
        { name: 'path', label: '路径', required: true, placeholder: '/' },
        { name: 'schedule', label: '计划', required: true, placeholder: '1h' }
      ],
      initialValues: { path: '/' },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        path: String(values.path ?? ''),
        schedule: String(values.schedule ?? '')
      })
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'path', title: '路径' },
      { key: 'schedule', title: '计划' },
      { key: 'retention', title: '保留' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsAuthorizations: {
    title: 'CephFS 访问授权',
    path: '/filesystem/authorizations',
    requiredCapabilities: ['cephfs_volume'],
    createAction: {
      title: '新建 CephFS 访问授权',
      buttonLabel: '新建授权',
      path: '/filesystem/authorization',
      method: 'POST',
      successMessage: 'CephFS 授权创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', required: true },
        { name: 'client', label: '客户端', required: true, placeholder: 'client.app' },
        { name: 'path', label: '路径', placeholder: '/' },
        {
          name: 'access',
          label: '访问权限',
          type: 'select',
          required: true,
          options: [
            { label: '只读', value: 'r' },
            { label: '读写', value: 'rw' }
          ]
        }
      ],
      initialValues: { path: '/', access: 'rw' },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        client: String(values.client ?? ''),
        ...(values.path ? { path: String(values.path) } : {}),
        access: String(values.access ?? 'rw')
      })
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'client', title: '客户端' },
      { key: 'path', title: '路径' },
      { key: 'access', title: '权限' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsEntries: {
    title: 'CephFS 目录配额',
    path: '/filesystem/entries',
    requiredCapabilities: ['cephfs_data_access'],
    updateAction: {
      title: '更新目录配额',
      path: '/filesystem/entry/quota',
      method: 'PATCH',
      successMessage: '目录配额更新执行成功',
      fields: [
        { name: 'max_bytes', label: '最大容量（字节）', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        fs: fsName(row),
        path: String(row?.path ?? '/'),
        max_bytes: Number(values.max_bytes)
      })
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'path', title: '路径' },
      { key: 'quota', title: '配额' },
      { key: 'bytes_used', title: '已用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  nfsClusters: {
    title: 'NFS 集群',
    path: '/nfs/clusters',
    requiredCapabilities: ['nfs'],
    createAction: {
      title: '新建 NFS 集群',
      buttonLabel: '新建集群',
      path: '/nfs/cluster',
      method: 'POST',
      successMessage: 'NFS 集群创建执行成功',
      fields: [
        { name: 'name', label: '集群名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    deleteAction: {
      title: '删除 NFS 集群',
      path: '/nfs/cluster',
      action: 'nfs_cluster.delete',
      resourceKind: 'nfs_cluster',
      successMessage: 'NFS 集群删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, name: resourceName(row) }),
      resourceKey: (row) => `nfs/cluster/${resourceName(row)}`
    },
    columns: [
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'placement', title: '放置策略' },
      { key: 'virtual_ip', title: 'VIP' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  nfs: {
    title: 'NFS 导出',
    path: '/nfs/exports',
    requiredCapabilities: ['nfs'],
    rowKeyCandidates: ['natural_key', 'export_id', 'pseudo', 'name'],
    createAction: {
      title: '新建 NFS 导出',
      buttonLabel: '新建导出',
      path: '/nfs/export',
      method: 'POST',
      successMessage: 'NFS 导出创建执行成功',
      fields: [
        { name: 'cluster', label: 'NFS 集群', required: true },
        { name: 'pseudo', label: '伪路径', required: true, placeholder: '/export' },
        { name: 'path', label: 'CephFS 路径', required: true, placeholder: '/data' },
        { name: 'filesystem', label: '文件系统', required: true },
        { name: 'read_only', label: '只读', type: 'boolean' }
      ],
      initialValues: { pseudo: '/export', path: '/', read_only: false },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        cluster: String(values.cluster ?? ''),
        pseudo: String(values.pseudo ?? ''),
        path: String(values.path ?? ''),
        filesystem: String(values.filesystem ?? ''),
        read_only: Boolean(values.read_only)
      })
    },
    updateAction: {
      title: '更新 NFS 导出',
      path: '/nfs/export',
      method: 'PATCH',
      successMessage: 'NFS 导出更新执行成功',
      fields: [
        { name: 'cluster', label: 'NFS 集群', required: true },
        { name: 'pseudo', label: '伪路径', required: true },
        { name: 'path', label: 'CephFS 路径', required: true },
        { name: 'filesystem', label: '文件系统', required: true },
        { name: 'read_only', label: '只读', type: 'boolean' }
      ],
      initialValues: (row) => ({
        cluster: text(row?.cluster),
        pseudo: text(row?.pseudo),
        path: text(row?.path),
        filesystem: text(row?.filesystem),
        read_only: row?.read_only === true
      }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        export_id: exportId(row),
        cluster: String(values.cluster ?? ''),
        pseudo: String(values.pseudo ?? ''),
        path: String(values.path ?? ''),
        filesystem: String(values.filesystem ?? ''),
        read_only: Boolean(values.read_only)
      })
    },
    deleteAction: {
      title: '删除 NFS 导出',
      path: '/nfs/export',
      action: 'nfs_export.delete',
      resourceKind: 'nfs_export',
      successMessage: 'NFS 导出删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, export_id: exportId(row) }),
      resourceKey: (row) => `nfs/export/${exportId(row)}`
    },
    columns: [
      { key: 'export_id', title: '导出 ID' },
      { key: 'cluster', title: '集群' },
      { key: 'pseudo', title: '伪路径' },
      { key: 'path', title: '路径' },
      { key: 'filesystem', title: '文件系统' },
      { key: 'read_only', title: '只读' },
      { key: 'status', title: '状态' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  smbClusters: {
    title: 'SMB 集群',
    path: '/smb/clusters',
    requiredCapabilities: ['smb'],
    createAction: {
      title: '新建 SMB 集群',
      buttonLabel: '新建集群',
      path: '/smb/cluster',
      method: 'POST',
      successMessage: 'SMB 集群创建执行成功',
      fields: [
        { name: 'name', label: '集群名称', required: true },
        {
          name: 'auth_mode',
          label: '认证模式',
          type: 'select',
          options: [
            { label: '本地用户', value: 'user' },
            { label: 'Active Directory', value: 'active-directory' }
          ]
        }
      ],
      initialValues: { auth_mode: 'user' },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        name: String(values.name ?? ''),
        ...(values.auth_mode ? { auth_mode: String(values.auth_mode) } : {})
      })
    },
    updateAction: {
      title: '更新 SMB 集群',
      path: '/smb/cluster',
      method: 'PATCH',
      successMessage: 'SMB 集群更新执行成功',
      fields: [
        {
          name: 'auth_mode',
          label: '认证模式',
          type: 'select',
          required: true,
          options: [
            { label: '本地用户', value: 'user' },
            { label: 'Active Directory', value: 'active-directory' }
          ]
        }
      ],
      initialValues: (row) => ({ auth_mode: text(row?.auth_mode) || 'user' }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        name: resourceName(row),
        auth_mode: String(values.auth_mode ?? 'user')
      })
    },
    deleteAction: {
      title: '删除 SMB 集群',
      path: '/smb/cluster',
      action: 'smb_cluster.delete',
      resourceKind: 'smb_cluster',
      successMessage: 'SMB 集群删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, name: resourceName(row) }),
      resourceKey: (row) => `smb/cluster/${resourceName(row)}`
    },
    columns: [
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'placement', title: '放置策略' },
      { key: 'auth_mode', title: '认证' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  smb: {
    title: 'SMB 共享',
    path: '/smb/shares',
    requiredCapabilities: ['smb'],
    rowKeyCandidates: ['natural_key', 'share_id', 'name'],
    createAction: {
      title: '新建 SMB 共享',
      buttonLabel: '新建共享',
      path: '/smb/share',
      method: 'POST',
      successMessage: 'SMB 共享创建执行成功',
      fields: [
        { name: 'cluster', label: 'SMB 集群', required: true },
        { name: 'name', label: '共享名称', required: true },
        { name: 'filesystem', label: '文件系统', required: true },
        { name: 'path', label: 'CephFS 路径', required: true, placeholder: '/data' }
      ],
      initialValues: { path: '/' },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        cluster: String(values.cluster ?? ''),
        name: String(values.name ?? ''),
        filesystem: String(values.filesystem ?? ''),
        path: String(values.path ?? '')
      })
    },
    updateAction: {
      title: '更新 SMB 共享',
      path: '/smb/share',
      method: 'PATCH',
      successMessage: 'SMB 共享更新执行成功',
      fields: [
        { name: 'cluster', label: 'SMB 集群', required: true },
        { name: 'filesystem', label: '文件系统', required: true },
        { name: 'path', label: 'CephFS 路径' }
      ],
      initialValues: (row) => ({
        cluster: text(row?.cluster),
        filesystem: text(row?.filesystem),
        path: text(row?.path)
      }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        share_id: shareId(row),
        cluster: String(values.cluster ?? ''),
        filesystem: String(values.filesystem ?? ''),
        ...(values.path ? { path: String(values.path) } : {})
      })
    },
    deleteAction: {
      title: '删除 SMB 共享',
      path: '/smb/share',
      action: 'smb_share.delete',
      resourceKind: 'smb_share',
      successMessage: 'SMB 共享删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, share_id: shareId(row) }),
      resourceKey: (row) => `smb/share/${shareId(row)}`
    },
    columns: [
      { key: 'share_id', title: '共享 ID' },
      { key: 'cluster', title: '集群' },
      { key: 'name', title: '名称' },
      { key: 'filesystem', title: '文件系统' },
      { key: 'path', title: '路径' },
      { key: 'status', title: '状态' },
      { key: 'auth_mode', title: '认证' },
      { key: 'resource_version', title: '版本' }
    ]
  }
}

function resourceName(row?: Record<string, unknown>) {
  return String(row?.name ?? row?.natural_key ?? '').trim()
}

function fsName(row?: Record<string, unknown>) {
  return String(row?.fs ?? row?.filesystem ?? row?.filesystem_name ?? row?.name ?? '').trim()
}

function groupName(row?: Record<string, unknown>) {
  return String(row?.group ?? row?.group_name ?? row?.name ?? '').trim()
}

function subvolumeName(row?: Record<string, unknown>) {
  return String(row?.subvolume ?? row?.subvolume_name ?? row?.name ?? '').trim()
}

function clientId(row?: Record<string, unknown>) {
  return String(row?.client_id ?? row?.id ?? row?.natural_key ?? '').trim()
}

function exportId(row?: Record<string, unknown>) {
  return String(row?.export_id ?? row?.id ?? row?.natural_key ?? '').trim()
}

function shareId(row?: Record<string, unknown>) {
  return String(row?.share_id ?? row?.id ?? row?.natural_key ?? row?.name ?? '').trim()
}

function text(value: unknown) {
  return value === null || value === undefined ? '' : String(value)
}
