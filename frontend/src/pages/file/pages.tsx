import { SnapshotScheduleStatus } from './SnapshotScheduleStatus'
import { nfsExportEditReason, nfsExportInitialValues, nfsFSAL } from './nfsExportFields'
import { CephFSDirectoryBrowser } from './CephFSDirectoryBrowser'
import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'
import { listAllResources, listResource } from '../../api/resource'
import { SubvolumeSnapshotVisibility } from './SubvolumeSnapshotVisibility'
import { cephFSCreateQuota, groupUpdateBody, groupUpdateInitialValues } from './cephfsGroupForm'
import { filesystemEnabledText } from './cephfsFilesystemState'
import { formatDateTime } from '../../utils/time'
import { CephFSPermissions, CephFSUsage } from './CephFSResourceUsage'
import { cephFSBytes, cephFSQuota } from './cephfsSubvolumeSummary'
import { Alert, Space, Tag } from 'antd'
import { subvolumeReadyReason, subvolumeState, subvolumeType } from './cephfsSubvolumeState'
import { cloneFailure, cloneSource } from './cephfsCloneSummary'
import { CephFSCloneProgress } from './CephFSCloneProgress'
import { SnapshotCloneDependenciesPanel } from './SnapshotCloneDependenciesPanel'
import { snapshotDeleteReason, snapshotDependencies, snapshotPendingText } from './cephfsSnapshotDependencies'
import { CephFSSubvolumeMount } from './CephFSAttachCommands'
import { subvolumeUpdateBody, subvolumeUpdateInitialValues } from './cephfsSubvolumeForm'

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
  return <SnapshotScheduleStatus />
}

export function CephfsAuthorizationsPage() {
  return <ResourceListPage definition={definitions.cephfsAuthorizations} />
}

