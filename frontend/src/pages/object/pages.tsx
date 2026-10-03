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
import { rgwUserDisplayNamePatch } from './rgwUserDisplayName'
import { rgwAccountTextPatch } from './rgwAccountEdit'
import { RgwBucketDetails } from './RgwBucketDetails'
import { RgwAccountDetails } from './RgwAccountDetails'
import { rgwBucketIndexCount, rgwBucketIndexText } from './rgwBucketIndex'
import { rgwBucketVersioning, rgwBucketBooleanState, rgwBucketReshardState } from './rgwBucketState'
import { rgwUserPolicyBlocked, rgwUserPolicyInput, rgwUserPolicyOptions } from './rgwUserPolicy'
import { RgwRateLimit } from './RgwRateLimit'
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
import { RgwUserDetails } from './RgwUserDetails'
import { rgwUserOperationMaskInput, rgwUserOperationMaskOptions } from './rgwUserOperationMask'
import { rgwUserAccountRootBlocked, rgwUserAccountRootInput } from './rgwUserAccountRoot'
import { rgwSubuserOptions, rgwSubuserInput, rgwSubuserPermissionOptions } from './rgwUserSubuser'
import { rgwSubuserCreateInput } from './rgwSubuserCreate'
import { rgwSwiftRotationOptions, rgwSwiftRotationInput } from './rgwSwiftKeyRotation'
import { rgwS3KeyOwnerOptions, rgwS3KeyCreateInput } from './rgwS3KeyCreate'
import { rgwS3KeyDeleteOptions, rgwS3KeyDeleteInput } from './rgwS3KeyDelete'
import { rgwS3KeyRotateInput } from './rgwS3KeyRotate'
import { RgwGeneratedCredentialInput } from './RgwGeneratedCredentialInput'
import { rgwCapabilityOptions, rgwCapabilityInput } from './rgwUserCapsForm'
import { rgwUserCreateCredentials } from './rgwUserCreateCredentials'
import { rgwUserAccountMigrationBlocked, rgwUserAccountMigrationInput } from './rgwUserAccountMigration'
import { loadRgwMigrationAccountOptions } from './rgwMigrationAccountOptions'
import { loadRgwCreateAccountOptions, rgwUserCreateAccountInput } from './rgwUserCreateAccount'
import { rgwUserPlacementInput, rgwUserPlacementTagsInput } from './rgwUserPlacementForm'

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
    detailContent: (row, clusterId) => <RgwUserDetails row={row} clusterId={clusterId} />,
    requiredCapabilities: ['rgw_admin'],
    rowKeyCandidates: ['natural_key', 'uid', 'user_id'],
    createAction: {
      title: '新建 RGW 用户',
      buttonLabel: '新建用户',
      path: '/rgw/user',
      method: 'POST',
      successMessage: 'RGW 用户创建执行成功',
      initialValues: { account_mode: 'independent', credential_mode: 'none' },
      changedValues: (changed) => ({
        ...(Object.prototype.hasOwnProperty.call(changed, 'uid') || Object.prototype.hasOwnProperty.call(changed, 'account_mode') ? { account_id: undefined, account_root: undefined } : {}),
        ...(['uid', 'credential_mode', 'access_key', 'secret_key'].some(key => Object.prototype.hasOwnProperty.call(changed, key)) ? { credentials_saved: undefined } : {})
      }),
      confirmation: (values) => {
        const account = rgwUserCreateAccountInput(values)
        rgwUserCreateCredentials(values)
        const scope = account.account_id ? `在账户 ${JSON.stringify(account.account_id)} 中创建用户 ${JSON.stringify(values.uid)}；${account.account_root ? '授予账户根用户权限' : '普通账户用户需要策略授权才能访问资源'}，创建后不能迁出账户。` : `创建独立用户 ${JSON.stringify(values.uid)}。`
        return scope + (values.credential_mode === 's3' ? '同时创建已安全保存的 S3 凭据，提交后不提供密钥回显。' : '不创建访问密钥；之后可通过“创建 S3 访问密钥”配置访问凭据。')
      },
      fields: [
        { name: 'uid', label: 'UID', required: true },
        { name: 'account_mode', label: '用户归属', type: 'select', required: true, options: [{ label: '独立用户', value: 'independent' }, { label: '账户用户', value: 'account' }] },
        { name: 'account_id', label: '账户（按 UID 租户筛选）', type: 'select', required: true, visibleWhen: values => values.account_mode === 'account', optionsDependencies: ['uid', 'account_mode'], optionsLoader: (clusterId, _row, values) => loadRgwCreateAccountOptions(clusterId, values) },
        { name: 'account_root', label: '账户根用户权限', type: 'select', required: true, visibleWhen: values => values.account_mode === 'account', options: [{ label: '普通账户用户', value: 'disable' }, { label: '账户根用户', value: 'enable' }] },
        { name: 'display_name', label: '显示名', required:true },
        { name: 'email', label: '邮箱' },
        { name: 'credential_mode', label: '首次访问凭据', type: 'select', required: true, options: [{ label: '暂不创建密钥', value: 'none' }, { label: '创建已保存的 S3 凭据', value: 's3' }] },
        { name: 'access_key', label: 'Access Key', type: 'password', required: true, visibleWhen: values => values.credential_mode === 's3', renderControl: () => <RgwGeneratedCredentialInput kind="access" /> },
        { name: 'secret_key', label: 'Secret Key', type: 'password', required: true, visibleWhen: values => values.credential_mode === 's3', renderControl: () => <RgwGeneratedCredentialInput kind="secret" /> },
        { name: 'credentials_saved', label: '凭据已安全保存', type: 'select', required: true, visibleWhen: values => values.credential_mode === 's3', options: [{ label: '已安全保存 Access Key 和 Secret Key', value: 'saved' }] },
        { name:'max_buckets',label:'最大 Bucket 数（-1 禁止创建，0 无限制）',type:'number',min:-1,max:2147483647 }
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        uid: String(values.uid ?? ''),
        ...(values.display_name ? { display_name: String(values.display_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {}),
        ...rgwBucketLimitInput(values.max_buckets),
        ...rgwUserCreateCredentials(values),
        ...rgwUserCreateAccountInput(values)
      })
    },
    updateAction: {
      title: '更新 RGW 用户',
      path: '/rgw/user',
      method: 'PATCH',
      successMessage: 'RGW 用户更新执行成功',
      fields: [
        { name: 'display_name', label: '显示名（留空不修改）' },
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
      buildBody: (values, clusterId, row) => {
        const patch = {
          ...rgwUserDisplayNamePatch(values.display_name, row?.display_name),
          ...rgwUserEmailPatch(values),
          ...rgwBucketLimitPatch(values.max_buckets, row?.max_buckets),
          ...rgwUserFlagPatch(values)
        }
        if (Object.keys(patch).length === 0) throw new Error('没有需要提交的用户修改')
        return { cluster_id: clusterId, uid: userId(row), ...patch }
      }
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
      { title: '迁入账户（不可逆）', path: '/rgw/user', method: 'PATCH', successMessage: '迁入命令完成且用户账户归属已确认，请检查 Bucket 归属及访问策略',
        disabledWhen: rgwUserAccountMigrationBlocked,
        fields: [
          { name: 'target_account_id', label: '目标账户（同租户库存；无选项时请先采集账户）', type: 'select', required: true, optionsLoader: loadRgwMigrationAccountOptions },
          { name: 'migration_confirm_uid', label: '输入完整用户 UID，确认转移用户及其 Bucket 归属，且不能迁出账户', required: true }
        ],
        confirmation: (values, row) => {
          const input = rgwUserAccountMigrationInput(values, row)
          return `不可逆：将用户 ${JSON.stringify(userId(row))} 及其 Bucket 迁入账户 ${JSON.stringify(input.target_account_id)}。访问权限与配额语义将改变，不会自动授予根用户或托管策略。失败可能留下部分 Bucket 已迁移，必须人工检查，不能直接重试。确认继续？`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwUserAccountMigrationInput(values, row) })
      },
      { title: '设置账户根用户', path: '/rgw/user', method: 'PATCH', successMessage: '账户根用户状态设置执行成功',
        disabledWhen: rgwUserAccountRootBlocked,
        fields: [{ name: 'account_root', label: '账户根用户状态（显著改变账户访问权限）', type: 'select', required: true, options: [{ label: '设为账户根用户', value: 'enable' }, { label: '改为普通 RGW 用户', value: 'disable' }] }],
        confirmation: (values, row) => {
          const input = rgwUserAccountRootInput(values, row)
          return `将用户 ${JSON.stringify(userId(row))} 在账户 ${JSON.stringify(input.expected_account_id)} 中${input.account_root ? '提升为根用户，授予账户级根用户权限' : '降为普通 RGW 用户，访问将依赖其策略授权，可能失去现有访问权限'}。此操作不会迁移账户或修改关联策略。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwUserAccountRootInput(values, row) })
      },
      { title: '设置用户操作掩码', path: '/rgw/user', method: 'PATCH', successMessage: '操作掩码设置执行成功',
        fields: [{ name: 'op_mask', label: '允许的操作类别（整体替换，非完整有效权限）', type: 'select', required: true, options: rgwUserOperationMaskOptions }],
        confirmation: (values, row) => `将用户 ${JSON.stringify(userId(row))} 的操作掩码整体替换为 ${JSON.stringify(rgwUserOperationMaskInput(values).op_mask)}，可能限制现有访问。ACL 和策略仍独立生效。`,
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwUserOperationMaskInput(values) })
      },
      { title: '设置用户默认放置', path: '/rgw/user', method: 'PATCH', successMessage: '默认放置设置执行成功',
        fields: [
          { name: 'default_placement', label: '默认放置规则（必须存在，不支持清空）', required: true },
          { name: 'default_storage_class', label: '默认存储类（留空使用原生默认类）' }
        ],
        confirmation: (values, row) => {
          const placement = rgwUserPlacementInput(values)
          return `将替换用户 ${JSON.stringify(userId(row))} 的默认放置规则为 ${JSON.stringify(placement.default_placement)}，存储类为 ${placement.default_storage_class === '' ? '原生默认类（空值）' : JSON.stringify(placement.default_storage_class)}。不迁移已有 Bucket 或对象。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwUserPlacementInput(values) })
      },
      { title: '替换用户放置标签', path: '/rgw/user', method: 'PATCH', successMessage: '放置标签替换执行成功',
        fields: [{ name: 'placement_tags_csv', label: '完整标签列表（逗号分隔，不支持清空；空格属于标签）', required: true }],
        confirmation: (values, row) => `将整体替换用户 ${JSON.stringify(userId(row))} 的放置标签，可能改变可使用的放置目标；这不是追加操作。新列表：${JSON.stringify(rgwUserPlacementTagsInput(values).placement_tags_csv)}`,
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwUserPlacementTagsInput(values) })
      },
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
      { title: '创建 S3 访问密钥', path: '/rgw/user/key', method: 'POST', successMessage: 'S3 密钥及归属已回读验证；请使用预先保存的凭据',
        changedValues: changed => Object.keys(changed).some(key => ['owner', 'access_key', 'secret_key'].includes(key)) ? { credentials_saved: undefined, confirm_owner: undefined } : {},
        fields: [
          { name: 'owner', label: '凭据所属用户', type: 'select', required: true, optionsLoader: async (_clusterId, row) => rgwS3KeyOwnerOptions(row) },
          { name: 'access_key', label: '尚未使用的 Access Key（预先保存）', type: 'password', required: true, renderControl: () => <RgwGeneratedCredentialInput kind="access" /> },
          { name: 'secret_key', label: '高强度 Secret Key（预先保存，提交后不回显）', type: 'password', required: true, renderControl: () => <RgwGeneratedCredentialInput kind="secret" /> },
          { name: 'credentials_saved', label: '凭据保存确认', type: 'select', required: true, options: [{ label: '已安全保存本次凭据', value: 'saved' }] },
          { name: 'confirm_owner', label: '输入完整凭据所属用户 ID 确认', required: true }
        ],
        confirmation: (values, row) => `为 ${JSON.stringify(rgwS3KeyCreateInput(values, row).confirm_owner)} 创建已激活的 S3 访问密钥？该凭据将具有目标用户的现有权限；本系统不提供密钥查询，请确认已安全保存。`,
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwS3KeyCreateInput(values, row) })
      },
      { title: '轮换 S3 访问密钥', path: '/rgw/user/key', method: 'PATCH', successMessage: 'S3 Secret Key 已轮换并核验原激活状态；请更新客户端凭据',
        disabledWhen: row => rgwS3KeyDeleteOptions(row).length ? undefined : '没有已采集的 S3 密钥，请先刷新库存',
        changedValues: changed => Object.keys(changed).some(key => ['owner', 'access_key', 'secret_key'].includes(key)) ? { credentials_saved: undefined, confirm_owner: undefined } : {},
        fields: [
          { name: 'owner', label: '已有密钥所属用户', type: 'select', required: true, optionsLoader: async (_clusterId, row) => rgwS3KeyDeleteOptions(row) },
          { name: 'access_key', label: '原 Access Key（保持不变，请从安全保存位置提供）', type: 'password', required: true },
          { name: 'secret_key', label: '不同于旧值的新 Secret Key（预先保存，提交后不回显）', type: 'password', required: true, renderControl: () => <RgwGeneratedCredentialInput kind="secret" /> },
          { name: 'credentials_saved', label: '确认保存新凭据并准备更新客户端', type: 'select', required: true, options: [{ label: '已安全保存并准备切换客户端', value: 'saved' }] },
          { name: 'confirm_owner', label: '输入完整凭据所属用户 ID 确认', required: true }
        ],
        confirmation: (values, row) => `轮换 ${JSON.stringify(rgwS3KeyRotateInput(values, row).confirm_owner)} 的指定 S3 Secret Key？旧 Secret Key 将失效，客户端需更新；Access Key、权限、密钥归属及原激活状态保持不变，新密钥不提供后续查询。`,
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwS3KeyRotateInput(values, row) })
      },
      { title: '删除 S3 访问密钥', path: '/rgw/user/key', method: 'DELETE', successMessage: '指定 S3 密钥已删除并回读确认；请检查受影响的客户端',
        disabledWhen: row => rgwS3KeyDeleteOptions(row).length ? undefined : '没有已采集的 S3 密钥，请先刷新库存',
        changedValues: changed => Object.keys(changed).some(key => ['owner', 'access_key'].includes(key)) ? { confirm_owner: undefined } : {},
        fields: [
          { name: 'owner', label: '待删除密钥所属用户', type: 'select', required: true, optionsLoader: async (_clusterId, row) => rgwS3KeyDeleteOptions(row) },
          { name: 'access_key', label: '待删除的原 Access Key（库存已脱敏，请从安全保存位置提供）', type: 'password', required: true },
          { name: 'confirm_owner', label: '输入完整凭据所属用户 ID 确认', required: true }
        ],
        confirmation: (values, row) => `删除 ${JSON.stringify(rgwS3KeyDeleteInput(values, row).confirm_owner)} 的指定 S3 Access Key？使用该凭据的客户端将无法继续访问，操作不可恢复；不会删除用户、子用户或其他密钥。`,
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwS3KeyDeleteInput(values, row) })
      },
      { title: '创建子用户', path: '/rgw/user/subuser', method: 'POST', successMessage: '子用户、权限及凭据已创建并回读验证；密钥不会显示在库存中',
        changedValues: (changed) => Object.keys(changed).some(key => ['subuser', 'key_type', 'access_key', 'secret_key'].includes(key)) ? { credentials_saved: undefined, confirm_subuser: undefined } : {},
        fields: [
          { name: 'subuser', label: '本地子用户名（不含 UID 或冒号）', required: true },
          { name: 'subuser_permission', label: '子用户权限', type: 'select', required: true, options: rgwSubuserPermissionOptions },
          { name: 'key_type', label: '密钥类型', type: 'select', required: true, options: [{ label: 'S3', value: 's3' }, { label: 'Swift', value: 'swift' }] },
          { name: 'access_key', label: '预先保存的 S3 Access Key', type: 'password', required: true, visibleWhen: values => values.key_type === 's3', renderControl: () => <RgwGeneratedCredentialInput kind="access" /> },
          { name: 'secret_key', label: '预先保存的高强度 Secret Key（提交后不回显）', type: 'password', required: true, renderControl: () => <RgwGeneratedCredentialInput kind="secret" /> },
          { name: 'credentials_saved', label: '凭据保存确认', type: 'select', required: true, options: [{ label: '已将本次凭据保存在安全位置', value: 'saved' }] },
          { name: 'confirm_subuser', label: '输入完整子用户 ID 确认（UID:子用户名）', required: true }
        ],
        confirmation: (values, row) => {
          const input = rgwSubuserCreateInput(values, row)
          return `为 ${JSON.stringify(input.confirm_subuser)} 创建 ${input.key_type} 凭据，授予 ${JSON.stringify(input.subuser_permission)} 权限？请确认已安全保存密钥，本系统不会提供密钥查询。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwSubuserCreateInput(values, row) })
      },
      { title: '轮换 Swift 子用户密钥', path: '/rgw/user/subuser', method: 'POST', successMessage: 'Swift 密钥已轮换并核验原激活状态；请更新使用旧凭据的客户端',
        disabledWhen: row => rgwSwiftRotationOptions(row).length ? undefined : '没有状态明确的 Swift 子用户密钥，请先刷新库存',
        changedValues: changed => Object.keys(changed).some(key => ['subuser', 'secret_key'].includes(key)) ? { credentials_saved: undefined, confirm_subuser: undefined } : {},
        fields: [
          { name: 'subuser', label: 'Swift 子用户密钥', type: 'select', required: true, optionsLoader: async (_clusterId, row) => rgwSwiftRotationOptions(row) },
          { name: 'secret_key', label: '不同于旧值的新 Secret Key（预先保存，提交后不回显）', type: 'password', required: true, renderControl: () => <RgwGeneratedCredentialInput kind="secret" /> },
          { name: 'credentials_saved', label: '确认保存新凭据并准备更新客户端', type: 'select', required: true, options: [{ label: '已安全保存并准备切换客户端', value: 'saved' }] },
          { name: 'confirm_subuser', label: '输入完整子用户 ID 确认（UID:子用户名）', required: true }
        ],
        confirmation: (values, row) => {
          const input = rgwSwiftRotationInput(values, row)
          return `轮换 ${JSON.stringify(input.confirm_subuser)} 的 Swift 密钥？旧凭据将失效，客户端需更新；保留${input.expected_key_active ? '已激活' : '已停用'}状态，不修改权限或 S3 密钥。新密钥不会提供后续查询。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwSwiftRotationInput(values, row) })
      },
      { title: '管理已有子用户', path: '/rgw/user/subuser', method: 'POST', successMessage: '子用户变更已执行并回读验证',
        disabledWhen: (row) => rgwSubuserOptions(row).length ? undefined : '没有可操作的子用户，请先刷新用户库存',
        changedValues: (changed) => Object.prototype.hasOwnProperty.call(changed, 'subuser') || Object.prototype.hasOwnProperty.call(changed, 'action') ? { confirm_subuser: undefined, subuser_permission: undefined } : {},
        fields: [
          { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '修改权限', value: 'modify' }, { label: '删除子用户及其全部关联密钥', value: 'rm' }] },
          { name: 'subuser', label: '已有子用户', type: 'select', required: true, optionsLoader: async (_clusterId, row) => rgwSubuserOptions(row) },
          { name: 'subuser_permission', label: '子用户权限（整体替换）', type: 'select', required: true, visibleWhen: values => values.action === 'modify', options: rgwSubuserPermissionOptions },
          { name: 'confirm_subuser', label: '输入完整子用户 ID 确认（包含 UID 和冒号）', required: true }
        ],
        confirmation: (values, row) => {
          const input = rgwSubuserInput(values, row)
          return input.action === 'rm' ? `删除子用户 ${JSON.stringify(input.confirm_subuser)} 及其全部 S3、Swift 关联密钥？现有客户端将无法继续使用这些凭据，此操作不可恢复。` : `将子用户 ${JSON.stringify(input.confirm_subuser)} 的权限整体替换为 ${JSON.stringify(input.subuser_permission)}？这会改变访问权限，不会生成或轮换密钥。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwSubuserInput(values, row) })
      },
      { title: '管理用户权限（caps）', path: '/rgw/user/caps', method: 'POST', successMessage: '用户管理权限操作执行成功',
        initialValues: { action: 'add', permission: 'read' },
        fields: [
          { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '合并添加权限', value: 'add' }, { label: '移除指定权限', value: 'rm' }, { label: '替换已有类型的权限', value: 'replace' }] },
          { name: 'type', label: '权限类型', type: 'select', required: true, optionsDependencies: ['action'], optionsLoader: async (_clusterId, row, values) => rgwCapabilityOptions(row, values?.action) },
          { name: 'permission', label: '权限', type: 'select', required: true, options: [{ label: '读', value: 'read' }, { label: '写', value: 'write' }, { label: '读写', value: 'read,write' }, { label: '全部', value: '*' }] }
        ],
        confirmation: (values, row) => {
          const input = rgwCapabilityInput(values, row)
          const scope = `用户 ${JSON.stringify(userId(row))} 的 ${JSON.stringify(input.type)} 管理权限 ${JSON.stringify(input.permission)}`
          return input.action === 'replace' ? `整体替换${scope}？其它类型权限保留。这会先移除该类型权限，再添加新权限，并非原子操作；失败时权限可能已被移除，请检查实际状态后处理，不要盲目重试。RGW 管理权限可能允许访问其他用户的数据。` : `${input.action === 'add' ? '合并添加' : '仅移除'}${scope}？这不是整体替换：其它权限保留。RGW 管理权限可能允许访问其他用户的数据，请确认授权范围。`
        },
        buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, uid: userId(row), ...rgwCapabilityInput(values, row) })
      }],
    deleteAction: {
      title: '删除 RGW 用户',
      path: '/rgw/user',
      action: 'rgw_user.delete',
      resourceKind: 'rgw_user',
      successMessage: 'RGW 用户删除执行成功',
      confirmation: (row) => `删除 RGW 用户 ${JSON.stringify(userId(row))}？该用户的凭据将失效，操作不可恢复。本操作不清理 Bucket 或对象数据；若用户仍拥有 Bucket，Ceph 将拒绝删除，请先处理其归属。`,
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, uid: userId(row) }),
      resourceKey: (row) => `rgw/user/${userId(row)}`
    },
    columns: [
      { key: 'uid', title: 'UID' },
      { key: 'display_name', title: '显示名' },
      { key: 'email', title: '邮箱' },
      { key: 'max_buckets', title: '最大 Bucket 数', render: rgwBucketLimit },
      { key: 'suspended', title: '用户暂停状态', render: rgwUserSuspension },
      { key: 'system', title: '系统用户', render: rgwUserBooleanFlag },
      { key: 'admin', title: '管理员标志', render: rgwUserBooleanFlag },
      { key: 'status', title: '状态' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwAccounts: {
    title: 'RGW Accounts',
    path: '/rgw/accounts',
    detailContent: (row) => <RgwAccountDetails row={row} />,
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
        { name: 'tenant', label: 'Tenant' },
        ...['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((name, index) => ({ name, label: ['用户上限', '角色上限', '用户组上限', 'Bucket 上限', '每用户访问密钥上限'][index] + (name === 'max_buckets' ? '（-1 禁止创建，0 无限制；留空使用默认值）' : '（-1 无限制，0 禁止新增；留空使用默认值）'), type: 'number' as const, min: -1, max: 2147483647 }))
      ],
      buildBody: (values, clusterId) => ({
        cluster_id: clusterId,
        account_id: String(values.account_id ?? ''),
        ...(values.account_name ? { account_name: String(values.account_name) } : {}),
        ...(values.email ? { email: String(values.email) } : {}),
        ...(values.tenant ? { tenant: String(values.tenant) } : {}),
        ...rgwAccountLimitPatch(values)
      })
    },
    updateAction: {
      title: '编辑 RGW Account', path: '/rgw/account', method: 'PATCH', successMessage: '账户更新执行成功',
      fields: [
        { name: 'account_name', label: '账户名称（不支持清空）' },
        { name: 'email', label: '邮箱（不支持清空）' },
        ...['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((name, index) => ({ name, label: ['用户上限', '角色上限', '用户组上限', 'Bucket 上限', '每用户访问密钥上限'][index] + (name === 'max_buckets' ? '（-1 禁止创建，0 无限制）' : '（-1 无限制，0 禁止新增）'), type: 'number' as const, min: -1, max: 2147483647 }))
      ],
      initialValues: (row) => ({ account_name: text(row?.account_name), email: text(row?.email), ...Object.fromEntries(['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys'].map((key) => [key, numberOrUndefined(row?.[key])])) }),
      buildBody: (values, clusterId, row) => {
        const patch = { ...rgwAccountTextPatch(values, row), ...rgwAccountLimitPatch(values, row) }
        if (Object.keys(patch).length === 0) throw new Error('没有需要提交的账户修改')
        return { cluster_id: clusterId, account_id: String(row?.account_id ?? row?.natural_key ?? ''), ...patch }
      }
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
    detailContent: (row) => <RgwBucketDetails row={row} />,
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
      { key: 'index_type', title: '索引类型', render: rgwBucketIndexText },
      { key: 'num_shards', title: '索引分片数', render: rgwBucketIndexCount },
      { key: 'placement_rule', title: '放置规则' },
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
  const uid = row?.uid
  if (typeof uid !== 'string' || uid === '' || uid !== uid.trim() || uid.startsWith('-') || /[\/\0\r\n]/.test(uid)) {
    throw new Error('用户完整 UID 未返回或无效，请刷新库存后重试')
  }
  return uid
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
