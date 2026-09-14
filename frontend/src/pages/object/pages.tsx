import type { ApiRecord } from '../../api/client'
import { Alert, Button, Modal } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { mutateResource } from '../../api/resource'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { ResourceListPage, type ResourceListPageDefinition, type ResourceFormAction } from '../ResourceListPage'

export function RgwOverviewPage() {
  return <ResourceListPage definition={definitions.rgwOverview} />
}

export function RgwUsersPage() {
  return <ResourceListPage definition={definitions.rgwUsers} />
}

export function RgwAccountsPage() {
  return <ResourceListPage definition={definitions.rgwAccounts} />
}

export function RgwRolesPage() {
  return <ResourceListPage definition={definitions.rgwRoles} />
}

export function BucketManagementPage() {
  return <ResourceListPage definition={definitions.bucketManagement} />
}

export function BucketPolicyPage() {
  return <ExternalListPage definition={externalDefinitions.bucketPolicy} />
}

export function GatewayManagementPage() {
  return <ResourceListPage definition={definitions.gatewayManagement} />
}

export function MultisitePage() {
  return <ResourceListPage definition={definitions.multisite} />
}

export function RgwZonegroupsPage() {
  return <ResourceListPage definition={definitions.rgwZonegroups} />
}

export function RgwZonesPage() {
  return <ResourceListPage definition={definitions.rgwZones} />
}

export function RgwPeriodPage() {
  return <PeriodCommitPanel />
}

export function ObjectStorageConfigPage() {
  return <ResourceListPage definition={definitions.objectStorageConfig} />
}

const definitions: Record<
  | 'rgwOverview'
  | 'rgwUsers'
  | 'rgwAccounts'
  | 'rgwRoles'
  | 'bucketManagement'
  | 'gatewayManagement'
  | 'multisite'
  | 'rgwZonegroups'
  | 'rgwZones'
  | 'objectStorageConfig',
  ResourceListPageDefinition