export function CephfsEntriesPage() {
  return <CephFSDirectoryBrowser />
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
  | 'cephfsAuthorizations'
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
        {
          name: 'name',
          label: '文件系统名称',
          required: true,
          placeholder: '例如 cephfs',
          pattern: /^(?:\.[A-Za-z0-9_-]+|[A-Za-z][.A-Za-z0-9_-]*)$/,
          patternMessage: "名称须以字母或点开头，且只能包含字母、数字、点、'-' 或 '_'"
        },
        {
          name: 'placement_type',
          label: 'MDS 放置方式',
          type: 'select',
          options: [
            { label: '指定主机', value: 'hosts' },
            { label: '按主机标签', value: 'label' }
          ]
        },
        { name: 'placement_hosts', label: 'MDS 主机（逗号分隔）', placeholder: 'ceph-node-1, ceph-node-2' },
        { name: 'placement_label', label: 'MDS 主机标签', placeholder: '例如 mds' },
        { name: 'use_existing_pools', label: '使用现有存储池', type: 'boolean' },
        { name: 'metadata_pool', label: '元数据池', type: 'select', optionsLoader: cephfsPoolOptions },
        { name: 'data_pool', label: '数据池', type: 'select', optionsLoader: cephfsPoolOptions }
      ],
      initialValues: { placement_type: 'hosts', use_existing_pools: false },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        name: String(values.name ?? ''),
        ...(filesystemPlacement(values) ? { placement: filesystemPlacement(values) } : {}),
        ...(values.use_existing_pools && values.metadata_pool ? { metadata_pool: String(values.metadata_pool) } : {}),
        ...(values.use_existing_pools && values.data_pool ? { data_pool: String(values.data_pool) } : {})
      })
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
    extraActions: [{
      title: '重命名 CephFS 文件系统卷',
      buttonLabel: '重命名卷',
      path: '/filesystem',
      method: 'PUT',
      successMessage: 'CephFS 卷重命名完成；请重新授权客户端并检查 MDS 与存储池',
      confirmation: (values, row) => `将 ${resourceName(row)} 重命名为 ${String(values.new_name)}？该操作可能中断访问、重命名关联存储池和调整 MDS 服务。客户端 CephX 权限需要重新授权；多个数据池不会自动全部重命名。请在维护窗口执行。`,
      fields: [{
        name: 'new_name', label: '新文件系统名称', required: true,
        pattern: /^(?:\.[A-Za-z0-9_-]+|[A-Za-z][.A-Za-z0-9_-]*)$/,
        patternMessage: '须以字母或点开头，仅包含字母、数字、点、横线或下划线'
      }],
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId, fs: resourceName(row), new_name: String(values.new_name), confirmed: true
      })
    }],
    detailPath: (row) => `/file/cephfs/${encodeURIComponent(resourceName(row))}`,
    columns: [
      { key: 'name', title: '名称' },
      { key: 'enabled', title: '启用状态', render: filesystemEnabledText },
      { key: 'created', title: 'Ceph 创建时间', render: (value) => formatDateTime(value) },
      { key: 'status', title: '状态' },
      { key: 'metadata_pool', title: '元数据池' },
      { key: 'data_pools', title: '数据池' },
      { key: 'max_mds', title: '最大 MDS' },
      { key: 'up', title: '活跃 MDS' },
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
      { key: 'type', title: '类型' },
      { key: 'version', title: '版本' },
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
        { name: 'fs', label: '文件系统', type: 'select', required: true, optionsLoader: filesystemOptions },
        { name: 'name', label: '子卷组名称', required: true },
        { name: 'size', label: '配额大小（字节，0 表示不限制）', pattern: /^(0|[1-9][0-9]*)$/ },
        { name: 'pool', label: 'CephFS 数据池', type: 'select', required: true, optionsDependencies: ['fs'], optionsLoader: filesystemDataPoolOptions },
        { name: 'uid', label: 'UID', type: 'number', min: 0 },
        { name: 'gid', label: 'GID', type: 'number', min: 0 },
        { name: 'mode', label: '目录权限模式', placeholder: '0755' },
        {
          name: 'normalization',
          label: '名称规范化',
          type: 'select',
          options: [
            { label: '不指定', value: '' },
            { label: 'NFD', value: 'nfd' },
            { label: 'NFC', value: 'nfc' },
            { label: 'NFKD', value: 'nfkd' },
            { label: 'NFKC', value: 'nfkc' }
          ]
        },
        { name: 'case_sensitive', label: '区分大小写', type: 'boolean' }
      ],
      initialValues: { size: '0', uid: 0, gid: 0, mode: '0755', normalization: '', case_sensitive: true },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        name: String(values.name ?? ''),
        size: cephFSCreateQuota(values.size),
        pool: String(values.pool ?? ''),
        uid: Number(values.uid ?? 0),
        gid: Number(values.gid ?? 0),
        mode: String(values.mode ?? '0755'),
        ...(values.normalization ? { normalization: String(values.normalization) } : {}),
        case_sensitive: Boolean(values.case_sensitive)
      })
    },
    updateAction: {
      title: '更新子卷组配额与属性',
      path: '/filesystem/subvolume/group',
      method: 'PATCH',
      successMessage: '子卷组更新执行成功',
      confirmation: (values) => values.edit_attributes ? '将修改子卷组自身的数据池布局、所有者和权限，不递归修改已有子卷。配额与属性分步执行，后续步骤失败时前面的修改可能已生效，确认继续？' : values.unlimited ? '确认取消该子卷组的配额限制？' : undefined,
      fields: [
        { name: 'size', label: '配额大小（十进制字节）', required: true, pattern: /^[1-9][0-9]*$/, patternMessage: '请输入正整数字节数', visibleWhen: (values) => !values.unlimited },
        { name: 'unlimited', label: '取消配额限制', type: 'boolean' },
        { name: 'no_shrink', label: '不允许配额低于已用空间', type: 'boolean', visibleWhen: (values) => !values.unlimited },
        { name: 'edit_attributes', label: '同时更新数据池、所有者与权限', type: 'boolean' },
        { name: 'pool', label: '后续子卷的数据池（不迁移既有数据）', type: 'select', required: true, optionsLoader: cephfsPoolOptions, visibleWhen: (values) => Boolean(values.edit_attributes) },
        { name: 'uid', label: 'UID', type: 'number', required: true, min: 0, max: 4294967295, visibleWhen: (values) => Boolean(values.edit_attributes) },
        { name: 'gid', label: 'GID', type: 'number', required: true, min: 0, max: 4294967295, visibleWhen: (values) => Boolean(values.edit_attributes) },
        { name: 'mode', label: '目录权限（八进制）', required: true, pattern: /^[0-7]{3,4}$/, patternMessage: '请输入 3 或 4 位八进制权限，例如 0755', visibleWhen: (values) => Boolean(values.edit_attributes) }
      ],
      initialValues: groupUpdateInitialValues,
      buildBody: (values, clusterId, row) => groupUpdateBody(values, clusterId, fsName(row), groupName(row))
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
      { key: 'filesystem', title: '文件系统' },
      { key: 'name', title: '名称' },
      { key: 'data_pool', title: '数据池' },
      { key: 'bytes_quota', title: '配额', render: cephFSQuota },
      { key: 'bytes_used', title: '已用', render: cephFSBytes },
      { key: 'bytes_pcent', title: '配额使用率', render: (value, row) => <CephFSUsage quota={row.bytes_quota} used={row.bytes_used} percent={value} /> },
      { key: 'mode', title: '权限', render: (value) => <CephFSPermissions mode={value} /> },
      { key: 'ceph_created_at', title: '创建时间', render: (value) => formatDateTime(value) },
      { key: 'resource_version', title: '版本' }
    ]
  },
  subvolumes: {
    title: '子卷',
    path: '/filesystem/subvolumes',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    detailContent: (row, clusterId) => subvolumeReadyReason(row) ? <Alert type="warning" showIcon message={subvolumeReadyReason(row)} description="当前不生成挂载命令或读取/修改子卷目录的快照可见性；保留快照仍可在 CephFS 快照页面查看和管理。" /> : <Space direction="vertical" style={{ width: '100%' }}><CephFSSubvolumeMount clusterId={clusterId} filesystem={fsName(row)} path={row.path} /><SubvolumeSnapshotVisibility key={`${clusterId}/${fsName(row)}/${groupName(row)}/${subvolumeName(row)}`} clusterId={clusterId} filesystem={fsName(row)} subvolume={subvolumeName(row)} group={groupName(row)} resourceVersion={row.resource_version == null ? undefined : String(row.resource_version)} /></Space>,
    createAction: {
      title: '新建子卷',
      buttonLabel: '新建子卷',
      path: '/filesystem/subvolume',
      method: 'POST',
      successMessage: '子卷创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', type: 'select', required: true, optionsLoader: filesystemOptions },
        { name: 'name', label: '子卷名称', required: true },
        { name: 'group', label: '子卷组', type: 'select', required: true, optionsDependencies: ['fs'], optionsLoader: snapshotGroupOptions },
        { name: 'size', label: '配额大小（字节，0 表示不限制）', pattern: /^(0|[1-9][0-9]*)$/ },
        { name: 'pool', label: 'CephFS 数据池', type: 'select', required: true, optionsDependencies: ['fs'], optionsLoader: filesystemDataPoolOptions },
        { name: 'uid', label: 'UID', type: 'number', min: 0 },
        { name: 'gid', label: 'GID', type: 'number', min: 0 },
        { name: 'mode', label: '目录权限模式', placeholder: '0755' },
        { name: 'namespace_isolated', label: '使用独立 RADOS 命名空间', type: 'boolean' },
        { name: 'earmark', label: '用途标记', placeholder: '例如 nfs 或 smb.cluster.team' },
        {
          name: 'normalization',
          label: 'Unicode 规范化',
          type: 'select',
          options: [
            { label: '不指定', value: '' },
            { label: 'NFD', value: 'nfd' },
            { label: 'NFC', value: 'nfc' },
            { label: 'NFKD', value: 'nfkd' },
            { label: 'NFKC', value: 'nfkc' }
          ]
        },
        { name: 'case_sensitive', label: '区分文件名大小写', type: 'boolean' }
      ],
      initialValues: { group: '_nogroup', size: '0', mode: '0755', namespace_isolated: false, normalization: '', case_sensitive: true },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        name: String(values.name ?? ''),
        group: String(values.group ?? '_nogroup'),
        size: cephFSCreateQuota(values.size),
        pool: String(values.pool ?? ''),
        ...(values.uid !== undefined && values.uid !== null && values.uid !== '' ? { uid: Number(values.uid) } : {}),
        ...(values.gid !== undefined && values.gid !== null && values.gid !== '' ? { gid: Number(values.gid) } : {}),
        mode: String(values.mode ?? '0755'),
        namespace_isolated: Boolean(values.namespace_isolated),
        ...(values.earmark ? { earmark: String(values.earmark) } : {}),
        ...(values.normalization ? { normalization: String(values.normalization) } : {}),
        case_sensitive: Boolean(values.case_sensitive)
      })
    },
    updateAction: {
      title: '更新子卷',
      disabledWhen: subvolumeReadyReason,
      path: '/filesystem/subvolume',
      method: 'PATCH',
      successMessage: '子卷更新执行成功',
      fields: [
        { name: 'size', label: '大小（十进制字节）', required: true, pattern: /^[1-9][0-9]*$/, patternMessage: '请输入正整数字节数', visibleWhen: (values) => !values.unlimited },
        { name: 'unlimited', label: '取消配额限制', type: 'boolean' },
        { name: 'no_shrink', label: '不允许配额低于已用空间', type: 'boolean', visibleWhen: (values) => !values.unlimited }
      ],
      initialValues: subvolumeUpdateInitialValues,
      buildBody: (values, clusterId, row) => subvolumeUpdateBody(values, clusterId, fsName(row), subvolumeName(row), groupName(row))
    },
    extraActions: [
      {
        title: '取消克隆',
        path: '/filesystem/subvolume/clone/cancel',
        method: 'POST',
        successMessage: '克隆取消请求执行成功',
        fields: [],
        confirmation: (_values, row) => `确认取消子卷 ${subvolumeName(row)} 的克隆任务吗？`,
        buildBody: (_values, clusterId, row) => ({ cluster_id: clusterId, fs: fsName(row), subvolume: subvolumeName(row), group: groupName(row) }),
        disabledWhen: (row) => ['pending', 'in-progress'].includes(String(row.clone_state ?? '')) ? undefined : '只有等待中或进行中的克隆任务可以取消'
      },
      {
        title: '高级删除',
        path: '/filesystem/subvolume',
        method: 'DELETE',
        successMessage: '子卷删除执行成功',
        fields: [
          { name: 'retain_snapshots', label: '保留已有快照', type: 'boolean' },
          { name: 'force', label: '强制删除失败或已取消的克隆', type: 'boolean' }
        ],
        initialValues: { retain_snapshots: true, force: false },
        confirmation: (values, row) => `确认删除子卷 ${subvolumeName(row)} 吗？${values.retain_snapshots ? ' 已有快照将被保留。' : ''}${values.force ? ' 将启用强制删除。' : ''}`,
        buildBody: (values, clusterId, row) => ({
          cluster_id: clusterId,
          fs: fsName(row),
          subvolume: subvolumeName(row),
          group: groupName(row),
          retain_snapshots: Boolean(values.retain_snapshots),
          force: Boolean(values.force)
        })
      }
    ],
    deleteAction: {
      title: '删除子卷',
      path: '/filesystem/subvolume',
      action: 'subvolume.delete',
      resourceKind: 'subvolume',
      successMessage: '子卷删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, fs: fsName(row), subvolume: subvolumeName(row), group: groupName(row) }),
      resourceKey: (row) => `filesystem/${fsName(row)}/subvolume/${subvolumeName(row)}`
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'group', title: '子卷组' },
      { key: 'name', title: '名称' },
      { key: 'state', title: '子卷状态', render: (value) => { const state = subvolumeState(value); return <Tag color={state.color}>{state.text}</Tag> } },
      { key: 'type', title: '类型', render: subvolumeType },
      { key: 'pool_namespace', title: 'RADOS 隔离命名空间' },
      { key: 'clone_state', title: '克隆状态', render: (value, row) => { if (row.type !== 'clone') return '—'; const state = subvolumeState(value); return <Tag color={state.color}>{state.text}</Tag> } },
      { key: 'source', title: '克隆来源', render: (value, row) => row.type === 'clone' ? cloneSource(row.clone_source ?? value) : '—' },
      { key: 'clone_progress', title: '克隆进度', render: (value) => <CephFSCloneProgress report={value} /> },
      { key: 'clone_failure', title: '克隆错误', render: cloneFailure },
      { key: 'path', title: '路径' },
      { key: 'data_pool', title: '数据池' },
      { key: 'bytes_quota', title: '配额', render: cephFSQuota },
      { key: 'bytes_used', title: '已用', render: cephFSBytes },
      { key: 'bytes_pcent', title: '配额使用率', render: (value, row) => <CephFSUsage quota={row.bytes_quota} used={row.bytes_used} percent={value} /> },
      { key: 'mode', title: '权限', render: (value) => <CephFSPermissions mode={value} /> },
      { key: 'ceph_created_at', title: '创建时间', render: (value) => formatDateTime(value) },
      { key: 'resource_version', title: '版本' }
    ]
  },
  cephfsSnapshots: {
    title: 'CephFS 快照',
    path: '/filesystem/subvolume/snapshots',
    requiredCapabilities: ['cephfs_volume'],
    rowKeyCandidates: ['natural_key', 'name'],
    detailContent: (row) => <SnapshotCloneDependenciesPanel row={row} />,
    createAction: {
      title: '新建 CephFS 快照',
      buttonLabel: '新建快照',
      path: '/filesystem/subvolume/snapshot',
      method: 'POST',
      successMessage: 'CephFS 快照创建执行成功',
      fields: [
        { name: 'fs', label: '文件系统', type: 'select', required: true, optionsLoader: filesystemOptions },
        { name: 'group', label: '子卷组', type: 'select', required: true, optionsDependencies: ['fs'], optionsLoader: snapshotGroupOptions },
        { name: 'subvolume', label: '子卷', type: 'select', required: true, optionsDependencies: ['fs', 'group'], optionsLoader: snapshotSubvolumeOptions },
        { name: 'name', label: '快照名称', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        subvolume: String(values.subvolume ?? ''),
        name: String(values.name ?? ''),
        ...(values.group ? { group: String(values.group) } : {})
      })
    },
    extraActions: [{
      title: '克隆快照',
      path: '/filesystem/subvolume/snapshot/clone',
      method: 'POST',
      successMessage: 'CephFS 快照克隆已启动',
      fields: [
        { name: 'target', label: '目标子卷名称', required: true },
        { name: 'target_group', label: '目标子卷组', type: 'select', required: true, optionsLoader: cloneTargetGroupOptions },
        { name: 'pool_layout', label: '目标数据池', type: 'select', optionsLoader: filesystemDataPoolOptions }
      ],
      initialValues: { target_group: '_nogroup' },
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        fs: fsName(row),
        subvolume: subvolumeName(row),
        snap: resourceName(row),
        target: String(values.target ?? ''),
        group: groupName(row),
        target_group: String(values.target_group ?? '_nogroup'),
        ...(values.pool_layout ? { pool_layout: String(values.pool_layout) } : {})
      })
    }],
    deleteAction: {
      title: '删除 CephFS 快照',
      disabledWhen: snapshotDeleteReason,
      path: '/filesystem/subvolume/snapshot',
      action: 'cephfs_snapshot.delete',
      resourceKind: 'cephfs_snapshot',
      successMessage: 'CephFS 快照删除执行成功',
      buildBody: (row, clusterId) => ({
        cluster_id: clusterId,
        fs: fsName(row),
        subvolume: subvolumeName(row),
        snap: resourceName(row),
        ...(row.group ? { group: String(row.group) } : {})
      }),
      resourceKey: (row) => `filesystem/${fsName(row)}/subvolume/${subvolumeName(row)}/snapshot/${resourceName(row)}`
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'subvolume', title: '子卷' },
      { key: 'group', title: '子卷组' },
      { key: 'name', title: '快照' },
      { key: 'ceph_created_at', title: '创建时间', render: (value) => formatDateTime(value) },
      { key: 'data_pool', title: '数据池' },
      { key: 'has_pending_clones', title: '克隆依赖', render: snapshotPendingText },
      { key: 'orphan_clones_count', title: '孤儿克隆记录', filterKey: false, render: (_value, row) => snapshotDependencies(row).orphans ?? '未知' },
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
        { name: 'fs', label: '文件系统', type: 'select', required: true, optionsLoader: filesystemOptions },
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
        },
        { name: 'quota', label: '允许设置布局和配额（p）', type: 'boolean', visibleWhen: (values) => values.access === 'rw' },
        { name: 'snapshot', label: '允许创建和删除快照（s）', type: 'boolean', visibleWhen: (values) => values.access === 'rw' },
        { name: 'root_squash', label: '启用 root squash', type: 'boolean' }
      ],
      initialValues: { path: '/', access: 'rw', quota: false, snapshot: false, root_squash: false },
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        fs: String(values.fs ?? ''),
        client: String(values.client ?? ''),
        ...(values.path ? { path: String(values.path) } : {}),
        access: String(values.access ?? 'rw'),
        quota: Boolean(values.quota),
        snapshot: Boolean(values.snapshot),
        root_squash: Boolean(values.root_squash)
      })
    },
    columns: [
      { key: 'fs', title: '文件系统' },
      { key: 'client', title: '客户端' },
      { key: 'path', title: '路径' },
      { key: 'access', title: '权限' },
      { key: 'permissions', title: 'MDS 权限' },
      { key: 'quota', title: '布局/配额', render: (value) => value ? '是' : '否' },
      { key: 'snapshot', title: '快照', render: (value) => value ? '是' : '否' },
      { key: 'root_squash', title: 'Root squash', render: (value) => value ? '是' : '否' },
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
      disabledWhen: nfsExportEditReason,
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
      initialValues: nfsExportInitialValues,
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
      { key: 'cluster_id', title: 'NFS 集群' },
      { key: 'pseudo', title: '伪路径' },
      { key: 'path', title: '路径' },
      { key: 'fsal_name', title: '存储后端', filterKey: false, render: (_, row) => text(nfsFSAL(row).name) },
      { key: 'fsal_filesystem', title: '文件系统', filterKey: false, render: (_, row) => text(nfsFSAL(row).fs_name) },
      { key: 'fsal_user', title: '用户', filterKey: false, render: (_, row) => text(nfsFSAL(row).user_id) },
      { key: 'access_type', title: '访问类型' },
      { key: 'protocols', title: 'NFS 协议' },
      { key: 'transports', title: '传输协议' },
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

function filesystemPlacement(values: Record<string, unknown>) {
  if (values.placement_type === 'label') {
    const label = String(values.placement_label ?? '').trim()
    return label ? `label:${label}` : ''
  }
  const hosts = String(values.placement_hosts ?? '')
    .split(',')
    .map((host) => host.trim())
    .filter(Boolean)
  return hosts.length ? hosts.join(';') : ''
}

async function filesystemOptions(clusterId: number) {
  const payload = await listAllResources('/filesystems', clusterId)
  return payload.items
    .map((row) => resourceName(row))
    .filter(Boolean)
    .map((name) => ({ label: name, value: name }))
}

async function cloneTargetGroupOptions(clusterId: number, row?: Record<string, unknown>) {
  const fs = fsName(row)
  if (!fs) throw new Error('缺少源文件系统')
  const payload = await listAllResources('/filesystem/subvolume/groups', clusterId, { body: { fs } })
  const names = new Set(['_nogroup', ...payload.items.map(resourceName).filter(Boolean)])
  return Array.from(names, (name) => ({ label: name === '_nogroup' ? '默认组（_nogroup）' : name, value: name }))
}

async function snapshotGroupOptions(clusterId: number, _row?: Record<string, unknown>, values?: Record<string, unknown>) {
  if (!values?.fs) return []
  return cloneTargetGroupOptions(clusterId, { fs: values.fs })
}

async function snapshotSubvolumeOptions(clusterId: number, _row?: Record<string, unknown>, values?: Record<string, unknown>) {
  if (!values?.fs || !values.group) return []
  const payload = await listAllResources('/filesystem/subvolumes', clusterId, { body: { fs: values.fs, group: values.group } })
  return payload.items.filter((row) => !subvolumeReadyReason(row)).map((row) => ({ label: resourceName(row), value: resourceName(row) }))
}

async function filesystemDataPoolOptions(clusterId: number, row?: Record<string, unknown>, values?: Record<string, unknown>) {
  const fs = values?.fs ? String(values.fs) : row ? fsName(row) : ''
  if (!fs) return []
  const [filesystems, pools] = await Promise.all([
    listAllResources('/filesystems', clusterId),
    listAllResources('/pools', clusterId)
  ])
  const filesystem = filesystems.items.find((item) => resourceName(item) === fs)
  const ids = new Set(Array.isArray(filesystem?.data_pools) ? filesystem.data_pools.map(String) : [])
  return pools.items.filter((pool) => pool.id != null && ids.has(String(pool.id)))
    .map((pool) => ({ label: resourceName(pool), value: resourceName(pool) }))
}

async function cephfsPoolOptions(clusterId: number) {
  const payload = await listResource('/pools', clusterId)
  return payload.items
    .filter((row) => {
      const applications = Array.isArray(row.applications) ? row.applications.map(String) : []
      return applications.includes('cephfs') || Boolean((row.application_metadata as Record<string, unknown> | undefined)?.cephfs)
    })
    .map((row) => resourceName(row))
    .filter(Boolean)
    .map((name) => ({ label: name, value: name }))
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
