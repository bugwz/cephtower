import { Card, Tabs } from 'antd'
import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'

export function FilePoolsPage() {
  return <ResourceListPage definition={definitions.filePools} />
}

export function CephfsPage() {
  return (
    <Card className="page-surface-card tab-surface-card">
      <Tabs
        className="page-tabs"
        items={[
          { key: 'filesystems', label: '文件系统', children: <ResourceListPage definition={definitions.cephfs} embedded /> },
          { key: 'clients', label: 'Clients', children: <ResourceListPage definition={definitions.cephfsClients} embedded /> },
          { key: 'groups', label: 'Subvolume Groups', children: <ResourceListPage definition={definitions.subvolumeGroups} embedded /> },
          { key: 'subvolumes', label: 'Subvolumes', children: <ResourceListPage definition={definitions.subvolumes} embedded /> },
          { key: 'snapshots', label: 'Snapshots', children: <ResourceListPage definition={definitions.cephfsSnapshots} embedded /> },
          { key: 'schedules', label: 'Schedules', children: <ResourceListPage definition={definitions.snapshotSchedules} embedded /> },
          { key: 'auth', label: 'Authorizations', children: <ResourceListPage definition={definitions.cephfsAuthorizations} embedded /> },
          { key: 'entries', label: '目录配额', children: <ResourceListPage definition={definitions.cephfsEntries} embedded /> }
        ]}
      />
    </Card>
  )
}

export function NfsPage() {
  return (
    <Card className="page-surface-card tab-surface-card">
      <Tabs
        className="page-tabs"
        items={[
          { key: 'clusters', label: 'Clusters', children: <ResourceListPage definition={definitions.nfsClusters} embedded /> },
          { key: 'exports', label: 'Exports', children: <ResourceListPage definition={definitions.nfs} embedded /> }
        ]}
      />
    </Card>
  )
}

