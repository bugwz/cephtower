import { Alert, Button, Modal } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import { mutateResource } from '../../api/resource'
import { useFeatureRequirements } from '../../hooks/useFeatureRequirements'
import { useMutationOperation } from '../../hooks/useMutationOperation'
import { useClusterContext } from '../../state/ClusterContext'
import { message } from '../../utils/appMessage'
import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { ResourceListPage, type ResourceListPageDefinition } from '../ResourceListPage'

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
      { key: 'name', title: '名称' },
      { key: 'status', title: '状态' },
      { key: 'realm', title: 'Realm' },
      { key: 'zonegroup', title: 'ZoneGroup' },
      { key: 'zone', title: 'Zone' },
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
        { name: 'display_name', label: '显示名' },
        { name: 'email', label: '邮箱' }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        uid: String(values.uid ?? ''),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {})
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
        { name: 'max_buckets', label: '最大 Bucket 数', type: 'number', min: 0 },
        { name: 'suspended', label: '暂停用户', type: 'boolean' },
        { name: 'system', label: '系统用户', type: 'boolean' }
      ],
      initialValues: (row) => ({
        display_name: text(row?.display_name),
        email: text(row?.email),
        max_buckets: numberOrUndefined(row?.max_buckets),
        suspended: row?.suspended === true,
        system: row?.system === true
      }),
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        uid: userId(row),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {}),
        ...(values.max_buckets !== undefined ? { max_buckets: Number(values.max_buckets) } : {}),
        suspended: Boolean(values.suspended),
        system: Boolean(values.system)
      })
    },
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
    columns: [
      { key: 'account_id', title: 'Account ID' },
      { key: 'account_name', title: '名称' },
      { key: 'email', title: '邮箱' },
      { key: 'tenant', title: 'Tenant' },
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
        { name: 'path', label: 'Path' },
        { name: 'assume_role_policy', label: 'Assume Role Policy', type: 'textarea' }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        name: String(values.name ?? ''),
        ...(values.path ? { path: String(values.path) } : {}),
        ...(values.assume_role_policy ? { assume_role_policy: String(values.assume_role_policy) } : {})
      })
    },
    columns: [
      { key: 'name', title: 'Role' },
      { key: 'path', title: 'Path' },
      { key: 'arn', title: 'ARN' },
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
      initialValues: { versioning: 'enabled' },
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        bucket_id: bucketId(row),
        versioning: String(values.versioning ?? 'enabled')
      })
    },
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
      { key: 'owner', title: 'Owner' },
      { key: 'status', title: '状态' },
      { key: 'usage', title: '使用量' },
      { key: 'bucket_id', title: 'Bucket ID' },
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
        { name: 'name', label: 'Realm 名称', required: true }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? '') })
    },
    columns: [
      { key: 'name', title: 'Realm' },
      { key: 'status', title: '状态' },
      { key: 'id', title: 'ID' },
      { key: 'current_period', title: 'Current Period' },
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
      { key: 'master_zone', title: 'Master Zone' },
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
      { key: 'endpoints', title: 'Endpoints' },
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
  return String(row?.bucket_id ?? row?.name ?? row?.natural_key ?? '').trim()
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
        await operationMutation.run(() => mutateResource('/rgw/period/commit', 'POST', parameters), 'RGW Period commit 执行成功')
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
