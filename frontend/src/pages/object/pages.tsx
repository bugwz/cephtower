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
import { ServiceDaemons } from '../cluster/ServiceDaemons'
import { RgwQuota } from './RgwQuota'
import { RgwStorage } from './RgwStorage'
import { rgwStorageScope } from './rgwStorageDetails'
import { RgwBucketIndexDetails } from './RgwBucketIndexDetails'
import { RgwBucketPlacementDetails } from './RgwBucketPlacementDetails'
import { rgwBucketVersioning, rgwBucketBooleanState, rgwBucketReshardState } from './rgwBucketState'
import { rgwUserPolicyBlocked, rgwUserPolicyInput, rgwUserPolicyOptions } from './rgwUserPolicy'
import { RgwRateLimit } from './RgwRateLimit'
import { RgwPermissions } from './RgwPermissions'
import { rgwBucketLimit, rgwBucketLimitInput, rgwBucketLimitPatch } from './rgwBucketLimit'
import { rgwAccountLimit, rgwAccountLimitPatch } from './rgwAccountLimit'
import { rgwUserSuspension, rgwUserBooleanFlag } from './rgwUserFlags'
import { rgwUserFlagPatch } from './rgwUserFlagPatch'
import { rgwUserEmailPatch } from './rgwUserEmailPatch'
import { rgwRateLimitInitial, rgwRateLimitInput } from './rgwRateLimitForm'
import { rgwQuotaInitial, rgwQuotaInput } from './rgwQuotaForm'
import { rgwRoleInitial, rgwRolePatch } from './rgwRoleEdit'
import { RgwRoleTagsTable } from './RgwRoleTagsTable'
import { RgwPolicyDocument, RgwRolePolicyDetails, RgwRoleManagedPolicies } from './RgwRolePolicyDetails'
import { rgwPolicyChanged, rgwPolicyConfirmation, rgwPolicyDeleteOptions, rgwPolicyMutation } from './rgwRolePolicies'
import { RgwUserIdentityDetails, RgwUserPlacementDetails } from './RgwUserIdentityDetails'

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
  const { selectedClusterId } = useClusterContext()
  return <ResourceListPage key={selectedClusterId ?? 'none'} definition={definitions.gatewayManagement} />
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
        { name:'max_buckets',label:'最大 Bucket 数（-1 禁止创建，0 无限制）',type:'number',min:-1 }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        uid: String(values.uid ?? ''),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {}),
        ...rgwBucketLimitInput(values.max_buckets)
      })
    },
    updateAction: {
      title: '更新 RGW 用户',
      path: '/rgw/user',
      method: 'PATCH',
      successMessage: 'RGW 用户更新执行成功',
      fields: [
        { name: 'display_name', label: '显示名' },
        { name: 'email_action', label: '邮箱操作', type: 'select', options: [{ label: '保持不变', value: 'keep' }, { label: '设置邮箱', value: 'set' }, { label: '清空邮箱', value: 'clear' }] },
        { name: 'email', label: '邮箱', required: true, visibleWhen: (values) => values.email_action === 'set' },
        { name: 'max_buckets', label: '最大 Bucket 数（留空不修改，-1 禁止创建，0 无限制）', type: 'number', min: -1, max: 2147483647 },
        { name: 'suspended', label: '暂停用户', type: 'select', options: [{ label: '保持不变', value: 'keep' }, { label: '暂停', value: 'enable' }, { label: '解除暂停', value: 'disable' }] },
        { name: 'system', label: '系统用户', type: 'select', options: [{ label: '保持不变', value: 'keep' }, { label: '启用', value: 'enable' }, { label: '关闭', value: 'disable' }] }
      ],
      initialValues: (row) => ({
        display_name: text(row?.display_name),
        email_action: 'keep',
        email: text(row?.email),
        max_buckets: numberOrUndefined(row?.max_buckets),
        suspended: 'keep',
        system: 'keep'
      }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        uid: userId(row),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        ...rgwUserEmailPatch(values),
        ...rgwBucketLimitPatch(values.max_buckets, row?.max_buckets),
        ...rgwUserFlagPatch(values)
      })
    },
    extraActions: [...(['user', 'bucket'] as const).map<ResourceFormAction>((scope) => ({
      title: scope === 'user' ? '用户总配额' : '默认 Bucket 配额', path: '/rgw/user/quota', method: 'PUT' as const, successMessage: '用户配额更新执行成功',
      fields: [
        { name: 'enabled', label: '配额状态', type: 'select' as const, required: true, options: [{ label: '启用', value: 'enable' }, { label: '关闭', value: 'disable' }] },
        { name: 'max_size', label: '容量上限（字节，向上取整至 KiB；-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true },
        { name: 'max_objects', label: '对象数量上限（-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true }
      ],
      initialValues: (row) => rgwQuotaInitial(row?.[scope === 'user' ? 'user_quota' : 'bucket_quota']),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), scope, ...rgwQuotaInput(values) })
    })),
      { title: '用户限流设置', path: '/rgw/user/ratelimit', method: 'PUT', successMessage: '用户限流设置执行成功',
        fields: [
          { name: 'enabled', label: '限流状态', type: 'select', required: true, options: [{ label: '启用', value: 'enable' }, { label: '关闭', value: 'disable' }] },
          ...['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((name,index) => ({ name, label: ['读请求数', '写请求数', '读取字节数', '写入字节数'][index] + '（每 RGW 每分钟；0 为无限制）', type: 'number' as const, min: 0, max: Number.MAX_SAFE_INTEGER, required: true }))
        ],
        initialValues: (row) => rgwRateLimitInitial(row?.rate_limit),
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwRateLimitInput(values) })
      },
      { title: '管理用户托管策略', path: '/rgw/user/policy', method: 'POST', successMessage: '托管策略操作执行成功',
        disabledWhen: rgwUserPolicyBlocked,
        initialValues: { action: 'attach' },
        fields: [
          { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '关联策略', value: 'attach' }, { label: '解除关联', value: 'detach' }] },
          { name: 'policy_arn', label: '托管策略 ARN（由 Ceph 验证是否支持）', required: true, visibleWhen: (values) => values.action === 'attach' },
          { name: 'existing_policy', label: '已关联策略', type: 'select', required: true, visibleWhen: (values) => values.action === 'detach', optionsDependencies: ['action'], optionsLoader: async (_clusterId, row) => rgwUserPolicyOptions(row) }
        ],
        confirmation: (values, row) => {
          const policy = rgwUserPolicyInput(values, row)
          return `确认对用户 ${JSON.stringify(userId(row))} ${policy.action === 'attach' ? '关联' : '解除关联'}策略 ${JSON.stringify(policy.policy_arn)}？该操作会改变用户权限，可能导致权限扩大或现有访问失败。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwUserPolicyInput(values, row) })
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
      { key: 'user_identity', title: '身份与归属', ellipsis: false, render: (_value, row) => <RgwUserIdentityDetails row={row} /> },
      { key: 'user_placement', title: '用户放置配置', ellipsis: false, render: (_value, row) => <RgwUserPlacementDetails row={row} /> },
      { key: 'display_name', title: '显示名' },
      { key: 'email', title: '邮箱' },
      { key: 'max_buckets', title: '最大 Bucket 数', render: rgwBucketLimit },
      { key: 'suspended', title: '用户暂停状态', render: rgwUserSuspension },
      { key: 'system', title: '系统用户', render: rgwUserBooleanFlag },
      { key: 'admin', title: '管理员标志', render: rgwUserBooleanFlag },
      { key: 'status', title: '状态' },
      { key: 'caps', title: '管理权限', ellipsis: false, render: (value) => <RgwPermissions value={value} /> },
      { key: 'subusers', title: '子用户', ellipsis: false, render: (value) => <RgwPermissions value={value} subusers /> },
      { key: 'rate_limit', title: '用户限流（每 RGW）', ellipsis: false, render: (value) => <RgwRateLimit value={value} /> },
      { key: 'user_quota', title: '用户总配额', ellipsis: false, render: (value) => <RgwQuota value={value} /> },
      { key: 'bucket_quota', title: '默认 Bucket 配额', ellipsis: false, render: (value) => <RgwQuota value={value} /> },
      { key: 'stats_scope', title: '统计范围', ellipsis: false, render: (value, row) => rgwStorageScope(value, row.account_id) },
      { key: 'storage_stats', title: '容量与对象统计', ellipsis: false, render: (value) => <RgwStorage value={value} /> },
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
        ...['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((name, index) => ({ name, label: ['用户上限', '角色上限', '用户组上限', 'Bucket 上限', '每用户访问密钥上限'][index] + (name === 'max_buckets' ? '（-1 禁止创建，0 无限制）' : '（-1 无限制，0 禁止新增）'), type: 'number' as const, min: -1 }))
      ],
      initialValues: (row) => ({ account_name: text(row?.account_name), email: text(row?.email), ...Object.fromEntries(['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((key) => [key, numberOrUndefined(row?.[key])])) }),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, account_id: String(row?.account_id ?? row?.natural_key ?? ''),
        ...(values.account_name !== text(row?.account_name) ? { account_name: String(values.account_name ?? '') } : {}),
        ...(values.email !== text(row?.email) ? { email: String(values.email ?? '') } : {}),
        ...rgwAccountLimitPatch(values, row) })
    },
    extraActions: (['account', 'bucket'] as const).map((scope) => ({
      title: scope === 'account' ? '账户总配额' : '默认 Bucket 配额', path: '/rgw/account/quota', method: 'PUT' as const, successMessage: '账户配额更新执行成功',
      fields: [
        { name: 'enabled', label: '配额状态', type: 'select' as const, required: true, options: [{ label: '启用', value: 'enable' }, { label: '关闭', value: 'disable' }] },
        { name: 'max_size', label: '容量上限（字节，向上取整至 KiB；-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true },
        { name: 'max_objects', label: '对象数量上限（-1 为无限制）', type: 'number' as const, min: -1, max: Number.MAX_SAFE_INTEGER, required: true }
      ],
      initialValues: (row) => rgwQuotaInitial(row?.[scope === 'account' ? 'quota' : 'bucket_quota']),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, account_id: String(row?.account_id ?? row?.natural_key ?? ''), scope, ...rgwQuotaInput(values) })
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
      { key: 'quota', title: '账户配额', ellipsis: false, render: (value) => <RgwQuota value={value} /> },
      { key: 'bucket_quota', title: '默认 Bucket 配额', ellipsis: false, render: (value) => <RgwQuota value={value} /> },
      { key: 'storage_stats', title: '容量与对象统计', ellipsis: false, render: (value) => <RgwStorage value={value} account /> },
      { key: 'max_users', title: '用户上限', render: rgwAccountLimit },
      { key: 'max_roles', title: '角色上限', render: rgwAccountLimit },
      { key: 'max_groups', title: '用户组上限', render: rgwAccountLimit },
      { key: 'max_buckets', title: 'Bucket 上限', render: rgwBucketLimit },
      { key: 'max_access_keys', title: '每用户访问密钥上限', render: rgwAccountLimit },
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
        { name: 'assume_role_policy', label: '信任策略（JSON，留空不修改）', type: 'textarea' },
        { name: 'max_session_duration', label: '最大会话时长（秒，留空不修改）', type: 'number', min: 3600, max: 43200 }
      ],
      initialValues: (row) => rgwRoleInitial(row),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...(row?.AccountId ? { account_id: String(row.AccountId) } : {}), name: String(row?.RoleName ?? row?.natural_key ?? ''), ...rgwRolePatch(values, row) })
    },
    extraActions: [{
      title: '管理内联权限策略', path: '/rgw/role/policy', method: 'POST', successMessage: '角色权限策略操作执行成功',
      initialValues: { action: 'put' },
      changedValues: rgwPolicyChanged,
      fields: [
        { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '新增或替换', value: 'put' }, { label: '编辑已有策略', value: 'edit' }, { label: '删除', value: 'delete' }] },
        { name: 'policy_name', label: '策略名称', required: true, visibleWhen: (values) => values.action === 'put' },
        { name: 'existing_policy', label: '已有内联策略（切换将重新载入文档）', type: 'select', required: true, visibleWhen: (values) => values.action === 'delete' || values.action === 'edit', optionsDependencies: ['action'], optionsLoader: async (_clusterId, row, values) => values?.action === 'delete' || values?.action === 'edit' ? rgwPolicyDeleteOptions(row) : [] },
        { name: 'policy_document', label: '权限策略（JSON）', type: 'textarea', required: true, visibleWhen: (values) => values.action === 'put' || values.action === 'edit' }
      ],
      confirmation: rgwPolicyConfirmation,
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...(row?.AccountId ? { account_id: String(row.AccountId) } : {}), name: String(row?.RoleName ?? row?.natural_key ?? ''), ...rgwPolicyMutation(values, row) })
    }],
    deleteAction: {
      title: '删除 RGW Role', path: '/rgw/role', action: 'rgw_role.delete',
      resourceKind: 'rgw_role', successMessage: 'RGW Role 删除执行成功',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...(row.AccountId ? { account_id: String(row.AccountId) } : {}), name: String(row.RoleName ?? row.natural_key ?? '') }),
      resourceKey: (row) => `rgw/role/${row.AccountId ? `${String(row.AccountId)}/` : ''}${String(row.RoleName ?? row.natural_key ?? '')}`
    },
    columns: [
      { key: 'RoleName', title: 'Role' },
      { key: 'RoleId', title: '原生 Role ID' },
      { key: 'AccountId', title: '账户 ID' },
      { key: 'Path', title: 'Path' },
      { key: 'Description', title: '描述' },
      { key: 'Arn', title: 'ARN' },
      { key: 'AssumeRolePolicyDocument', title: '信任策略', ellipsis: false, render: (value) => <RgwPolicyDocument value={value} /> },
      { key: 'PermissionPolicies', title: '内联权限策略', ellipsis: false, render: (value) => <RgwRolePolicyDetails value={value} /> },
      { key: 'ManagedPermissionPolicies', title: '直接关联的托管策略 ARN', ellipsis: false, render: (value) => <RgwRoleManagedPolicies value={value} /> },
      { key: 'Tags', title: '角色标签', ellipsis: false, render: (value) => <RgwRoleTagsTable value={value} /> },
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
        { name: 'enabled', label: '配额状态', type: 'select', required: true, options: [{ label: '启用', value: 'enable' }, { label: '关闭', value: 'disable' }] },
        { name: 'max_size', label: '容量上限（字节，向上取整至 KiB；-1 为无限制）', type: 'number', min: -1, max: Number.MAX_SAFE_INTEGER, required: true },
        { name: 'max_objects', label: '对象上限（-1 为无限制）', type: 'number', min: -1, max: Number.MAX_SAFE_INTEGER, required: true }
      ],
      initialValues: (row) => rgwQuotaInitial(row?.bucket_quota),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, bucket_id: bucketId(row), ...rgwQuotaInput(values) })
    }, { title: 'Bucket 限流设置', path: '/rgw/bucket/ratelimit', method: 'PUT', successMessage: 'Bucket 限流设置执行成功',
        fields: [
          { name: 'enabled', label: '限流状态', type: 'select', required: true, options: [{ label: '启用', value: 'enable' }, { label: '关闭', value: 'disable' }] },
          ...['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map((name,index) => ({ name, label: ['读请求数', '写请求数', '读取字节数', '写入字节数'][index] + '（每 RGW 每分钟；0 为无限制）', type: 'number' as const, min: 0, max: Number.MAX_SAFE_INTEGER, required: true }))
        ],
        initialValues: (row) => rgwRateLimitInitial(row?.rate_limit),
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, bucket_id: bucketId(row), ...rgwRateLimitInput(values) })
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
      { key: 'versioning', title: '版本控制', ellipsis: false, render: rgwBucketVersioning },
      { key: 'bucket_index', title: '索引详情', ellipsis: false, render: (_value, row) => <RgwBucketIndexDetails row={row} /> },
      { key: 'placement_rule', title: '放置规则' },
      { key: 'explicit_placement', title: '显式存储池', ellipsis: false, render: (value) => <RgwBucketPlacementDetails value={value} /> },
      { key: 'zonegroup', title: 'Zonegroup' },
      { key: 'bucket_quota', title: 'Bucket 配额', ellipsis: false, render: (value) => <RgwQuota value={value} /> },
      { key: 'object_lock_enabled', title: '对象锁启用标记（非保留策略）', render: rgwBucketBooleanState },
      { key: 'mfa_enabled', title: 'MFA Delete', render: rgwBucketBooleanState },
      { key: 'reshard_status', title: '重新分片状态（采集时）', ellipsis: false, render: rgwBucketReshardState },
      { key: 'status', title: '状态' },
      { key: 'usage', title: '使用量', ellipsis: false, render: (value) => <RgwStorage value={value} categorized /> },
      { key: 'rate_limit', title: 'Bucket 限流（每 RGW）', ellipsis: false, render: (value) => <RgwRateLimit value={value} /> },
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
    detailContent: (row, clusterId) => clusterId && typeof row.name === 'string'
      ? <ServiceDaemons key={`${clusterId}:${row.name}`} clusterId={clusterId} name={row.name} /> : null,
    columns: [
      { key: 'name', title: '服务名' },
      { key: 'placement', title: '放置策略' },
      { key: 'running', title: '运行数' },
      { key: 'size', title: '目标数' },
      { key: 'unmanaged', title: '管理模式', filterKey: false, render: (value) => value === true ? '非托管' : value === false ? '编排器管理' : '未采集' },
      { key: 'last_refresh', title: 'Ceph 最近刷新', filterKey: false },
      { key: 'networks', title: '绑定网段', filterKey: false },
      { key: 'ports', title: '端口', filterKey: false },
      { key: 'service_url', title: '服务访问地址', filterKey: false },
      { key: 'virtual_ip', title: '虚拟 IP', filterKey: false },
      { key: 'container_image_name', title: '容器镜像', filterKey: false },
      { key: 'resource_version', title: '资源版本' }
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
        { name: 'name', label: 'ZoneGroup 名称', required: true },
        { name: 'realm', label: 'Realm（留空使用原生默认值）' },
        { name: 'endpoints', label: '端点（多个地址以逗号分隔）' },
        { name: 'master', label: '主 Zonegroup', type: 'boolean' },
        { name: 'default', label: '设为默认 Zonegroup', type: 'boolean' }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? ''), ...(values.realm ? { realm: String(values.realm).trim() } : {}), ...(values.endpoints ? { endpoints: String(values.endpoints).trim() } : {}), master: Boolean(values.master), default: Boolean(values.default) })
    },
    updateAction: {
      title: '编辑 Zonegroup', path: '/rgw/zonegroup', method: 'PATCH',
      successMessage: 'Zonegroup 更新执行成功',
      fields: [
        { name: 'new_name', label: 'Zonegroup 名称', required: true },
        { name: 'realm_id', label: '所属 Realm ID（无 Realm 时留空）' },
        { name: 'endpoints', label: '端点（逗号分隔，留空保持原值）' },
        { name: 'add_zones', label: '添加成员 Zone（逗号分隔）' },
        { name: 'remove_zones', label: '移除成员 Zone（逗号分隔）' },
        { name: 'master', label: '设为主 Zonegroup', type: 'boolean' },
        { name: 'default', label: '设为默认 Zonegroup', type: 'boolean' }
      ],
      initialValues: (row) => ({ new_name: text(row?.name), realm_id: text(row?.realm_id), endpoints: Array.isArray(row?.endpoints) ? row.endpoints.join(',') : '', master: false, default: false }),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, name: text(row?.name), new_name: String(values.new_name ?? ''), realm_id: String(values.realm_id ?? ''), ...(values.endpoints ? { endpoints: String(values.endpoints) } : {}), add_zones: String(values.add_zones ?? '').split(',').map((zone) => zone.trim()).filter(Boolean), remove_zones: String(values.remove_zones ?? '').split(',').map((zone) => zone.trim()).filter(Boolean), master: Boolean(values.master), default: Boolean(values.default) })
    },
    columns: [
      { key: 'name', title: 'ZoneGroup' },
      { key: 'status', title: '状态' },
      { key: 'id', title: 'ID' },
      { key: 'is_default', title: '当前上下文默认' },
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
        { name: 'name', label: 'Zone 名称', required: true },
        { name: 'zonegroup', label: 'Zonegroup（留空使用原生默认值）' },
        { name: 'endpoints', label: '端点（多个地址以逗号分隔）' },
        { name: 'access_key', label: '系统用户 Access Key（与 Secret Key 成对填写）', type: 'password' },
        { name: 'secret_key', label: '系统用户 Secret Key', type: 'password' },
        { name: 'archive_zone', label: '归档 Zone', type: 'boolean' },
        { name: 'sync_from_all', label: '从所有 Zone 同步', type: 'boolean' },
        { name: 'sync_from', label: '同步来源 Zone（逗号分隔）' },
        { name: 'master', label: '设为主 Zone', type: 'boolean' },
        { name: 'default', label: '设为默认 Zone', type: 'boolean' }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? ''), ...(values.zonegroup ? { zonegroup: String(values.zonegroup).trim() } : {}), ...(values.endpoints ? { endpoints: String(values.endpoints).trim() } : {}), ...(values.access_key ? { access_key: String(values.access_key) } : {}), ...(values.secret_key ? { secret_key: String(values.secret_key) } : {}), ...(values.archive_zone ? { tier_type: 'archive' } : {}), ...(values.sync_from_all !== undefined ? { sync_from_all: Boolean(values.sync_from_all) } : {}), ...(values.sync_from ? { sync_from: String(values.sync_from).trim() } : {}), master: Boolean(values.master), default: Boolean(values.default) })
    },
    updateAction: {
      title: '编辑 Zone', path: '/rgw/zone', method: 'PATCH',
      successMessage: 'Zone 更新执行成功',
      fields: [
        { name: 'new_name', label: '新名称', required: true },
        { name: 'zonegroup', label: '同步更新成员名称的 Zonegroup（留空使用默认组）' },
        { name: 'access_key', label: '新系统 Access Key（与 Secret Key 成对填写）', type: 'password' },
        { name: 'secret_key', label: '新系统 Secret Key（留空保持原值）', type: 'password' },
        { name: 'read_only_mode', label: '只读模式', type: 'select', options: [{ label: '保持原值', value: 'keep' }, { label: '开启', value: 'true' }, { label: '关闭', value: 'false' }] },
        { name: 'tier_mode', label: 'Zone 类型', type: 'select', options: [{ label: '保持原值', value: 'keep' }, { label: '普通 Zone', value: 'normal' }, { label: '归档 Zone', value: 'archive' }] },
        { name: 'master', label: '设为主 Zone', type: 'boolean' },
        { name: 'default', label: '设为默认 Zone', type: 'boolean' },
        { name: 'sync_mode', label: '从所有 Zone 同步', type: 'select', options: [{ label: '保持原值', value: 'keep' }, { label: '开启', value: 'true' }, { label: '关闭', value: 'false' }] },
        { name: 'sync_from', label: '来源 Zone（逗号分隔；开启全部同步时表示移除这些来源）' },
        { name: 'endpoints', label: '新端点（逗号分隔，留空保持原值）' },
        { name: 'realm_id', label: '提交 Period 的 Realm ID（无 Realm 时留空）' }
      ],
      initialValues: (row) => ({ new_name: text(row?.name), realm_id: text(row?.realm_id), zonegroup: Array.isArray(row?.zonegroup_memberships) && row.zonegroup_memberships.length === 1 ? text((row.zonegroup_memberships[0] as ApiRecord).zonegroup_name) : '' }),
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, name: text(row?.name), new_name: String(values.new_name ?? ''), zonegroup: String(values.zonegroup ?? ''), realm_id: String(values.realm_id ?? ''), ...(values.access_key ? { access_key: String(values.access_key) } : {}), ...(values.secret_key ? { secret_key: String(values.secret_key) } : {}), ...(values.read_only_mode === 'true' || values.read_only_mode === 'false' ? { read_only: values.read_only_mode === 'true' } : {}), ...(values.tier_mode === 'normal' || values.tier_mode === 'archive' ? { tier_type: values.tier_mode === 'normal' ? '' : 'archive' } : {}), master: Boolean(values.master), default: Boolean(values.default), ...(values.sync_mode === 'true' || values.sync_mode === 'false' ? { sync_from_all: values.sync_mode === 'true' } : {}), ...(values.sync_from ? { sync_from: String(values.sync_from).trim() } : {}), ...(values.endpoints ? { endpoints: String(values.endpoints).trim() } : {}) })
    },
    columns: [
      { key: 'name', title: 'Zone' },
      { key: 'status', title: '状态' },
      { key: 'id', title: 'ID' },
      { key: 'is_default', title: '当前上下文默认' },
      { key: 'realm_id', title: 'Realm ID' },
      { key: 'zonegroup_memberships', title: '所属 Zonegroup、端点与同步配置' },
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
  return typeof value === 'number' && Number.isInteger(value) && value >= -2147483648 && value <= 2147483647
    ? value : undefined
}

function parseJSONDocument(value: unknown) {
  return JSON.parse(String(value || '{}')) as unknown
}