> = {
  rgwOverview: {
    title: 'RGW 状态',
    path: '/rgw/status',
    requiredCapabilities: ['rgw_admin'],
    columns: [
      { key: 'realms', title: 'Realms' },
      { key: 'global_rate_limit', title: '默认 Realm 全局限流（每 RGW 每分钟）' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwUsers: {
    title: 'RGW 用户',
    path: '/rgw/users',
    requiredCapabilities: ['rgw_admin'],
    rowKeyCandidates: ['natural_key', 'uid', 'user_id'],
    createAction: {
      title: '新建 RGW 用户',
      buttonLabel: '新建用户',
      path: '/rgw/user',
      method: 'POST',
      successMessage: 'RGW 用户创建执行成功',
      fields: [
        { name: 'uid', label: 'UID', required: true },
        { name: 'display_name', label: '显示名', required:true },
        { name: 'email', label: '邮箱' },
        { name:'max_buckets',label:'最大 Bucket 数（-1 为无限制）',type:'number',min:-1 }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        uid: String(values.uid ?? ''),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {}),
        ...(values.max_buckets != null ? {max_buckets:Number(values.max_buckets)} : {})
      })
    },
    updateAction: {
      title: '更新 RGW 用户',
      path: '/rgw/user',
      method: 'PATCH',
      successMessage: 'RGW 用户更新执行成功',
      fields: [
        { name: 'display_name', label: '显示名' },
        { name: 'email', label: '邮箱' },
        { name: 'max_buckets', label: '最大 Bucket 数（-1 为无限制）', type: 'number', min: -1 },
        { name: 'suspended', label: '暂停用户', type: 'boolean' },
        { name: 'system', label: '系统用户', type: 'boolean' }
      ],
      initialValues: (row) => ({
        display_name: text(row?.display_name),
        email: text(row?.email),
        max_buckets: numberOrUndefined(row?.max_buckets),
        suspended: row?.suspended === true || row?.suspended === 1,
        system: row?.system === true
      }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        uid: userId(row),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        email: String(values.email ?? ''),
        ...(values.max_buckets !== undefined ? { max_buckets: Number(values.max_buckets) } : {}),
        suspended: Boolean(values.suspended),
        system: Boolean(values.system)
      })
    },
    extraActions: [...(['user', 'bucket'] as const).map<ResourceFormAction>((scope) => ({
      title: scope === 'user' ? '用户总配额' : '默认 Bucket 配额', path: '/rgw/user/quota', method: 'PUT' as const, successMessage: '用户配额更新执行成功',
      fields: [
        { name: 'enabled', label: '启用配额', type: 'boolean' as const },
        { name: 'max_size', label: '容量上限（字节，向上取整至 KiB；-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true },
        { name: 'max_objects', label: '对象数量上限（-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true }
      ],
      initialValues: (row) => { const quota = row?.[scope === 'user' ? 'user_quota' : 'bucket_quota'] as ApiRecord | undefined; return { enabled: quota?.enabled === true, max_size: Number(quota?.max_size ?? -1), max_objects: Number(quota?.max_objects ?? -1) } },
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), scope, enabled: Boolean(values.enabled), max_size: Number(values.max_size), max_objects: Number(values.max_objects) })
    })),
      { title: '用户限流设置', path: '/rgw/user/ratelimit', method: 'PUT', successMessage: '用户限流设置执行成功',
        fields: [
          { name: 'enabled', label: '启用限流', type: 'boolean' },
          ...['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((name,index) => ({ name, label: ['读请求数', '写请求数', '读取字节数', '写入字节数'][index] + '（每 RGW 每分钟；0 为无限制）', type: 'number' as const, min: 0, max: Number.MAX_SAFE_INTEGER, required: true }))
        ],
        initialValues: (row) => { const limits = row?.rate_limit as ApiRecord | undefined; return { enabled: limits?.enabled === true, ...Object.fromEntries(['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((key) => [key, Number(limits?.[key] ?? 0)])) } },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), enabled: Boolean(values.enabled), ...Object.fromEntries(['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((key) => [key, Number(values[key])])) })
      },
      { title: '管理用户权限（caps）', path: '/rgw/user/caps', method: 'POST', successMessage: '用户管理权限操作执行成功',
        initialValues: { action: 'add', permission: 'read' },
        fields: [
          { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '添加权限', value: 'add' }, { label: '移除权限', value: 'rm' }] },
          { name: 'type', label: '权限类型', required: true, placeholder: '例如 users、buckets、usage' },
          { name: 'permission', label: '权限', type: 'select', required: true, options: [{ label: '读', value: 'read' }, { label: '写', value: 'write' }, { label: '读写', value: 'read,write' }, { label: '全部', value: '*' }] }
        ],
        confirmation: () => '确认修改该用户的 RGW 管理权限？',
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), action: String(values.action), type: String(values.type ?? ''), permission: String(values.permission) })
      }],
    deleteAction: {
      title: '删除 RGW 用户',
      path: '/rgw/user',
      action: 'rgw_user.delete',
      resourceKind: 'rgw_user',
      successMessage: 'RGW 用户删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, uid: userId(row) }),
      resourceKey: (row) => `rgw/user/${userId(row)}`
    },
    columns: [
      { key: 'uid', title: 'UID' },
      { key: 'display_name', title: '显示名' },
      { key: 'email', title: '邮箱' },
      { key: 'status', title: '状态' },
      { key: 'caps', title: 'Caps' },
      { key: 'rate_limit', title: '用户限流' },
      { key: 'user_quota', title: '用户总配额' },
      { key: 'bucket_quota', title: '默认 Bucket 配额' },
      { key: 'stats_scope', title: '统计范围', render:(value)=>value === 'account' ? '所属账户' : '用户' },
      { key: 'storage_stats', title: '容量与对象统计' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwAccounts: {
    title: 'RGW Accounts',
    path: '/rgw/accounts',
    requiredCapabilities: ['rgw_admin'],
    createAction: {
      title: '新建 RGW Account',
      buttonLabel: '新建 Account',
      path: '/rgw/account',
      method: 'POST',
      successMessage: 'RGW Account 创建执行成功',
      fields: [
        { name: 'account_id', label: 'Account ID', required: true },
        { name: 'account_name', label: 'Account Name' },
        { name: 'email', label: 'Email' },
        { name: 'tenant', label: 'Tenant' }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        account_id: String(values.account_id ?? ''),
        ...(values.account_name ? { account_name: String(values.account_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {}),
        ...(values.tenant ? { tenant: String(values.tenant) } : {})
      })
    },
    updateAction: {
      title: '编辑 RGW Account', path: '/rgw/account', method: 'PATCH', successMessage: '账户更新执行成功',
      fields: [
        { name: 'account_name', label: '账户名称（不支持清空）' },
        { name: 'email', label: '邮箱（不支持清空）' },
        ...['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((name, index) => ({ name, label: ['用户上限', '角色上限', '用户组上限', 'Bucket 上限', '访问密钥上限'][index] + '（-1 为无限制）', type: 'number' as const, min: -1 }))
      ],
      initialValues: (row) => ({ account_name: text(row?.account_name), email: text(row?.email), ...Object.fromEntries(['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((key) => [key, numberOrUndefined(row?.[key])])) }),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, account_id: String(row?.account_id ?? row?.natural_key ?? ''),
        ...(values.account_name !== text(row?.account_name) ? { account_name: String(values.account_name ?? '') } : {}),
        ...(values.email !== text(row?.email) ? { email: String(values.email ?? '') } : {}),
        ...Object.fromEntries(['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].filter((key) => values[key] != null).map((key) => [key, Number(values[key])])) })
    },
    extraActions: (['account', 'bucket'] as const).map((scope) => ({
      title: scope === 'account' ? '账户总配额' : '默认 Bucket 配额', path: '/rgw/account/quota', method: 'PUT' as const, successMessage: '账户配额更新执行成功',
      fields: [
        { name: 'enabled', label: '启用配额', type: 'boolean' as const },
        { name: 'max_size', label: '容量上限（字节，向上取整至 KiB；-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true },
        { name: 'max_objects', label: '对象数量上限（-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true }
      ],
      initialValues: (row) => { const quota = row?.[scope === 'account' ? 'quota' : 'bucket_quota'] as ApiRecord | undefined; return { enabled: quota?.enabled === true, max_size: Number(quota?.max_size ?? -1), max_objects: Number(quota?.max_objects ?? -1) } },
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, account_id: String(row?.account_id ?? row?.natural_key ?? ''), scope, enabled: Boolean(values.enabled), max_size: Number(values.max_size), max_objects: Number(values.max_objects) })
    })),
    deleteAction: {
      title: '删除 RGW Account', path: '/rgw/account', action: 'rgw_account.delete', resourceKind: 'rgw_account',
      successMessage: 'RGW Account 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, account_id: String(row.account_id ?? row.natural_key ?? '') }),
      resourceKey: (row) => `rgw/account/${String(row.account_id ?? row.natural_key ?? '')}`
    },
    columns: [
      { key: 'account_id', title: 'Account ID' },
      { key: 'account_name', title: '名称' },
      { key: 'email', title: '邮箱' },
      { key: 'tenant', title: 'Tenant' },
      { key: 'quota', title: '账户配额' },
      { key: 'bucket_quota', title: '默认 Bucket 配额' },
      { key: 'storage_stats', title: '容量与对象统计' },
      { key: 'max_users', title: '用户上限' },
      { key: 'max_roles', title: '角色上限' },
      { key: 'max_groups', title: '用户组上限' },
      { key: 'max_buckets', title: 'Bucket 上限' },
      { key: 'max_access_keys', title: '访问密钥上限' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwRoles: {
    title: 'RGW Roles',
    path: '/rgw/roles',
    requiredCapabilities: ['rgw_admin'],
    createAction: {
      title: '新建 RGW Role',
      buttonLabel: '新建 Role',
      path: '/rgw/role',
      method: 'POST',
      successMessage: 'RGW Role 创建执行成功',
      fields: [
        { name: 'name', label: 'Role 名称', required: true },
        { name: 'account_id', label: '账户 ID（留空使用默认租户）' },
        { name: 'path', label: 'Path' },
        { name: 'description', label: '描述', type: 'textarea' },
        { name: 'max_session_duration', label: '最大会话时长（秒，默认 3600）', type: 'number', min: 3600, max: 43200 },
        { name: 'assume_role_policy', label: 'Assume Role Policy（JSON）', type: 'textarea', required: true }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        ...(values.account_id ? { account_id: String(values.account_id) } : {}),
        name: String(values.name ?? ''),
        ...(values.path ? { path: String(values.path) } : {}),
        ...(values.description ? { description: String(values.description) } : {}),
        ...(values.max_session_duration != null ? { max_session_duration: Number(values.max_session_duration) } : {}),
        ...(values.assume_role_policy ? { assume_role_policy: String(values.assume_role_policy) } : {})
      })
    },
    updateAction: {
      title: '更新 RGW Role', path: '/rgw/role', method: 'PATCH', successMessage: 'RGW Role 更新执行成功',
      fields: [
        { name: 'assume_role_policy', label: '信任策略（JSON）', type: 'textarea', required: true },
        { name: 'max_session_duration', label: '最大会话时长（秒）', type: 'number', min: 3600, max: 43200, required: true }
      ],
      initialValues: (row) => ({ assume_role_policy: typeof row?.AssumeRolePolicyDocument === 'string' ? row.AssumeRolePolicyDocument : JSON.stringify(row?.AssumeRolePolicyDocument ?? {}, null, 2), max_session_duration: Number(row?.MaxSessionDuration ?? 3600) }),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...(row?.AccountId ? { account_id: String(row.AccountId) } : {}), name: String(row?.RoleName ?? row?.natural_key ?? ''), ...(String(values.assume_role_policy ?? '') !== (typeof row?.AssumeRolePolicyDocument === 'string' ? row.AssumeRolePolicyDocument : JSON.stringify(row?.AssumeRolePolicyDocument ?? {}, null, 2)) ? { assume_role_policy: String(values.assume_role_policy ?? '') } : {}), max_session_duration: Number(values.max_session_duration) })
    },
    extraActions: [{
      title: '管理内联权限策略', path: '/rgw/role/policy', method: 'POST', successMessage: '角色权限策略操作执行成功',
      initialValues: { action: 'put' },
      fields: [
        { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '新增或替换', value: 'put' }, { label: '删除', value: 'delete' }] },
        { name: 'policy_name', label: '策略名称', required: true },
        { name: 'policy_document', label: '权限策略（JSON）', type: 'textarea', required: true, visibleWhen: (values) => values.action === 'put' }
      ],
      confirmation: (values) => values.action === 'delete' ? '确认删除该角色的指定内联权限策略？' : '同名策略将被替换，确认提交？',
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...(row?.AccountId ? { account_id: String(row.AccountId) } : {}), name: String(row?.RoleName ?? row?.natural_key ?? ''), action: String(values.action), policy_name: String(values.policy_name ?? ''), ...(values.action === 'put' ? { policy_document: String(values.policy_document ?? '') } : {}) })
    }],
    deleteAction: {
      title: '删除 RGW Role', path: '/rgw/role', action: 'rgw_role.delete',
      resourceKind: 'rgw_role', successMessage: 'RGW Role 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...(row.AccountId ? { account_id: String(row.AccountId) } : {}), name: String(row.RoleName ?? row.natural_key ?? '') }),
      resourceKey: (row) => `rgw/role/${row.AccountId ? `${String(row.AccountId)}/` : ''}${String(row.RoleName ?? row.natural_key ?? '')}`
    },
    columns: [
      { key: 'RoleName', title: 'Role' },
      { key: 'AccountId', title: '账户 ID' },
      { key: 'Path', title: 'Path' },
      { key: 'Description', title: '描述' },
      { key: 'Arn', title: 'ARN' },
      { key: 'AssumeRolePolicyDocument', title: '信任策略' },
      { key: 'PermissionPolicies', title: '内联权限策略' },
      { key: 'ManagedPermissionPolicies', title: '托管权限策略' },
      { key: 'MaxSessionDuration', title: '最大会话时长（秒）' },
      { key: 'CreateDate', title: '创建时间' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  bucketManagement: {
    title: 'Bucket 管理',
    path: '/rgw/buckets',
    requiredCapabilities: ['rgw_admin'],
    rowKeyCandidates: ['natural_key', 'bucket_id', 'name'],
    createAction: {
      title: '新建 Bucket',
      buttonLabel: '新建 Bucket',
      path: '/rgw/bucket',
      method: 'POST',
      successMessage: 'Bucket 创建执行成功',
      fields: [
        { name: 'name', label: 'Bucket 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    updateAction: {
      title: '更新 Bucket',
      path: '/rgw/bucket',
      method: 'PATCH',
      successMessage: 'Bucket 更新执行成功',
      fields: [
        {
          name: 'versioning',
          label: '版本控制',
          type: 'select',
          required: true,
          options: [
            { label: '启用', value: 'enabled' },
            { label: '暂停', value: 'suspended' }
          ]
        }
      ],
      initialValues: (row) => ({ versioning: row?.versioning === 'suspended' ? 'suspended' : 'enabled' }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        bucket_id: bucketId(row),
        versioning: String(values.versioning ?? 'enabled')
      })
    },
    extraActions: [{ title: 'Bucket 配额设置', path: '/rgw/bucket/quota', method: 'PUT', successMessage: 'Bucket 配额设置执行成功',
      fields: [
        { name: 'enabled', label: '启用配额', type: 'boolean' },
        { name: 'max_size', label: '容量上限（字节，向上取整至 KiB；-1 为无限制）', type: 'number', min: -1, max: Number.MAX_SAFE_INTEGER, required: true },
        { name: 'max_objects', label: '对象上限（-1 为无限制）', type: 'number', min: -1, max: Number.MAX_SAFE_INTEGER, required: true }
      ],
      initialValues: (row) => { const quota=row?.bucket_quota as ApiRecord | undefined; return { enabled: quota?.enabled === true, max_size: Number(quota?.max_size ?? -1), max_objects: Number(quota?.max_objects ?? -1) } },
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, bucket_id: bucketId(row), enabled: Boolean(values.enabled), max_size: Number(values.max_size), max_objects: Number(values.max_objects) })
    }, { title: 'Bucket 限流设置', path: '/rgw/bucket/ratelimit', method: 'PUT', successMessage: 'Bucket 限流设置执行成功',
        fields: [
          { name: 'enabled', label: '启用限流', type: 'boolean' },
          ...['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((name,index) => ({ name, label: ['读请求数', '写请求数', '读取字节数', '写入字节数'][index] + '（每 RGW 每分钟；0 为无限制）', type: 'number' as const, min: 0, max: Number.MAX_SAFE_INTEGER, required: true }))
        ],
        initialValues: (row) => { const limits = row?.rate_limit as ApiRecord | undefined; return { enabled: limits?.enabled === true, ...Object.fromEntries(['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((key) => [key, Number(limits?.[key] ?? 0)])) } },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, bucket_id: bucketId(row), enabled: Boolean(values.enabled), ...Object.fromEntries(['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((key) => [key, Number(values[key])])) })
      }],
    deleteAction: {
      title: '删除 Bucket',
      path: '/rgw/bucket',
      action: 'rgw_bucket.delete',
      resourceKind: 'rgw_bucket',
      successMessage: 'Bucket 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, bucket_id: bucketId(row) }),
      resourceKey: (row) => `rgw/bucket/${bucketId(row)}`
    },
    columns: [
      { key: 'name', title: 'Bucket' },
      { key: 'tenant', title: '租户' },
      { key: 'owner', title: 'Owner' },
      { key: 'versioning', title: '版本控制' },
      { key: 'num_shards', title: '索引分片数' },
      { key: 'placement_rule', title: '放置规则' },
      { key: 'zonegroup', title: 'Zonegroup' },
      { key: 'bucket_quota', title: 'Bucket 配额' },
      { key: 'object_lock_enabled', title: '对象锁' },
      { key: 'mfa_enabled', title: 'MFA' },
      { key: 'reshard_status', title: '重新分片状态' },
      { key: 'status', title: '状态' },
      { key: 'usage', title: '使用量' },
      { key: 'rate_limit', title: 'Bucket 限流' },
      { key: 'id', title: '原生 Bucket ID' },
      { key: 'creation_time', title: '创建时间' },
      { key: 'mtime', title: '修改时间' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  gatewayManagement: {
    title: 'RGW 网关',
    path: '/services',
    body: { service_type: 'rgw' },
    columns: [
      { key: 'name', title: '名称' },
      { key: 'service_name', title: '服务' },
      { key: 'status', title: '状态' },
      { key: 'placement', title: '放置策略' },
      { key: 'running', title: '运行数' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  multisite: {
    title: 'RGW Multisite',
    path: '/rgw/realms',
    requiredCapabilities: ['rgw_admin'],
    createAction: {
      title: '新建 Realm',
      buttonLabel: '新建 Realm',
      path: '/rgw/realm',
      method: 'POST',
      successMessage: 'Realm 创建执行成功',
      fields: [
        { name: 'name', label: 'Realm 名称', required: true },
        { name: 'default', label: '设为默认 Realm', type: 'boolean' }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? ''), default: Boolean(values.default) })
    },
    updateAction: {
      title: '编辑 Realm', path: '/rgw/realm', method: 'PATCH',
      successMessage: 'Realm 更新执行成功',
      fields: [
        { name: 'new_name', label: 'Realm 名称', required: true },
        { name: 'default', label: '设为默认 Realm', type: 'boolean' }
      ],
      initialValues: (row) => ({ new_name: text(row?.name), default: false }),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, name: text(row?.name), new_name: String(values.new_name ?? ''), default: Boolean(values.default) })
    },
    columns: [
      { key: 'name', title: 'Realm' },
      { key: 'status', title: '状态' },
      { key: 'id', title: 'ID' },
      { key: 'is_default', title: '默认 Realm' },
      { key: 'current_period', title: 'Current Period' },
      { key: 'epoch', title: 'Epoch' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwZonegroups: {
    title: 'RGW ZoneGroups',
    path: '/rgw/zonegroups',
    requiredCapabilities: ['rgw_admin'],
    createAction: {
      title: '新建 ZoneGroup',
      buttonLabel: '新建 ZoneGroup',
      path: '/rgw/zonegroup',
      method: 'POST',
      successMessage: 'ZoneGroup 创建执行成功',
      fields: [
        { name: 'name', label: 'ZoneGroup 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    columns: [
      { key: 'name', title: 'ZoneGroup' },
      { key: 'status', title: '状态' },
      { key: 'id', title: 'ID' },
      { key: 'realm_id', title: 'Realm ID' },
      { key: 'api_name', title: 'API 名称' },
      { key: 'is_master', title: '主 Zonegroup' },
      { key: 'master_zone', title: 'Master Zone' },
      { key: 'endpoints', title: '端点' },
      { key: 'zones', title: '成员 Zone' },
      { key: 'placement_targets', title: '放置目标' },
      { key: 'default_placement', title: '默认放置目标' },
      { key: 'hostnames', title: '主机名' },
      { key: 'hostnames_s3website', title: '静态网站主机名' },
      { key: 'sync_policy', title: '同步策略' },
      { key: 'enabled_features', title: '启用特性' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwZones: {
    title: 'RGW Zones',
    path: '/rgw/zones',
    requiredCapabilities: ['rgw_admin'],
    createAction: {
      title: '新建 Zone',
      buttonLabel: '新建 Zone',
      path: '/rgw/zone',
      method: 'POST',
      successMessage: 'Zone 创建执行成功',
      fields: [
        { name: 'name', label: 'Zone 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    columns: [
      { key: 'name', title: 'Zone' },
      { key: 'status', title: '状态' },
      { key: 'id', title: 'ID' },
      { key: 'realm_id', title: 'Realm ID' },
      { key: 'placement_pools', title: '放置池与存储类别' },
      { key: 'domain_root', title: '元数据根池' },
      { key: 'control_pool', title: '控制池' },
      { key: 'gc_pool', title: '垃圾回收池' },
      { key: 'lc_pool', title: '生命周期池' },
      { key: 'log_pool', title: '日志池' },
      { key: 'reshard_pool', title: '重分片池' },
      { key: 'tier_config', title: '分层配置' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  objectStorageConfig: {
    title: '对象存储配置',
    path: '/configuration/values',
    body: { who: 'client.rgw' },
    updateAction: {
      title: '更新配置项',
      path: '/configuration/value',
      method: 'PUT',
      successMessage: '配置更新执行成功',
      fields: [
        { name: 'value', label: '配置值', type: 'textarea', required: true }
      ],
      initialValues: (row) => ({ value: text(row?.value) }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        who: String(row?.who ?? 'client.rgw'),
        name: String(row?.name ?? ''),
        value: String(values.value ?? '')
      })
    },
    columns: [
      { key: 'who', title: 'Who' },
      { key: 'name', title: '配置项' },
      { key: 'value', title: '值' },
      { key: 'level', title: '级别' },
      { key: 'resource_version', title: '版本' }
    ]
  }
}

const externalDefinitions: Record<'bucketPolicy', ExternalListPageDefinition> = {
  bucketPolicy: {
    title: 'Bucket Policy',
    path: '/rgw/bucket/policy',
    requiredCapabilities: ['rgw_admin'],
    requiredEndpoints: ['s3'],
    rowKeyCandidates: ['bucket_id', 'name', 'kind'],
    filterFields: [
      { name: 'bucket_id', label: 'Bucket ID', required: true }
    ],
    createAction: {
      title: '更新 Bucket Policy',
      buttonLabel: '更新配置',
      path: '/rgw/bucket/policy',
      method: 'PATCH',
      successMessage: 'Bucket Policy 更新执行成功',
      fields: [
        { name: 'bucket_id', label: 'Bucket ID', required: true },
        {
          name: 'kind',
          label: '配置类型',
          type: 'select',
          required: true,
          options: [
            { label: 'Policy', value: 'policy' },
            { label: 'CORS', value: 'cors' },
            { label: 'Lifecycle', value: 'lifecycle' },
            { label: 'Encryption', value: 'encryption' }
          ]
        },
        { name: 'document_json', label: 'JSON 文档', type: 'textarea', required: true }
      ],
      initialValues: { kind: 'policy', document_json: '{}' },
      buildBody: (values, clusterId) => {
        const kind = String(values.kind ?? 'policy')
        return {
          cluster_id: clusterId,
          bucket_id: String(values.bucket_id ?? ''),
          kind,
          [kind]: parseJSONDocument(values.document_json)
        }
      }
    },
    columns: [
      { key: 'bucket_id', title: 'Bucket ID' },
      { key: 'name', title: 'Bucket' },
      { key: 'policy', title: 'Policy' },
      { key: 'cors', title: 'CORS' },
      { key: 'lifecycle', title: 'Lifecycle' },
      { key: 'encryption', title: 'Encryption' }
    ]
  }
}

function userId(row?: Record<string, unknown>) {
  return String(row?.uid ?? row?.user_id ?? row?.natural_key ?? row?.name ?? '').trim()
}

function bucketId(row?: Record<string, unknown>) {
  return String(row?.natural_key ?? row?.bucket_id ?? '').trim()
}

function PeriodCommitPanel() {
  const { selectedClusterId } = useClusterContext()
  const operationMutation = useMutationOperation()
  const featureStatus = useFeatureRequirements(selectedClusterId, { requiredCapabilities: ['rgw_admin'] })
  const blocked = featureStatus.loading || featureStatus.blocked || Boolean(featureStatus.error)

  async function commitPeriod() {
    if (!selectedClusterId || blocked) {
      message.error('请先选择集群')
      return
    }
    const parameters = {
      cluster_id: selectedClusterId
    }
    Modal.confirm({
      title: '提交 RGW Period',
      content: '该操作为高风险操作，确认后将直接执行操作。',
      okText: '提交',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        await operationMutation.run(() => mutateResource('/rgw/period/commit', 'POST', parameters), false)
        window.setTimeout(() => {
          message.success('RGW Period commit 执行成功')
        })
      }
    })
  }

  return (
    <div className="page-embedded-list">
      <div className="page-embedded-list-head">
        <span className="page-embedded-list-title">RGW Period</span>
        <div className="page-embedded-list-actions">
          <Button type="primary" danger icon={<ReloadOutlined />} disabled={!selectedClusterId || blocked} onClick={commitPeriod}>提交 Period</Button>
        </div>
      </div>
      <div className="page-embedded-list-body">
        <FeatureRequirementAlert status={featureStatus} />
      </div>
    </div>
  )
}

function FeatureRequirementAlert({ status }: { status: ReturnType<typeof useFeatureRequirements> }) {
  if (status.loading) {
    return <Alert type="info" showIcon message="正在校验当前集群的功能依赖" />
  }
  if (status.error) {
    return <Alert type="warning" showIcon message="功能依赖检查失败" description={status.error} />
  }
  if (status.reasons.length) {
    return <Alert type="warning" showIcon message="当前集群暂不可执行该页面的变更操作" description={status.reasons.join('; ')} />
  }
  return null
}

function text(value: unknown) {
  return value === null || value === undefined ? '' : String(value)
}

function numberOrUndefined(value: unknown) {
  const parsed = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

function parseJSONDocument(value: unknown) {
  return JSON.parse(String(value || '{}')) as unknown
}