export function SmbPage() {
  return (
    <Card className="page-surface-card tab-surface-card">
      <Tabs
        className="page-tabs"
        items={[
          { key: 'clusters', label: 'Clusters', children: <ResourceListPage definition={definitions.smbClusters} embedded /> },
          { key: 'shares', label: 'Shares', children: <ResourceListPage definition={definitions.smb} embedded /> }
        ]}
      />
    </Card>
  )
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
      title: '新建 CephFS',
      buttonLabel: '新建文件系统',
      path: '/filesystem',
      method: 'POST',
      successMessage: 'CephFS 创建执行成功',
      fields: [
        { name: 'name', label: '文件系统名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新 CephFS',
      path: '/filesystem',
      method: 'PATCH',
      successMessage: 'CephFS 更新执行成功',
      fields: [
        { name: 'max_mds', label: 'Max MDS', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        fs: resourceName(row),
        max_mds: Number(values.max_mds)
      })
    },
    deleteAction: {
      title: '删除 CephFS',
      path: '/filesystem',
      action: 'filesystem.delete',
      resourceKind: 'filesystem',
      successMessage: 'CephFS 删除执行成功',
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
    title: 'CephFS Clients',
    path: '/filesystem/clients',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'client_id', 'id'],
    deleteAction: {
      title: '驱逐 CephFS Client',
      path: '/filesystem/client',
      action: 'cephfs_client.evict',
      resourceKind: 'cephfs_client',
      successMessage: 'CephFS Client 驱逐执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), client_id: clientId(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/client/${clientId(row)}`
    },
    columns: [
      { key: 'fs', title: 'FS' },
      { key: 'filesystem', title: 'Filesystem' },
      { key: 'client_id', title: 'Client ID' },
      { key: 'hostname', title: '主机' },
      { key: 'root', title: 'Root' },
      { key: 'state', title: '状态' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  subvolumeGroups: {
    title: 'Subvolume Groups',
    path: '/filesystem/subvolume/groups',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建 Subvolume Group',
      buttonLabel: '新建 Group',
      path: '/filesystem/subvolume/group',
      method: 'POST',
      successMessage: 'Subvolume Group 创建执行成功',
      fields: [
        { name: 'fs', label: 'Filesystem', required: true },
        { name: 'name', label: 'Group 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, fs: String(values.fs ?? ''), name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新 Subvolume Group',
      path: '/filesystem/subvolume/group',
      method: 'PATCH',
      successMessage: 'Subvolume Group 更新执行成功',
      fields: [
        { name: 'size', label: '大小 bytes', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, fs: fsName(row), group: groupName(row), size: Number(values.size) })
    },
    deleteAction: {
      title: '删除 Subvolume Group',
      path: '/filesystem/subvolume/group',
      action: 'subvolume_group.delete',
      resourceKind: 'subvolume_group',
      successMessage: 'Subvolume Group 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), group: groupName(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/subvolume-group/${groupName(row)}`
    },
    columns: [
      { key: 'fs', title: 'FS' },
      { key: 'filesystem', title: 'Filesystem' },
      { key: 'name', title: '名称' },
      { key: 'size', title: '大小' },
      { key: 'bytes_used', title: '已用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  subvolumes: {
    title: 'Subvolumes',
    path: '/filesystem/subvolumes',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建 Subvolume',
      buttonLabel: '新建 Subvolume',
      path: '/filesystem/subvolume',
      method: 'POST',
      successMessage: 'Subvolume 创建执行成功',
      fields: [
        { name: 'fs', label: 'Filesystem', required: true },
        { name: 'name', label: 'Subvolume 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, fs: String(values.fs ?? ''), name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新 Subvolume',
      path: '/filesystem/subvolume',
      method: 'PATCH',
      successMessage: 'Subvolume 更新执行成功',
      fields: [
        { name: 'size', label: '大小 bytes', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, fs: fsName(row), subvolume: subvolumeName(row), size: Number(values.size) })
    },
    deleteAction: {
      title: '删除 Subvolume',
      path: '/filesystem/subvolume',
      action: 'subvolume.delete',
      resourceKind: 'subvolume',
      successMessage: 'Subvolume 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), subvolume: subvolumeName(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/subvolume/${subvolumeName(row)}`
    },
    columns: [
      { key: 'fs', title: 'FS' },
      { key: 'group', title: 'Group' },
      { key: 'name', title: '名称' },
      { key: 'path', title: '路径' },
      { key: 'size', title: '大小' },
      { key: 'bytes_used', title: '已用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsSnapshots: {
    title: 'CephFS Snapshots',
    path: '/filesystem/subvolume/snapshots',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    createAction: {
      title: '新建 CephFS Snapshot',
      buttonLabel: '新建 Snapshot',
      path: '/filesystem/subvolume/snapshot',
      method: 'POST',
      successMessage: 'CephFS Snapshot 创建执行成功',
      fields: [
        { name: 'fs', label: 'Filesystem', required: true },
        { name: 'subvolume', label: 'Subvolume', required: true },
        { name: 'name', label: 'Snapshot 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        subvolume: String(values.subvolume ?? ''),
        name: String(values.name ?? '')
      })
    },
    columns: [
      { key: 'fs', title: 'FS' },
      { key: 'subvolume', title: 'Subvolume' },
      { key: 'name', title: 'Snapshot' },
      { key: 'created_at', title: '创建时间' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  snapshotSchedules: {
    title: 'Snapshot Schedules',
    path: '/filesystem/snapshot/schedules',
    requiredCapabilities: ['cephfs_volume'],
    createAction: {
      title: '新建 Snapshot Schedule',
      buttonLabel: '新建 Schedule',
      path: '/filesystem/snapshot/schedule',
      method: 'POST',
      successMessage: 'Snapshot Schedule 创建执行成功',
      fields: [
        { name: 'fs', label: 'Filesystem', required: true },
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
      { key: 'fs', title: 'FS' },
      { key: 'path', title: '路径' },
      { key: 'schedule', title: '计划' },
      { key: 'retention', title: '保留' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsAuthorizations: {
    title: 'CephFS Authorizations',
    path: '/filesystem/authorizations',
    requiredCapabilities: ['cephfs_volume'],
    createAction: {
      title: '新建 CephFS Authorization',
      buttonLabel: '新建授权',
      path: '/filesystem/authorization',
      method: 'POST',
      successMessage: 'CephFS 授权创建执行成功',
      fields: [
        { name: 'fs', label: 'Filesystem', required: true },
        { name: 'client', label: 'Client', required: true, placeholder: 'client.app' },
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
      { key: 'fs', title: 'FS' },
      { key: 'client', title: 'Client' },
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
        { name: 'max_bytes', label: '最大 bytes', type: 'number', required: true, min: 1 }
      ],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        fs: fsName(row),
        path: String(row?.path ?? '/'),
        max_bytes: Number(values.max_bytes)
      })
    },
    columns: [
      { key: 'fs', title: 'FS' },
      { key: 'path', title: '路径' },
      { key: 'quota', title: '配额' },
      { key: 'bytes_used', title: '已用' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  nfsClusters: {
    title: 'NFS Cluster',
    path: '/nfs/clusters',
    requiredCapabilities: ['nfs'],
    createAction: {
      title: '新建 NFS Cluster',
      buttonLabel: '新建 Cluster',
      path: '/nfs/cluster',
      method: 'POST',
      successMessage: 'NFS Cluster 创建执行成功',
      fields: [
        { name: 'name', label: 'Cluster 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    deleteAction: {
      title: '删除 NFS Cluster',
      path: '/nfs/cluster',
      action: 'nfs_cluster.delete',
      resourceKind: 'nfs_cluster',
      successMessage: 'NFS Cluster 删除执行成功',
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
    title: 'NFS Export',
    path: '/nfs/exports',
    requiredCapabilities: ['nfs'],
    rowKeyCandidates: ['natural_key', 'export_id', 'pseudo', 'name'],
    createAction: {
      title: '新建 NFS Export',
      buttonLabel: '新建 Export',
      path: '/nfs/export',
      method: 'POST',
      successMessage: 'NFS Export 创建执行成功',
      fields: [
        { name: 'cluster', label: 'NFS Cluster', required: true },
        { name: 'pseudo', label: 'Pseudo Path', required: true, placeholder: '/export' },
        { name: 'path', label: 'CephFS Path', required: true, placeholder: '/data' },
        { name: 'filesystem', label: 'Filesystem', required: true },
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
      title: '更新 NFS Export',
      path: '/nfs/export',
      method: 'PATCH',
      successMessage: 'NFS Export 更新执行成功',
      fields: [
        { name: 'cluster', label: 'NFS Cluster', required: true },
        { name: 'pseudo', label: 'Pseudo Path', required: true },
        { name: 'path', label: 'CephFS Path', required: true },
        { name: 'filesystem', label: 'Filesystem', required: true },
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
      title: '删除 NFS Export',
      path: '/nfs/export',
      action: 'nfs_export.delete',
      resourceKind: 'nfs_export',
      successMessage: 'NFS Export 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, export_id: exportId(row) }),
      resourceKey: (row) => `nfs/export/${exportId(row)}`
    },
    columns: [
      { key: 'export_id', title: 'Export ID' },
      { key: 'cluster', title: 'Cluster' },
      { key: 'pseudo', title: 'Pseudo' },
      { key: 'path', title: '路径' },
      { key: 'filesystem', title: 'Filesystem' },
      { key: 'read_only', title: '只读' },
      { key: 'status', title: '状态' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  smbClusters: {
    title: 'SMB Cluster',
    path: '/smb/clusters',
    requiredCapabilities: ['smb'],
    createAction: {
      title: '新建 SMB Cluster',
      buttonLabel: '新建 Cluster',
      path: '/smb/cluster',
      method: 'POST',
      successMessage: 'SMB Cluster 创建执行成功',
      fields: [
        { name: 'name', label: 'Cluster 名称', required: true },
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
      title: '更新 SMB Cluster',
      path: '/smb/cluster',
      method: 'PATCH',
      successMessage: 'SMB Cluster 更新执行成功',
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
      title: '删除 SMB Cluster',
      path: '/smb/cluster',
      action: 'smb_cluster.delete',
      resourceKind: 'smb_cluster',
      successMessage: 'SMB Cluster 删除执行成功',
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
    title: 'SMB Share',
    path: '/smb/shares',
    requiredCapabilities: ['smb'],
    rowKeyCandidates: ['natural_key', 'share_id', 'name'],
    createAction: {
      title: '新建 SMB Share',
      buttonLabel: '新建 Share',
      path: '/smb/share',
      method: 'POST',
      successMessage: 'SMB Share 创建执行成功',
      fields: [
        { name: 'cluster', label: 'SMB Cluster', required: true },
        { name: 'name', label: 'Share 名称', required: true },
        { name: 'filesystem', label: 'Filesystem', required: true },
        { name: 'path', label: 'CephFS Path', required: true, placeholder: '/data' }
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
      title: '更新 SMB Share',
      path: '/smb/share',
      method: 'PATCH',
      successMessage: 'SMB Share 更新执行成功',
      fields: [
        { name: 'cluster', label: 'SMB Cluster', required: true },
        { name: 'filesystem', label: 'Filesystem', required: true },
        { name: 'path', label: 'CephFS Path' }
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
      title: '删除 SMB Share',
      path: '/smb/share',
      action: 'smb_share.delete',
      resourceKind: 'smb_share',
      successMessage: 'SMB Share 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, share_id: shareId(row) }),
      resourceKey: (row) => `smb/share/${shareId(row)}`
    },
    columns: [
      { key: 'share_id', title: 'Share ID' },
      { key: 'cluster', title: 'Cluster' },
      { key: 'name', title: '名称' },
      { key: 'filesystem', title: 'Filesystem' },
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
