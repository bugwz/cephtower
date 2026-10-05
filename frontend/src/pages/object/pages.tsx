import type { ApiRecord } from '../../api/client'
import { rgwRealmImportAction } from './rgwRealmImport'
import { rgwRealmSetupAction } from './rgwRealmSetup'
import { rgwRealmMigrationAction } from './rgwRealmMigration'
import { RgwRealmToken } from './RgwRealmToken'
import { realmDeleteBlocked, realmDeleteInput, realmDeleteConfirmation } from './rgwRealmDelete'
import { zonegroupDeleteBlocked, zonegroupDeleteInput, zonegroupDeleteConfirmation } from './rgwZonegroupDelete'
import { zoneDeleteBlocked, zoneDeleteInput, zoneDeleteConfirmation } from './rgwZoneDelete'
import { zonePlacementBlocked, zonePlacementGroups, zonePlacementOptions, zonePlacementClasses, zonePlacementChanged, zonePlacementInput, zonePlacementConfirmation, zonePlacementCompressions } from './rgwZonePlacement'
import { zoneStorageClassInput, zoneStorageClassConfirmation } from './rgwZonePlacement'
import { zonePlacementCreateBlocked, zonePlacementCreateInput, zonePlacementCreateConfirmation } from './rgwZonePlacement'
import { groupStorageClassBlocked, groupStorageClassOptions, groupStorageClassInput, groupStorageClassConfirmation } from './rgwZonegroupStorageClass'
import { groupPlacementCreateBlocked, groupPlacementCreateInput, groupPlacementCreateConfirmation } from './rgwZonegroupStorageClass'
import { groupPlacementDefaultClasses, groupPlacementDefaultInput, groupPlacementDefaultConfirmation } from './rgwZonegroupStorageClass'
import { groupPlacementTagsChanged, groupPlacementTagsInput, groupPlacementTagsConfirmation } from './rgwZonegroupStorageClass'
import { groupStorageClassDeleteInput, groupStorageClassDeleteConfirmation } from './rgwZonegroupStorageClass'
import { groupLocalClassDeleteZones, groupLocalClassDeleteClasses, groupLocalClassDeleteInput, groupLocalClassDeleteConfirmation } from './rgwZonegroupStorageClass'
import { RgwTopologyView } from './RgwTopology'
import { RgwSyncStatus } from './RgwSyncStatus'
import { RgwZonePoolReferences } from './RgwZonePoolReferences'
import { RgwPlacementClasses } from './RgwPlacementClasses'
import { RgwLocalClassDetails } from './RgwLocalClassDetails'
import { cloudRestoreBlocked, cloudRestoreTargets, cloudRestoreClasses, cloudRestoreLocalClasses, cloudRestoreChanged, cloudRestoreInput, cloudRestoreConfirmation } from './rgwCloudRestore'
import { RgwRealmTransfer } from './RgwRealmTransfer'
import { bucketSyncPipeZonesSelectionChanged } from './rgwBucketSyncGroupForm'
import { zonegroupPipeZonesSelectionChanged } from './rgwZonegroupSyncGroup'
import { bucketSyncPipeGroupOptions, bucketSyncPipeOptions, bucketSyncPipeSelectionChanged } from './rgwBucketSyncGroupForm'
import { zonegroupPipeGroupOptions, zonegroupPipeOptions, zonegroupPipeSelectionChanged } from './rgwZonegroupSyncGroup'
import { RgwTopicDetails, topicText, topicBoolean, topicEndpoint } from './RgwTopicDetails'
import { topicDeleteBlocked, topicDeleteInput, topicDeleteConfirmation } from './rgwTopicDelete'
import { topicPolicyBlocked, topicPolicyInitial, topicPolicyInput, topicPolicyConfirmation } from './rgwTopicPolicy'
import { topicAttributeOptions, topicAttributeBlocked, topicAttributeInitial, topicAttributeInput, topicAttributeConfirmation } from './rgwTopicAttribute'
import { topicEndpointBlocked, topicEndpointInitial, topicEndpointInput, topicEndpointConfirmation } from './rgwTopicEndpoint'
import { topicWritableOptions, topicOptionBlocked, topicOptionInitial, topicOptionInput, topicOptionConfirmation } from './rgwTopicOption'
import { topicCreateInput, topicCreateConfirmation } from './rgwTopicCreate'
import { RgwTopicOptionsEditor } from './RgwTopicOptionsEditor'
import { roleManagedPolicyBlocked, roleManagedPolicyInitial, roleManagedPolicyInput, roleManagedPolicyConfirmation, roleManagedPolicyOptions } from './rgwRoleManagedPolicy'
import { bucketReplicationFormBlocked, bucketReplicationFormInitial, bucketReplicationFormInput, bucketReplicationFormConfirmation } from './rgwBucketReplicationForm'
import { useClusterContext } from '../../state/ClusterContext'
import { periodCommitInitial, periodCommitInput, periodCommitConfirmation, periodCommitBlocked } from './rgwPeriodCommit'
import { RgwCurrentPeriod } from './RgwCurrentPeriod'
import { zonegroupSyncInitial, zonegroupSyncBlocked, zonegroupSyncInput, zonegroupSyncConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupSyncCreateInitial, zonegroupSyncCreateBlocked, zonegroupSyncCreateInput, zonegroupSyncCreateConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupSyncDeleteInput, zonegroupSyncDeleteConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupFlowCreateInput, zonegroupFlowCreateConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupFlowDeleteInput, zonegroupFlowDeleteConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupFlowUpdateInput, zonegroupFlowUpdateConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupPipeCreateInput, zonegroupPipeCreateConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupPipeDeleteInput, zonegroupPipeDeleteConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupPipeUpdateInput, zonegroupPipeUpdateConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupPipeZonesInput, zonegroupPipeZonesConfirmation } from './rgwZonegroupSyncGroup'
import { zonegroupReplicationPrepareBlocked, zonegroupReplicationPrepareInitial, zonegroupReplicationPrepareInput, zonegroupReplicationPrepareConfirmation } from './rgwZonegroupSyncGroup'
import { ExternalListPage, type ExternalListPageDefinition } from '../ExternalListPage'
import { ResourceListPage, type ResourceListPageDefinition, type ResourceFormAction } from '../ResourceListPage'
import { ServiceDaemons } from '../cluster/ServiceDaemons'
import { RgwQuota } from './RgwQuota'
import { RgwStorage } from './RgwStorage'
import { rgwUserDisplayNamePatch } from './rgwUserDisplayName'
import { rgwAccountTextPatch } from './rgwAccountEdit'
import { RgwBucketDetails } from './RgwBucketDetails'
import { RgwBucketTagEntries } from './RgwBucketTagEntries'
import { RgwBucketTagEditor } from './RgwBucketTagEditor'
import { bucketTagFormBlocked, bucketTagFormInitial, bucketTagFormInput, bucketTagFormConfirmation } from './rgwBucketTagForm'
import { RgwAccountDetails } from './RgwAccountDetails'
import { rgwBucketIndexCount, rgwBucketIndexText } from './rgwBucketIndex'
import { rgwBucketVersioning, rgwBucketBooleanState, rgwBucketReshardState } from './rgwBucketState'
import { rgwUserPolicyBlocked, rgwUserPolicyInput, rgwUserPolicyOptions, rgwUserPolicyAttachOptions } from './rgwUserPolicy'
import { RgwRateLimit } from './RgwRateLimit'
import { rgwBucketLimit, rgwBucketLimitInput, rgwBucketLimitPatch } from './rgwBucketLimit'
import { rgwAccountLimit, rgwAccountLimitPatch } from './rgwAccountLimit'
import { rgwUserSuspension, rgwUserBooleanFlag } from './rgwUserFlags'
import { rgwUserFlagPatch, rgwUserUpdateConfirmation } from './rgwUserFlagPatch'
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
import { rgwUserCreateFlags } from './rgwUserCreateFlags'
import { rgwBucketConfigurationOptions, rgwBucketConfigurationInput, rgwBucketConfigurationDeleteBlocked, rgwBucketConfigurationDeleteInput, rgwBucketConfigurationDeleteConfirmation, rgwBucketConfigurationEditBlocked, rgwBucketConfigurationEditInitial, rgwBucketConfigurationEditInput, rgwBucketConfigurationUpdateConfirmation } from './rgwBucketConfiguration'
import { rgwBucketEncryptionSummary } from './rgwBucketEncryptionSummary'
import { RgwBucketCorsRules } from './RgwBucketCorsRules'
import { RgwBucketLifecycleRules } from './RgwBucketLifecycleRules'
import { rgwLifecycleProgress } from './rgwLifecycleProgress'
import { rgwBucketSyncPolicy } from './rgwBucketSyncPolicy'
import { RgwBucketSyncFlows } from './RgwBucketSyncFlows'
import { RgwBucketSyncPipes } from './RgwBucketSyncPipes'
import { bucketSyncGroupBlocked, bucketSyncGroupInitial, bucketSyncGroupInput, bucketSyncGroupConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncGroupCreateBlocked, bucketSyncGroupCreateInitial, bucketSyncGroupCreateInput, bucketSyncGroupCreateConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncGroupDeleteInput, bucketSyncGroupDeleteConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncFlowInput, bucketSyncFlowConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncFlowUpdateInput, bucketSyncFlowUpdateConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncPipeUpdateInput, bucketSyncPipeUpdateConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncPipeZonesInput, bucketSyncPipeZonesConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncFlowDeleteInput, bucketSyncFlowDeleteConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncPipeDeleteInput, bucketSyncPipeDeleteConfirmation } from './rgwBucketSyncGroupForm'
import { bucketSyncPipeCreateInput, bucketSyncPipeCreateConfirmation } from './rgwBucketSyncGroupForm'
import { rgwBucketConfigurationReadOptions } from './rgwBucketConfiguration'
import { rgwBucketObjectLockSummary } from './rgwBucketObjectLockSummary'
import { RgwBucketAcl } from './RgwBucketAcl'
import { RgwBucketReplication } from './RgwBucketReplication'
import { RgwBucketNotifications } from './RgwBucketNotifications'
import { bucketVersioningSummary, bucketMFABlocked, bucketMFAInitial, bucketMFAInput, bucketMFAConfirmation } from './rgwBucketMFA'
import { RgwBucketNotificationEditor } from './RgwBucketNotificationEditor'
import { notificationFormBlocked, notificationFormInitial, notificationFormInput, notificationFormConfirmation } from './rgwBucketNotificationForm'
import { bucketNotificationDeleteBlocked, bucketNotificationDeleteInitial, bucketNotificationDeleteInput, bucketNotificationDeleteConfirmation } from './rgwBucketNotificationDelete'
import { bucketAclOptions, bucketAclFormBlocked, bucketAclFormInitial, bucketAclFormInput, bucketAclFormConfirmation } from './rgwBucketAclForm'
import { objectLockFormInitial, objectLockFormBlocked, objectLockFormInput, objectLockFormConfirmation } from './rgwBucketObjectLockForm'
import { RgwBucketLifecycleEditor } from './RgwBucketLifecycleEditor'
import { lifecycleFormInitial, lifecycleFormBlocked, lifecycleFormInput, lifecycleFormConfirmation } from './rgwBucketLifecycleForm'
import { RgwBucketCorsEditor } from './RgwBucketCorsEditor'
import { corsFormInitial, corsFormBlocked, corsFormInput, corsFormConfirmation } from './rgwBucketCorsForm'
import { bucketDeleteInput, bucketDeleteBlocked, bucketDeleteConfirmation } from './rgwBucketDelete'
import { bucketVersioningInitial, bucketVersioningInput, bucketVersioningConfirmation } from './rgwBucketVersioningForm'
import { bucketEncryptionFormInitial, bucketEncryptionFormBlocked, bucketEncryptionFormInput, bucketEncryptionFormConfirmation } from './rgwBucketEncryptionForm'
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

export function RgwTopicsPage() {
  return <ResourceListPage definition={definitions.rgwTopics} />
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
  const { selectedClusterId } = useClusterContext()
  return <><RgwTopologyView key={selectedClusterId ?? 'none'} clusterId={selectedClusterId} /><ResourceListPage definition={definitions.multisite} /></>
}

export function RgwZonegroupsPage() {
  return <ResourceListPage definition={definitions.rgwZonegroups} />
}

export function RgwZonesPage() {
  return <ResourceListPage definition={definitions.rgwZones} />
}

export function RgwPeriodPage() {
  return <ResourceListPage definition={{ ...definitions.multisite, title: 'RGW Period（按 Realm 提交）', createAction: undefined, toolbarActions: undefined, updateAction: undefined, extraActions: [{
    title: '提交 Realm Period', path: '/rgw/period/commit', method: 'POST',
    successMessage: 'Realm 当前 period 已回读核验（不代表远端同步完成）',
    disabledWhen: periodCommitBlocked, initialValues: periodCommitInitial, confirmation: periodCommitConfirmation,
    fields: [
      { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
      { name: 'expected_current_period', label: '采集时当前 Period（不可更改）', readOnly: true },
      { name: 'confirm_commit', label: 'Realm 范围发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已核对全部待发布多站点变更，了解非事务及部分生效风险' }] }
    ],
    buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...periodCommitInput(values, row) })
  }] }} />
}

export function ObjectStorageConfigPage() {
  return <ResourceListPage definition={definitions.objectStorageConfig} />
}

const definitions: Record<
  | 'rgwOverview'
  | 'rgwUsers'
  | 'rgwAccounts'
  | 'rgwRoles'
  | 'rgwTopics'
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
      initialValues: { account_mode: 'independent', credential_mode: 'none', system: 'disable', suspended: 'disable' },
      changedValues: (changed) => ({
        ...(Object.prototype.hasOwnProperty.call(changed, 'uid') || Object.prototype.hasOwnProperty.call(changed, 'account_mode') ? { account_id: undefined, account_root: undefined } : {}),
        ...(['uid', 'credential_mode', 'access_key', 'secret_key'].some(key => Object.prototype.hasOwnProperty.call(changed, key)) ? { credentials_saved: undefined } : {})
      }),
      confirmation: (values) => {
        const account = rgwUserCreateAccountInput(values)
        rgwUserCreateCredentials(values)
        const flags = rgwUserCreateFlags(values)
        const scope = account.account_id ? `在账户 ${JSON.stringify(account.account_id)} 中创建用户 ${JSON.stringify(values.uid)}；${account.account_root ? '授予账户根用户权限' : '普通账户用户需要策略授权才能访问资源'}，创建后不能迁出账户。` : `创建独立用户 ${JSON.stringify(values.uid)}。`
        return scope + (flags.system ? '启用系统用户标志，授予 RGW 内部系统操作能力，请确认确有此需求。' : '不启用系统用户标志。') + (flags.suspended ? '创建后另行暂停，并非原子操作；期间用户尚未暂停，第二步失败时用户可能已存在且仍启用，请检查状态，不要重复创建。' : '用户创建后保持启用。') + (values.credential_mode === 's3' ? '同时创建已安全保存的 S3 凭据，提交后不提供密钥回显。' : '不创建访问密钥；之后可通过“创建 S3 访问密钥”配置访问凭据。')
      },
      fields: [
        { name: 'uid', label: 'UID', required: true },
        { name: 'account_mode', label: '用户归属', type: 'select', required: true, options: [{ label: '独立用户', value: 'independent' }, { label: '账户用户', value: 'account' }] },
        { name: 'account_id', label: '账户（按 UID 租户筛选）', type: 'select', required: true, visibleWhen: values => values.account_mode === 'account', optionsDependencies: ['uid', 'account_mode'], optionsLoader: (clusterId, _row, values) => loadRgwCreateAccountOptions(clusterId, values) },
        { name: 'account_root', label: '账户根用户权限', type: 'select', required: true, visibleWhen: values => values.account_mode === 'account', options: [{ label: '普通账户用户', value: 'disable' }, { label: '账户根用户', value: 'enable' }] },
        { name: 'display_name', label: '显示名', required:true },
        { name: 'email', label: '邮箱' },
        { name: 'system', label: '系统用户标志（高权限）', type: 'select', required: true, options: [{ label: '关闭', value: 'disable' }, { label: '启用系统用户', value: 'enable' }] },
        { name: 'suspended', label: '创建后状态', type: 'select', required: true, options: [{ label: '保持启用', value: 'disable' }, { label: '创建后暂停（非原子操作）', value: 'enable' }] },
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
        ...rgwUserCreateFlags(values),
        ...rgwUserCreateAccountInput(values)
      })
    },
    updateAction: {
      title: '更新 RGW 用户',
      path: '/rgw/user',
      method: 'PATCH',
      successMessage: 'RGW 用户更新执行成功',
      confirmation: (values, row) => rgwUserUpdateConfirmation(values, userId(row)),
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
        initialValues: { action: 'attach', policy_source: 'reference' },
        fields: [
          { name: 'action', label: '操作', type: 'select', required: true, options: [{ label: '关联策略', value: 'attach' }, { label: '解除关联', value: 'detach' }] },
          { name: 'policy_source', label: '策略来源（参考列表不代表目标版本支持）', type: 'select', required: true, visibleWhen: values => values.action === 'attach', options: [{label:'参考 Dashboard S3 策略',value:'reference'},{label:'手填目标版本支持的 ARN',value:'custom'}] },
          { name: 'reference_policy', label: '未关联的参考策略', type: 'select', required: true, visibleWhen: values => values.action === 'attach' && values.policy_source === 'reference', optionsDependencies: ['action','policy_source'], optionsLoader: async (_clusterId,row) => rgwUserPolicyAttachOptions(row) },
          { name: 'policy_arn', label: '托管策略 ARN（由 Ceph 验证是否支持）', required: true, visibleWhen: (values) => values.action === 'attach' && values.policy_source === 'custom' },
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
  rgwTopics: {
    title: 'RGW 通知目标（Topics）', path: '/rgw/topics', requiredCapabilities: ['rgw_admin'],
    rowKeyCandidates: ['natural_key'],
    createAction: {
      title:'新建 RGW Topic',buttonLabel:'新建 Topic',path:'/rgw/topic',method:'POST',successMessage:'Topic 已创建并回读核验（桶通知规则需单独配置）',confirmation:topicCreateConfirmation,
      initialValues:{scope:'',time_to_live:'None',max_retries:'None',retry_sleep_duration:'None',options:'{}',policy:'',opaque_data:''},
      fields:[
        {name:'name',label:'Topic 名称',required:true},{name:'owner_uid',label:'已配置 S3 永久密钥所属完整 UID（后端核验）',required:true},
        {name:'scope',label:'租户 / Account ID（全局租户留空）'},{name:'zonegroup',label:'RGW 端点所属 Zonegroup 名称',required:true},
        {name:'endpoint_mode',label:'推送端点',type:'select',required:true,options:[{value:'none',label:'无推送端点'},{value:'url',label:'指定完整 URL'},{value:'fields',label:'按协议、主机和凭据分项输入'}]},
        {name:'endpoint_secret',label:'完整推送 URL（可含凭据，预先保存）',type:'password',visibleWhen:values=>values.endpoint_mode==='url'},
        {name:'push_protocol',label:'推送协议（Kafka TLS 由 use-ssl 参数控制）',type:'select',required:true,options:[{value:'https',label:'HTTPS'},{value:'http',label:'HTTP（明文）'},{value:'amqps',label:'AMQPS'},{value:'amqp',label:'AMQP（明文）'},{value:'kafka',label:'Kafka'}],visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_host',label:'主机名 / IPv4（不含协议、端口、路径）',required:true,visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_port',label:'端口（1–65535，留空不在 URL 中指定端口）',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_path',label:'HTTP 路径 / AMQP VHost（以 / 开头，Kafka 留空；不自动编码）',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_user_secret',label:'推送用户名（与密码同时填写，不自动编码）',type:'password',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_password_secret',label:'推送密码（不会出现在确认文案）',type:'password',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'persistent',label:'持久化',type:'select',required:true,options:[{value:'true',label:'开启'},{value:'false',label:'关闭'}]},
        {name:'time_to_live',label:'TTL 秒数（0 无限，None 全局默认）',required:true},{name:'max_retries',label:'最大重试（0 无限，None 全局默认）',required:true},{name:'retry_sleep_duration',label:'重试间隔秒数（0 无延迟，None 全局默认）',required:true},
        {name:'opaque_data',label:'Opaque Data（加入通知正文）',type:'textarea'},{name:'policy',label:'完整 Policy JSON（可为空）',type:'textarea'},
        {name:'options',label:'非凭据投递参数',required:true,renderControl:disabled=><RgwTopicOptionsEditor disabled={disabled} />},
        {name:'confirm_create',label:'影响确认',type:'select',required:true,options:[{value:'acknowledged',label:'已核对凭据、范围和端点，理解队列、权限及并发风险'}]}
      ],
      buildBody:(values,clusterId)=>({cluster_id:clusterId,...topicCreateInput(values)})
    },
    extraActions: [{
      title: '修改 Topic Policy', buttonLabel: '修改 Policy', path: '/rgw/topic/policy', method: 'PATCH',
      successMessage: 'Topic Policy 已回读核验',
      disabledWhen: topicPolicyBlocked, initialValues: topicPolicyInitial, confirmation: topicPolicyConfirmation,
      fields: [
        { name: 'topic_id', label: 'Topic ID（不可更改）', readOnly: true },
        { name: 'topic_arn', label: 'Topic ARN（不可更改）', readOnly: true },
        { name: 'policy_mode', label: '操作', type: 'select', required: true, options: [{ value: 'set', label: '替换完整 Policy' }, { value: 'clear', label: '清除 Policy' }] },
        { name: 'policy', label: '完整 Policy JSON（仅替换时使用）', type: 'textarea' }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...topicPolicyInput(values, row) })
    }, {
      title: '修改 Topic 通知属性', buttonLabel: '修改通知属性', path: '/rgw/topic/attribute', method: 'PATCH',
      successMessage: 'Topic 通知属性已回读核验（不代表消息已投递）',
      disabledWhen: topicAttributeBlocked, initialValues: topicAttributeInitial, confirmation: topicAttributeConfirmation,
      fields: [
        { name: 'topic_id', label: 'Topic ID（不可更改）', readOnly: true },
        { name: 'topic_arn', label: 'Topic ARN（不可更改）', readOnly: true },
        { name: 'attribute', label: '单次修改属性（当前值见列表或详情）', type: 'select', required: true, options: topicAttributeOptions },
        { name: 'mode', label: '操作（持久化忽略此项）', type: 'select', options: [{ value: 'set', label: '设置数值 / Opaque Data' }, { value: 'clear', label: '清除 Opaque Data' }, { value: 'default', label: '使用全局默认（仅数值参数）' }] },
        { name: 'value', label: '数值（0–2147483647）或 Opaque Data（保留原文）', type: 'textarea' },
        { name: 'persistent', label: '持久化（仅持久化属性使用）', type: 'select', options: [{ value: 'true', label: '开启持久化' }, { value: 'false', label: '关闭持久化（队列及未投递消息可能丢失）' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...topicAttributeInput(values, row) })
    }, {
      title: '修改 Topic 推送端点', buttonLabel: '修改推送端点', path: '/rgw/topic/endpoint', method: 'PATCH',
      successMessage: 'Topic 端点配置已回读核验（不代表投递成功）',
      disabledWhen: topicEndpointBlocked, initialValues: topicEndpointInitial, confirmation: topicEndpointConfirmation,
      fields: [
        { name: 'topic_id', label: 'Topic ID（不可更改）', readOnly: true },
        { name: 'topic_arn', label: 'Topic ARN（不可更改）', readOnly: true },
        { name: 'endpoint_mode', label: '操作', type: 'select', required: true, options: [{ value: 'replace', label: '替换完整 URL（含所需凭据和查询参数）' }, { value: 'fields', label: '分项填写完整新端点（旧凭据不保留）' }, { value: 'clear', label: '清空推送端点（队列可能被删除）' }] },
        { name: 'endpoint_secret', label: '完整新 URL（预先保存，不从脱敏数据回填）', type: 'password', visibleWhen: values => values.endpoint_mode === 'replace' },
        {name:'push_protocol',label:'推送协议（Kafka TLS 由 use-ssl 参数控制）',type:'select',required:true,options:[{value:'https',label:'HTTPS'},{value:'http',label:'HTTP（明文）'},{value:'amqps',label:'AMQPS'},{value:'amqp',label:'AMQP（明文）'},{value:'kafka',label:'Kafka'}],visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_host',label:'新主机名 / IPv4（不含协议、端口、路径）',required:true,visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_port',label:'新端口（1–65535，留空不在 URL 中指定端口）',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_path',label:'新 HTTP 路径 / AMQP VHost（以 / 开头，Kafka 留空；不自动编码）',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_user_secret',label:'新推送用户名（与密码同时填写，留空不保留旧 URL 凭据）',type:'password',visibleWhen:values=>values.endpoint_mode==='fields'},
        {name:'push_password_secret',label:'新推送密码（不回填，不出现在确认文案）',type:'password',visibleWhen:values=>values.endpoint_mode==='fields'},
        { name: 'confirm_endpoint', label: '影响确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已保存完整配置，确认目标可信、传输安全、保留参数及队列风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...topicEndpointInput(values, row) })
    }, {
      title: '修改 Topic 投递参数', buttonLabel: '修改协议参数', path: '/rgw/topic/option', method: 'PATCH',
      successMessage: 'Topic 投递参数已回读核验（不代表实际投递成功）',
      disabledWhen: topicOptionBlocked, initialValues: topicOptionInitial, confirmation: topicOptionConfirmation,
      fields: [
        {name:'topic_id',label:'Topic ID（不可更改）',readOnly:true}, {name:'topic_arn',label:'Topic ARN（不可更改）',readOnly:true},
        {name:'option',label:'单次修改参数（当前状态见详情）',type:'select',required:true,options:topicWritableOptions},
        {name:'value_mode',label:'赋值方式',type:'select',required:true,options:[{value:'set',label:'设置指定值'},{value:'empty',label:'显式空值（仅 CA / Exchange，不是删除）'}]},
        {name:'value',label:'新值（SASL: PLAIN / SCRAM-SHA-256 / SCRAM-SHA-512 / GSSAPI / OAUTHBEARER）'},
        {name:'confirm_option',label:'影响确认',type:'select',required:true,options:[{value:'acknowledged',label:'确认当前协议支持此值，理解 TLS、认证及投递可靠性风险'}]}
      ],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,...topicOptionInput(values,row)})
    }],
    deleteAction: {
      title: '删除 RGW 通知目标', path: '/rgw/topic', action: 'rgw_topic.delete', resourceKind: 'rgw_topic', risk: 'high',
      successMessage: 'Topic 删除命令成功，元数据不存在已核验；请检查桶通知引用',
      disabledWhen: topicDeleteBlocked, confirmation: topicDeleteConfirmation,
      resourceKey: (row) => String(row.metadata_key),
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...topicDeleteInput(row) })
    },
    detailContent: (row) => <RgwTopicDetails row={row} />,
    columns: [
      { key: 'name', title: '名称', render: topicText },
      { key: 'owner', title: 'Owner', render: topicText },
      { key: 'scope', title: '租户 / Account', render: (value) => value === '' ? '全局租户' : topicText(value) },
      { key: 'arn', title: 'ARN', render: topicText },
      { key: 'push_endpoint', title: '推送端点（已脱敏）', render: (_, row) => topicEndpoint(row) },
      { key: 'persistent', title: '持久化', render: topicBoolean },
      { key: 'time_to_live', title: 'TTL（原生值）', render: topicText },
      { key: 'max_retries', title: '重试次数（原生值）', render: topicText },
      { key: 'retry_sleep_duration', title: '重试等待（原生值）', render: topicText }
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
    }, {
      title:'管理角色托管策略',buttonLabel:'管理托管策略',path:'/rgw/role/managed/policy',method:'PATCH',successMessage:'角色托管策略集合已回读核验',
      disabledWhen:roleManagedPolicyBlocked,initialValues:roleManagedPolicyInitial,confirmation:roleManagedPolicyConfirmation,
      fields:[
        {name:'account_id',label:'Account ID（不可更改）',readOnly:true},{name:'name',label:'角色名称（不可更改）',readOnly:true},
        {name:'owner_uid',label:'已配置 S3 永久密钥所属完整 UID（后端核验 Account）',required:true},
        {name:'mode',label:'操作',type:'select',required:true,options:[{value:'attach',label:'关联托管策略'},{value:'detach',label:'解除托管策略'}]},
        {name:'policy_source',label:'策略来源（参考列表不代表目标版本支持）',type:'select',required:true,visibleWhen:values=>values.mode === 'attach',options:[{value:'reference',label:'参考版本内置策略'},{value:'custom',label:'手填目标版本支持的 ARN'}]},
        {name:'policy_arn',label:'托管策略（解除仅显示当前已关联项）',type:'select',required:true,visibleWhen:values=>values.mode === 'detach' || (values.mode === 'attach' && values.policy_source === 'reference'),optionsDependencies:['mode','policy_source'],optionsLoader:async (_clusterId,row,values)=>roleManagedPolicyOptions(row,values?.mode)},
        {name:'custom_policy_arn',label:'完整托管策略 ARN（由目标 Ceph 判定支持）',required:true,visibleWhen:values=>values.mode === 'attach' && values.policy_source === 'custom'},
        {name:'confirm_managed_policy',label:'影响确认',type:'select',required:true,options:[{value:'acknowledged',label:'已确认准确角色、权限变化和非原子操作风险'}]}
      ],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,...roleManagedPolicyInput(values,row)})
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
        { name: 'name', label: 'Bucket 名称', required: true },
        { name: 'tenant', label: '租户（留空明确表示全局租户；须与 S3 凭据权限匹配）' }
      ],
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, name: String(values.name ?? ''), tenant: String(values.tenant ?? '') })
    },
    updateAction: {
      title: '更新 Bucket 版本控制',
      path: '/rgw/bucket',
      method: 'PATCH',
      successMessage: 'Bucket 版本控制更新成功并已回读核验',
      confirmation: (values, row) => bucketVersioningConfirmation(values, bucketId(row)),
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
      initialValues: bucketVersioningInitial,
      buildBody: (values, clusterId, row) => ({
        cluster_id: clusterId,
        ...bucketVersioningInput(values, bucketId(row))
      })
    },
    extraActions: [{
      title: '创建桶同步组', path: '/rgw/bucket/sync/group', method: 'POST',
      successMessage: '桶本地同步组已创建并回读核验（尚无数据流和管道）',
      disabledWhen: bucketSyncGroupCreateBlocked, initialValues: bucketSyncGroupCreateInitial, confirmation: bucketSyncGroupCreateConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '新的同步组 ID（不可与已有组重复）', required: true },
        { name: 'status', label: '初始状态', type: 'select', required: true, options: [{ value: 'enabled', label: 'enabled 启用' }, { value: 'allowed', label: 'allowed 允许但不启用' }, { value: 'forbidden', label: 'forbidden 禁止' }] },
        { name: 'confirm_create', label: '创建范围确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认仅创建桶本地空组，不会建立完整复制链路' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncGroupCreateInput(values, row) })
    }, {
      title: '删除桶同步组', path: '/rgw/bucket/sync/group', method: 'DELETE',
      successMessage: '桶同步组删除已回读核验',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncGroupDeleteConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '要删除的完整同步组 ID', required: true },
        { name: 'confirm_delete', label: '删除范围确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认删除整个组及其全部数据流和管道，已备份策略' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncGroupDeleteInput(values, row) })
    }, {
      title: '创建桶数据流', path: '/rgw/bucket/sync/flow', method: 'POST',
      successMessage: '桶本地数据流创建已回读核验（不代表同步完成）',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncFlowConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'flow_type', label: '数据流类型', type: 'select', required: true, options: [{ value: 'symmetrical', label: '对称' }, { value: 'directional', label: '定向' }] },
        { name: 'flow_id', label: '对称流 ID（定向流留空）' },
        { name: 'zones_json', label: '对称 Zone ID 数组，如 ["zone-id-a","zone-id-b"]（定向流留空）', type: 'textarea' },
        { name: 'source_zone', label: '定向源 Zone ID（不是名称，对称流留空）' },
        { name: 'dest_zone', label: '定向目标 Zone ID（不是名称，对称流留空）' },
        { name: 'confirm_flow', label: '数据流影响确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认 Zone ID 正确，已备份策略并了解复制影响' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncFlowInput(values, row) })
    }, {
      title: '编辑桶对称数据流', path: '/rgw/bucket/sync/flow', method: 'PATCH',
      successMessage: '对称流 Zone 成员已回读核验（不代表同步完成）',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncFlowUpdateConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'flow_id', label: '已有对称流 ID', required: true },
        { name: 'zones_json', label: '完整目标 Zone ID JSON 数组（不是名称，不可为空）', type: 'textarea', required: true },
        { name: 'confirm_flow_update', label: '分步修改确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解先添加再移除及部分生效风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncFlowUpdateInput(values, row) })
    }, {
      title: '删除桶数据流', path: '/rgw/bucket/sync/flow', method: 'DELETE',
      successMessage: '桶本地数据流删除已回读核验',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncFlowDeleteConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'flow_type', label: '数据流类型', type: 'select', required: true, options: [{ value: 'symmetrical', label: '对称（删除整个流）' }, { value: 'directional', label: '定向（删除一对源/目标）' }] },
        { name: 'flow_id', label: '已有对称流 ID（定向流留空）' },
        { name: 'source_zone', label: '定向源 Zone ID（不是名称，对称流留空）' },
        { name: 'dest_zone', label: '定向目标 Zone ID（不是名称，对称流留空）' },
        { name: 'confirm_flow_delete', label: '删除确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认删除整个数据流，已备份策略并了解复制影响' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncFlowDeleteInput(values, row) })
    }, {
      title: '删除桶同步管道', path: '/rgw/bucket/sync/pipe', method: 'DELETE',
      successMessage: '桶本地同步管道删除已回读核验',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncPipeDeleteConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'pipe_id', label: '要删除的完整管道 ID', required: true },
        { name: 'confirm_pipe_delete', label: '删除范围确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认删除整个管道及其全部选择器和参数，已备份策略' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncPipeDeleteInput(values, row) })
    }, {
      title: '创建桶同步管道', path: '/rgw/bucket/sync/pipe', method: 'POST',
      successMessage: '桶本地同步管道创建已回读核验（不代表同步完成）',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncPipeCreateConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'pipe_id', label: '新管道 ID', required: true },
        { name: 'source_zones_json', label: '源 Zone ID 数组（所有 Zone 填 ["*"]）', type: 'textarea', required: true },
        { name: 'dest_zones_json', label: '目标 Zone ID 数组（所有 Zone 填 ["*"]）', type: 'textarea', required: true },
        { name: 'source_tenant', label: '源租户（空或 * 均不限定租户）' },
        { name: 'source_bucket', label: '源桶名（* 为通配）', required: true },
        { name: 'source_bucket_id', label: '源桶实例 ID（空或 * 均不限定实例）' },
        { name: 'dest_tenant', label: '目标租户（空或 * 均不限定租户）' },
        { name: 'dest_bucket', label: '目标桶名（* 为通配）', required: true },
        { name: 'dest_bucket_id', label: '目标桶实例 ID（空或 * 均不限定实例）' },
        { name: 'mode', label: '权限模式', type: 'select', required: true, options: [{ value: 'system', label: 'system 系统模式' }, { value: 'user', label: 'user 指定用户模式' }] },
        { name: 'user', label: '完整用户 UID（user 模式必填，system 留空）' },
        { name: 'confirm_pipe_create', label: '范围确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认匹配范围和模式，已备份策略并了解复制影响' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncPipeCreateInput(values, row) })
    }, {
      title: '编辑桶同步管道 Zone', path: '/rgw/bucket/sync/pipe/zones', method: 'PATCH',
      successMessage: '管道 Zone 成员已回读核验（不代表同步完成）',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncPipeZonesConfirmation,
      changedValues: bucketSyncPipeZonesSelectionChanged,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', type: 'select', required: true, optionsLoader: async (_clusterId, row) => bucketSyncPipeGroupOptions(row) },
        { name: 'pipe_id', label: '已有管道 ID（选择后载入成员）', type: 'select', required: true, optionsDependencies: ['group_id'], optionsLoader: async (_clusterId, row, values) => bucketSyncPipeOptions(row, values?.group_id) },
        { name: 'pipe_load_error', label: '成员载入失败', readOnly: true, visibleWhen: values => !!values.pipe_load_error },
        { name: 'source_zones_json', label: '完整源 Zone ID JSON 数组（不是名称；全部填 ["*"]）', type: 'textarea', required: true },
        { name: 'dest_zones_json', label: '完整目标 Zone ID JSON 数组（不是名称；全部填 ["*"]）', type: 'textarea', required: true },
        { name: 'confirm_pipe_zones', label: '范围确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解通配范围及分步修改可能部分生效' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncPipeZonesInput(values, row) })
    }, {
      title: '编辑桶同步管道配置', path: '/rgw/bucket/sync/pipe', method: 'PATCH',
      successMessage: '管道选择器和模式已回读核验（不代表同步完成）',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncPipeUpdateConfirmation,
      changedValues: bucketSyncPipeSelectionChanged,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', type: 'select', required: true, optionsLoader: async (_clusterId, row) => bucketSyncPipeGroupOptions(row) },
        { name: 'pipe_id', label: '已有管道 ID（选择后回填）', type: 'select', required: true, optionsDependencies: ['group_id'], optionsLoader: async (_clusterId, row, values) => bucketSyncPipeOptions(row, values?.group_id) },
        { name: 'pipe_load_error', label: '配置载入失败', readOnly: true, visibleWhen: values => !!values.pipe_load_error },
        { name: 'source_tenant', label: '源租户（空或 * 不限定租户）' },
        { name: 'source_bucket', label: '完整目标源桶名（* 通配）', required: true },
        { name: 'source_bucket_id', label: '源桶实例 ID（空或 * 不限定实例）' },
        { name: 'dest_tenant', label: '目标租户（空或 * 不限定租户）' },
        { name: 'dest_bucket', label: '完整目标桶名（* 通配）', required: true },
        { name: 'dest_bucket_id', label: '目标桶实例 ID（空或 * 不限定实例）' },
        { name: 'mode', label: '权限模式', type: 'select', required: true, options: [{ value: 'system', label: 'system 系统模式' }, { value: 'user', label: 'user 指定用户模式' }] },
        { name: 'user', label: '完整 UID（user 必填；system 留空，保留已存储 UID）' },
        { name: 'priority', label: '优先级（留空保持原值；可能改变匹配管道选择）', type: 'number', min: -2147483648, max: 2147483647 },
        { name: 'storage_class_mode', label: '目标存储类变更', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'set',label:'设置完整存储类名称'},{value:'empty',label:'设置空字符串（不是移除覆盖字段）'}] },
        { name: 'storage_class', label: '目标存储类（需自行确认目标放置配置）', required: true, visibleWhen: values => values.storage_class_mode === 'set' },
        { name: 'prefix_mode', label: '源前缀过滤变更', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'set',label:'设置完整前缀'},{value:'empty',label:'设置空前缀（保留字段）'},{value:'remove',label:'移除前缀字段'}] },
        { name: 'source_prefix', label: '源对象前缀（保留空格，不是正则表达式）', required: true, visibleWhen: values => values.prefix_mode === 'set' },
        { name: 'tags_mode', label: '源标签过滤变更', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'change',label:'增删完整标签对'}] },
        { name: 'tags_remove_json', label: '移除标签对 JSON 数组（如 [{"key":"env","value":"old"}]；空白不移除）', type: 'textarea', visibleWhen: values => values.tags_mode === 'change' },
        { name: 'tags_add_json', label: '添加标签对 JSON 数组（键值不能含逗号，键不能含等号；空白不添加）', type: 'textarea', visibleWhen: values => values.tags_mode === 'change' },
        { name: 'acl_mode', label: '目标 ACL 转换', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'set',label:'设置目标所有者 UID'},{value:'remove',label:'移除 ACL 转换'}] },
        { name: 'dest_owner', label: '目标所有者完整 UID（不是 Account ID，不修改桶所有者）', required: true, visibleWhen: values => values.acl_mode === 'set' },
        { name: 'confirm_pipe_update', label: '修改确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份并确认选择器、优先级和权限变化；Zone 成员保持不变' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncPipeUpdateInput(values, row) })
    }, {
      title: '修改桶同步组状态', path: '/rgw/bucket/sync/group', method: 'PATCH',
      successMessage: '桶同步组状态已回读核验（不代表同步完成）',
      disabledWhen: bucketSyncGroupBlocked, initialValues: bucketSyncGroupInitial, confirmation: bucketSyncGroupConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID（从原生策略中复制）', required: true },
        { name: 'status', label: '目标状态', type: 'select', required: true, options: [{ value: 'enabled', label: 'enabled 启用' }, { value: 'allowed', label: 'allowed 允许但不启用' }, { value: 'forbidden', label: 'forbidden 禁止' }] },
        { name: 'confirm_change', label: '复制影响确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认此修改可能改变复制行为，且不会删除已有副本' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketSyncGroupInput(values, row) })
    }, { title: 'Bucket 配额设置', path: '/rgw/bucket/quota', method: 'PUT', successMessage: 'Bucket 配额设置执行成功',
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
      disabledWhen: bucketDeleteBlocked,
      confirmation: bucketDeleteConfirmation,
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...bucketDeleteInput(row) }),
      resourceKey: (row) => `rgw/bucket/${bucketDeleteInput(row).bucket_id}`
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
      { key: 'lifecycle_progress', title: '生命周期处理进度（采集时）', ellipsis: false, render: rgwLifecycleProgress },
      { key: 'bucket_sync_policy', title: '桶本地同步策略（采集时）', ellipsis: false, render: (value) => <div>{rgwBucketSyncPolicy(value)}<details><summary>查看数据流</summary><RgwBucketSyncFlows value={value} /></details><details><summary>查看同步管道</summary><RgwBucketSyncPipes value={value} /></details>{value != null && <details><summary>查看原生策略（数据流与管道）</summary><pre>{JSON.stringify(value, null, 2)}</pre></details>}</div> },
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
    deleteAction: {
      title: '删除 Realm 配置', action: 'rgw_realm.delete', resourceKind: 'rgw_realm', risk: 'high',
      resourceKey: (row) => String(row.name),
      path: '/rgw/realm',
      disabledWhen: realmDeleteBlocked,
      confirmation: realmDeleteConfirmation,
      successMessage: 'Realm ID 与名称索引删除已核验；关联资源与引用未清理',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...realmDeleteInput(row) })
    },
    toolbarActions: [rgwRealmImportAction, rgwRealmSetupAction, rgwRealmMigrationAction],
    title: 'RGW Multisite',
    detailContent: (row, clusterId) => <><RgwRealmToken key={`token:${clusterId}:${row.id}:${row.name}`} row={row} clusterId={clusterId} /><RgwRealmTransfer key={`transfer:${clusterId}:${row.id}:${row.name}`} row={row} clusterId={clusterId} /></>,
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
      { key: 'current_period_details', title: '当前 Period 快照', ellipsis: false, render: (value, row) => <RgwCurrentPeriod value={value} realm={row.id} current={row.current_period} /> },
      { key: 'epoch', title: 'Epoch' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwZonegroups: {
    deleteAction: {
      title: '删除 Zonegroup（保留 Zone 和池）', path: '/rgw/zonegroup', action: 'rgw_zonegroup.delete', resourceKind: 'rgw_zonegroup', risk: 'high',
      resourceKey: (row) => String(row.name),
      disabledWhen: zonegroupDeleteBlocked, confirmation: zonegroupDeleteConfirmation,
      successMessage: 'Zonegroup 删除及适用的 Period 发布已核验；Zone、池和服务未删除',
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...zonegroupDeleteInput(row) })
    },
    title: 'RGW ZoneGroups',
    detailContent: (row, clusterId) => <RgwLocalClassDetails row={row} clusterId={clusterId} />,
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
    extraActions: [{
      title: '新建 Zonegroup 放置目标', path: '/rgw/zonegroup/placement', method: 'POST',
      successMessage: '组放置目标创建已回读核验（未配置 Zone 池或发布 Period）',
      disabledWhen: groupPlacementCreateBlocked, confirmation: groupPlacementCreateConfirmation,
      initialValues: () => ({tags_json:'[]'}),
      changedValues: changed => Object.keys(changed).some(key => key !== 'confirm_create') ? {confirm_create:undefined} : {},
      fields: [
        { name: 'placement_id', label: '新放置目标名称（初始类为 STANDARD）', required: true },
        { name: 'tags_json', label: '目标标签 JSON 字符串数组（无标签填 []；标签不能含逗号）', type: 'textarea', required: true },
        { name: 'confirm_create', label: '范围确认', type: 'select', required: true, options: [{value:'acknowledged',label:'已备份，了解默认目标初始化、不配置 Zone 池和不发布 Period'}] }
      ],
      buildBody: (values,clusterId,row) => ({cluster_id:clusterId,...groupPlacementCreateInput(values,row)})
    }, {
      title: '设置默认放置规则', path: '/rgw/zonegroup/placement/default', method: 'PATCH',
      successMessage: '默认放置规则已回读核验（未发布 Period）',
      disabledWhen: groupStorageClassBlocked, confirmation: groupPlacementDefaultConfirmation,
      changedValues: changed => ({...(Object.prototype.hasOwnProperty.call(changed,'placement_id') ? {storage_class:undefined} : {}),...(Object.keys(changed).some(key => key !== 'confirm_default') ? {confirm_default:undefined} : {})}),
      fields: [
        {name:'placement_id',label:'已有放置目标',type:'select',required:true,optionsLoader:async (_clusterId,row) => groupStorageClassOptions(row)},
        {name:'storage_class',label:'已声明存储类',type:'select',required:true,optionsDependencies:['placement_id'],optionsLoader:async (_clusterId,row,values) => groupPlacementDefaultClasses(row,values?.placement_id)},
        {name:'confirm_default',label:'范围确认',type:'select',required:true,options:[{value:'acknowledged',label:'已备份，了解默认行为变化与后续发布要求'}]}
      ],
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,...groupPlacementDefaultInput(values,row)})
    }, {
      title: '编辑放置目标标签', path: '/rgw/zonegroup/placement/tags', method: 'PATCH',
      successMessage: '目标标签已回读核验（未发布 Period）',
      disabledWhen: groupStorageClassBlocked, confirmation: groupPlacementTagsConfirmation, changedValues: groupPlacementTagsChanged,
      fields: [
        {name:'placement_id',label:'已有放置目标',type:'select',required:true,optionsLoader:async (_clusterId,row) => groupStorageClassOptions(row)},
        {name:'storage_class',label:'已有类（命令使用；标签影响整个目标）',type:'select',required:true,optionsDependencies:['placement_id'],optionsLoader:async (_clusterId,row,values) => groupPlacementDefaultClasses(row,values?.placement_id)},
        {name:'tags_json',label:'目标标签 JSON 数组（[] 清空，不能含逗号）',type:'textarea',required:true},
        {name:'confirm_tags',label:'范围确认',type:'select',required:true,options:[{value:'acknowledged',label:'已备份，了解目标标签、默认值和发布范围'}]}
      ],
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,...groupPlacementTagsInput(values,row)})
    }, {
      title:'编辑云分层恢复配置',path:'/rgw/zonegroup/placement/restore',method:'PATCH',
      successMessage:'云分层恢复配置及 Realm Period 已回读核验（未直接恢复对象）',
      disabledWhen:cloudRestoreBlocked,confirmation:cloudRestoreConfirmation,changedValues:cloudRestoreChanged,
      fields:[
        {name:'placement_id',label:'已有云分层目标',type:'select',required:true,optionsLoader:async (_clusterId,row)=>cloudRestoreTargets(row)},
        {name:'storage_class',label:'已有云分层类',type:'select',required:true,optionsDependencies:['placement_id'],optionsLoader:async (_clusterId,row,values)=>cloudRestoreClasses(row,values?.placement_id)},
        {name:'tier_type',label:'原生分层类型（不可更改）',readOnly:true},
        {name:'retain_head_object',label:'保留头对象（影响后续恢复）',type:'select',required:true,options:[{value:'true',label:'是'},{value:'false',label:'否'}]},
        {name:'allow_read_through',label:'允许读穿透（可能产生远端请求及费用）',type:'select',required:true,options:[{value:'true',label:'是'},{value:'false',label:'否'}]},
        {name:'read_through_restore_days',label:'读穿透恢复天数（原生值）',type:'number',required:true,min:0,max:Number.MAX_SAFE_INTEGER},
        {name:'restore_storage_class',label:'恢复本地类（需自行核对各 Zone 池映射）',type:'select',required:true,optionsDependencies:['placement_id'],optionsLoader:async (_clusterId,row,values)=>cloudRestoreLocalClasses(row,values?.placement_id)},
        {name:'glacier_restore_days',label:'Glacier 恢复天数（不同于读穿透天数）',type:'number',required:true,min:0,max:Number.MAX_SAFE_INTEGER,visibleWhen:values=>values.tier_type==='cloud-s3-glacier'},
        {name:'glacier_restore_tier_type',label:'Glacier 恢复等级（核对目标支持和费用）',type:'select',required:true,visibleWhen:values=>values.tier_type==='cloud-s3-glacier',options:[{value:'Standard',label:'Standard'},{value:'Expedited',label:'Expedited'}]},
        {name:'confirm_restore',label:'变更与发布确认',type:'select',required:true,options:[{value:'acknowledged',label:'已备份并核对恢复类、头对象保留及 Realm 待发布变更'}]}
      ],
      buildBody:(values,clusterId,row)=>({cluster_id:clusterId,...cloudRestoreInput(values,row)})
    }, {
      title: '声明 Zonegroup 存储类', path: '/rgw/zonegroup/storage/class', method: 'POST',
      successMessage: '组存储类声明已回读核验（未配置 Zone 池或发布 Period）',
      disabledWhen: groupStorageClassBlocked, confirmation: groupStorageClassConfirmation,
      changedValues: changed => Object.keys(changed).some(key => key !== 'confirm_create') ? {confirm_create:undefined} : {},
      fields: [
        { name: 'placement_id', label: '已有放置目标', type: 'select', required: true, optionsLoader: async (_clusterId,row) => groupStorageClassOptions(row) },
        { name: 'storage_class', label: '新普通存储类名称', required: true },
        { name: 'confirm_create', label: '范围确认', type: 'select', required: true, options: [{value:'acknowledged',label:'已备份，了解默认目标初始化、不配置 Zone 池和不发布 Period'}] }
      ],
      buildBody: (values,clusterId,row) => ({cluster_id:clusterId,...groupStorageClassInput(values,row)})
    }, {
      title: '删除组存储类（保留 Zone 映射）', path: '/rgw/zonegroup/storage/class', method: 'DELETE',
      successMessage: '组存储类删除与适用的 Period 发布已核验（未删除对象或 Zone 映射）',
      disabledWhen: groupStorageClassBlocked, confirmation: groupStorageClassDeleteConfirmation,
      changedValues: changed => ({...(Object.prototype.hasOwnProperty.call(changed,'placement_id') ? {storage_class:undefined} : {}),...(Object.keys(changed).some(key => key !== 'confirm_delete') ? {confirm_delete:undefined} : {})}),
      fields: [
        {name:'placement_id',label:'已有放置目标',type:'select',required:true,optionsLoader:async (_clusterId,row) => groupStorageClassOptions(row)},
        {name:'storage_class',label:'待删除存储类（同时移除该类分层配置）',type:'select',required:true,optionsDependencies:['placement_id'],optionsLoader:async (_clusterId,row,values) => groupPlacementDefaultClasses(row,values?.placement_id)},
        {name:'confirm_delete',label:'删除与发布确认',type:'select',required:true,options:[{value:'acknowledged',label:'已备份并核对数据依赖，了解默认类回退、Realm 发布与 Zone 映射保留'}]}
      ],
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,...groupStorageClassDeleteInput(values,row)})
    }, {
      title: '联动删除本地存储类', path: '/rgw/zonegroup/storage/class/local', method: 'DELETE',
      successMessage: 'Zone 映射、组声明删除与适用的 Period 发布已核验（未删除池或对象）',
      disabledWhen: groupStorageClassBlocked, confirmation: groupLocalClassDeleteConfirmation,
      changedValues: changed => ({...(Object.prototype.hasOwnProperty.call(changed,'placement_id') ? {storage_class:undefined} : {}),...(Object.keys(changed).some(key => key !== 'confirm_delete') ? {confirm_delete:undefined} : {})}),
      fields: [
        {name:'zone_id',label:'仅删除此成员 Zone 的类映射',type:'select',required:true,optionsLoader:async (_clusterId,row) => groupLocalClassDeleteZones(row)},
        {name:'placement_id',label:'已有放置目标',type:'select',required:true,optionsLoader:async (_clusterId,row) => groupStorageClassOptions(row)},
        {name:'storage_class',label:'待删除本地类（不支持 STANDARD）',type:'select',required:true,optionsDependencies:['placement_id'],optionsLoader:async (_clusterId,row,values) => groupLocalClassDeleteClasses(row,values?.placement_id)},
        {name:'confirm_delete',label:'联动删除确认',type:'select',required:true,options:[{value:'acknowledged',label:'已备份，了解数据访问风险、部分生效和 Realm 发布范围'}]}
      ],
      buildBody:(values,clusterId,row) => ({cluster_id:clusterId,...groupLocalClassDeleteInput(values,row)})
    }, {
      title: '创建 Zonegroup 同步组', path: '/rgw/zonegroup/sync/group', method: 'POST',
      successMessage: 'Zonegroup 空同步组创建与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncCreateBlocked, initialValues: zonegroupSyncCreateInitial, confirmation: zonegroupSyncCreateConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '新同步组 ID', required: true },
        { name: 'status', label: '初始状态', type: 'select', required: true, options: [{ value: 'enabled', label: 'enabled 启用' }, { value: 'allowed', label: 'allowed 允许但不启用' }, { value: 'forbidden', label: 'forbidden 禁止' }] },
        { name: 'confirm_create', label: '创建与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解空组语义、Realm 发布范围及并发风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupSyncCreateInput(values,row) })
    }, {
      title: '创建 Zonegroup 同步流', path: '/rgw/zonegroup/sync/flow', method: 'POST',
      successMessage: 'Zonegroup 同步流创建与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupFlowCreateConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'flow_type', label: '流类型', type: 'select', required: true, options: [{ value: 'symmetrical', label: '对称' }, { value: 'directional', label: '定向' }] },
        { name: 'flow_id', label: '新流 ID（仅对称）' },
        { name: 'zones', label: 'Zone ID 列表（仅对称，逗号分隔）' },
        { name: 'source_zone', label: '源 Zone ID（仅定向）' },
        { name: 'dest_zone', label: '目标 Zone ID（仅定向）' },
        { name: 'confirm_flow', label: '创建与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份并了解复制路径、Realm 发布与并发风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupFlowCreateInput(values,row) })
    }, {
      title: '删除 Zonegroup 同步流', path: '/rgw/zonegroup/sync/flow', method: 'DELETE',
      successMessage: 'Zonegroup 同步流删除与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupFlowDeleteConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'flow_type', label: '流类型', type: 'select', required: true, options: [{ value: 'symmetrical', label: '整条对称流' }, { value: 'directional', label: '定向流' }] },
        { name: 'flow_id', label: '已有流 ID（仅对称，删除全部成员）' },
        { name: 'source_zone', label: '策略中的源 Zone ID（仅定向）' },
        { name: 'dest_zone', label: '策略中的目标 Zone ID（仅定向）' },
        { name: 'confirm_flow_delete', label: '删除与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份并了解整流删除、Realm 发布与并发风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupFlowDeleteInput(values,row) })
    }, {
      title: '修改 Zonegroup 对称流成员', path: '/rgw/zonegroup/sync/flow', method: 'PATCH',
      successMessage: 'Zonegroup 对称流成员与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupFlowUpdateConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'flow_id', label: '已有对称流 ID', required: true },
        { name: 'zones', label: '最终 Zone ID 列表（非空，逗号分隔）', required: true },
        { name: 'confirm_flow_update', label: '分步修改与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解中间范围扩大、Realm 发布和部分生效风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupFlowUpdateInput(values,row) })
    }, {
      title: '创建 Zonegroup 同步管道', path: '/rgw/zonegroup/sync/pipe', method: 'POST',
      successMessage: 'Zonegroup 管道创建与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupPipeCreateConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'pipe_id', label: '新管道 ID', required: true },
        { name: 'source_zones_json', label: '源 Zone ID 数组（所有 Zone 填 ["*"]）', type: 'textarea', required: true },
        { name: 'dest_zones_json', label: '目标 Zone ID 数组（所有 Zone 填 ["*"]）', type: 'textarea', required: true },
        { name: 'source_tenant', label: '源租户（空或 * 均不限定租户）' },
        { name: 'source_bucket', label: '源桶名（* 为通配）', required: true },
        { name: 'source_bucket_id', label: '源实例 ID（空或 * 均不限定实例）' },
        { name: 'dest_tenant', label: '目标租户（空或 * 均不限定租户）' },
        { name: 'dest_bucket', label: '目标桶名（* 为通配）', required: true },
        { name: 'dest_bucket_id', label: '目标实例 ID（空或 * 均不限定实例）' },
        { name: 'mode', label: '权限模式', type: 'select', required: true, options: [{ value: 'system', label: 'system 系统模式' }, { value: 'user', label: 'user 指定用户模式' }] },
        { name: 'user', label: '完整 UID（user 必填，system 留空）' },
        { name: 'confirm_pipe_create', label: '范围及发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解复制范围、权限模式和 Realm 发布风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupPipeCreateInput(values,row) })
    }, {
      title: '删除 Zonegroup 同步管道', path: '/rgw/zonegroup/sync/pipe', method: 'DELETE',
      successMessage: 'Zonegroup 管道删除与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupPipeDeleteConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'pipe_id', label: '要删除的完整管道 ID', required: true },
        { name: 'confirm_pipe_delete', label: '删除及发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，确认删除整个管道，了解 Realm 发布和部分生效风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupPipeDeleteInput(values,row) })
    }, {
      title: '编辑 Zonegroup 管道选择器与身份', path: '/rgw/zonegroup/sync/pipe', method: 'PATCH',
      successMessage: 'Zonegroup 管道配置与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupPipeUpdateConfirmation,
      changedValues: zonegroupPipeSelectionChanged,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', type: 'select', required: true, optionsLoader: async (_clusterId, row) => zonegroupPipeGroupOptions(row) },
        { name: 'pipe_id', label: '已有管道 ID（选择后回填）', type: 'select', required: true, optionsDependencies: ['group_id'], optionsLoader: async (_clusterId, row, values) => zonegroupPipeOptions(row, values?.group_id) },
        { name: 'pipe_load_error', label: '配置载入失败', readOnly: true, visibleWhen: values => !!values.pipe_load_error },
        { name: 'source_tenant', label: '完整源租户（空或 * 不限租户）' },
        { name: 'source_bucket', label: '完整源桶名（* 为通配）', required: true },
        { name: 'source_bucket_id', label: '完整源实例 ID（空或 * 不限实例）' },
        { name: 'dest_tenant', label: '完整目标租户（空或 * 不限租户）' },
        { name: 'dest_bucket', label: '完整目标桶名（* 为通配）', required: true },
        { name: 'dest_bucket_id', label: '完整目标实例 ID（空或 * 不限实例）' },
        { name: 'mode', label: '目标权限模式', type: 'select', required: true, options: [{ value: 'system', label: 'system（保留存储的 UID）' }, { value: 'user', label: 'user 指定用户模式' }] },
        { name: 'user', label: '完整 UID（user 必填，system 留空）' },
        { name: 'priority', label: '优先级（留空保持原值；可能改变匹配管道选择）', type: 'number', min: -2147483648, max: 2147483647 },
        { name: 'storage_class_mode', label: '目标存储类变更', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'set',label:'设置完整存储类名称'},{value:'empty',label:'设置空字符串（不是移除覆盖字段）'}] },
        { name: 'storage_class', label: '目标存储类（需自行确认目标放置配置）', required: true, visibleWhen: values => values.storage_class_mode === 'set' },
        { name: 'prefix_mode', label: '源前缀过滤变更', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'set',label:'设置完整前缀'},{value:'empty',label:'设置空前缀（保留字段）'},{value:'remove',label:'移除前缀字段'}] },
        { name: 'source_prefix', label: '源对象前缀（保留空格，不是正则表达式）', required: true, visibleWhen: values => values.prefix_mode === 'set' },
        { name: 'tags_mode', label: '源标签过滤变更', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'change',label:'增删完整标签对'}] },
        { name: 'tags_remove_json', label: '移除标签对 JSON 数组（如 [{"key":"env","value":"old"}]；空白不移除）', type: 'textarea', visibleWhen: values => values.tags_mode === 'change' },
        { name: 'tags_add_json', label: '添加标签对 JSON 数组（键值不能含逗号，键不能含等号；空白不添加）', type: 'textarea', visibleWhen: values => values.tags_mode === 'change' },
        { name: 'acl_mode', label: '目标 ACL 转换', type: 'select', options: [{value:'preserve',label:'保持原值'},{value:'set',label:'设置目标所有者 UID'},{value:'remove',label:'移除 ACL 转换'}] },
        { name: 'dest_owner', label: '目标所有者完整 UID（不是 Account ID，不修改桶所有者）', required: true, visibleWhen: values => values.acl_mode === 'set' },
        { name: 'confirm_pipe_update', label: '配置与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，确认完整选择器、优先级与权限模式，了解 Realm 发布及并发风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupPipeUpdateInput(values,row) })
    }, {
      title: '编辑 Zonegroup 管道 Zone', path: '/rgw/zonegroup/sync/pipe/zones', method: 'PATCH',
      successMessage: 'Zonegroup 管道 Zone 成员与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupPipeZonesConfirmation,
      changedValues: zonegroupPipeZonesSelectionChanged,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', type: 'select', required: true, optionsLoader: async (_clusterId, row) => zonegroupPipeGroupOptions(row) },
        { name: 'pipe_id', label: '已有管道 ID（选择后载入成员）', type: 'select', required: true, optionsDependencies: ['group_id'], optionsLoader: async (_clusterId, row, values) => zonegroupPipeOptions(row, values?.group_id) },
        { name: 'pipe_load_error', label: '成员载入失败', readOnly: true, visibleWhen: values => !!values.pipe_load_error },
        { name: 'source_zones_json', label: '最终源 Zone ID 数组（须为当前 Zonegroup 成员；全部填 ["*"]）', type: 'textarea', required: true },
        { name: 'dest_zones_json', label: '最终目标 Zone ID 数组（全部填 ["*"]）', type: 'textarea', required: true },
        { name: 'confirm_pipe_zones', label: '成员与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解中间范围变化、Realm 发布与部分生效风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupPipeZonesInput(values,row) })
    }, {
      title: '准备桶复制上层策略', path: '/rgw/zonegroup/replication/prepare', method: 'POST',
      successMessage: '上层 allowed 策略准备与 Period 发布已核验（未写入桶 S3 复制规则）',
      disabledWhen: zonegroupReplicationPrepareBlocked, initialValues: zonegroupReplicationPrepareInitial, confirmation: zonegroupReplicationPrepareConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'confirm_replication_prepare', label: '上层许可与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份，了解全桶通配许可、Realm 发布及部分生效风险；尚未启用桶复制' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupReplicationPrepareInput(values,row) })
    }, {
      title: '删除 Zonegroup 同步组', path: '/rgw/zonegroup/sync/group', method: 'DELETE',
      successMessage: 'Zonegroup 同步组删除与适用的 Period 发布已核验',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupSyncDeleteConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（不可更改）', readOnly: true },
        { name: 'group_id', label: '要删除的完整同步组 ID', required: true },
        { name: 'confirm_delete', label: '整组删除与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份整组，了解移除全部流/管道及 Realm 发布风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupSyncDeleteInput(values,row) })
    }, {
      title: '修改 Zonegroup 同步组状态', path: '/rgw/zonegroup/sync/group', method: 'PATCH',
      successMessage: 'Zonegroup 策略与适用的 Period 发布已回读核验（不代表远端同步完成）',
      disabledWhen: zonegroupSyncBlocked, initialValues: zonegroupSyncInitial, confirmation: zonegroupSyncConfirmation,
      fields: [
        { name: 'name', label: 'Zonegroup 名称（不可更改）', readOnly: true },
        { name: 'zonegroup_id', label: 'Zonegroup ID（不可更改）', readOnly: true },
        { name: 'realm_id', label: 'Realm ID（空表示无 Realm，不可更改）', readOnly: true },
        { name: 'group_id', label: '已有同步组 ID', required: true },
        { name: 'status', label: '目标状态', type: 'select', required: true, options: [{ value: 'enabled', label: 'enabled 启用' }, { value: 'allowed', label: 'allowed 允许但不启用' }, { value: 'forbidden', label: 'forbidden 禁止' }] },
        { name: 'confirm_change', label: '修改与发布确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份并核对 Realm 全部待发布变更，了解部分生效风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...zonegroupSyncInput(values,row) })
    }],
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
      { key: 'placement_targets', title: '放置目标与存储类', ellipsis: false, render: value => <RgwPlacementClasses value={value} /> },
      { key: 'default_placement', title: '默认放置目标' },
      { key: 'hostnames', title: '主机名' },
      { key: 'hostnames_s3website', title: '静态网站主机名' },
      { key: 'sync_policy', title: 'Zonegroup 同步策略（采集时）', ellipsis: false, render: (value) => <div>{rgwBucketSyncPolicy(value, 'zonegroup')}<details><summary>查看 Zonegroup 数据流</summary><RgwBucketSyncFlows value={value} scope="zonegroup" /></details><details><summary>查看 Zonegroup 管道</summary><RgwBucketSyncPipes value={value} scope="zonegroup" /></details>{value != null && <details><summary>查看 Zonegroup 原生策略</summary><pre>{JSON.stringify(value, null, 2)}</pre></details>}</div> },
      { key: 'enabled_features', title: '启用特性' },
      { key: 'resource_version', title: '版本' }
    ]
  },
  rgwZones: {
    extraActions: [{
      title: '新增 Zone 放置池配置', path: '/rgw/zone/placement', method: 'POST',
      successMessage: 'Zone 新放置配置与适用的 Period 发布已核验（未迁移数据）',
      disabledWhen: zonePlacementCreateBlocked, confirmation: zonePlacementCreateConfirmation,
      initialValues: () => ({data_extra_pool:'',compression:'none'}),
      changedValues: changed => Object.keys(changed).some(key => key !== 'confirm_placement') ? {confirm_placement:undefined} : {},
      fields: [
        { name: 'zonegroup_id', label: '所属 Zonegroup', type: 'select', required: true, optionsLoader: async (_clusterId,row) => zonePlacementGroups(row) },
        { name: 'placement_id', label: '组中已声明、此 Zone 尚未配置的目标名称（STANDARD）', required: true },
        { name: 'index_pool', label: '索引池引用', required: true },
        { name: 'data_pool', label: 'STANDARD 数据池引用', required: true },
        { name: 'data_extra_pool', label: '额外数据池（显式留空使用原生回退）' },
        { name: 'compression', label: '压缩算法', type: 'select', required: true, options: zonePlacementCompressions.map(value => ({value,label:value})) },
        { name: 'confirm_placement', label: '范围确认', type: 'select', required: true, options: [{value:'acknowledged',label:'已备份并核对组声明、池和全部相关 Zone，了解 Period 发布范围'}] }
      ],
      buildBody: (values,clusterId,row) => ({cluster_id:clusterId,...zonePlacementCreateInput(values,row)})
    }, {
      title: '新增 Zone 存储类', path: '/rgw/zone/storage/class', method: 'POST',
      successMessage: 'Zone 存储类新增已回读核验（未发布 Period 或迁移数据）',
      disabledWhen: zonePlacementBlocked, confirmation: zoneStorageClassConfirmation,
      changedValues: changed => Object.keys(changed).some(key => key !== 'confirm_placement') ? {confirm_placement:undefined} : {},
      fields: [
        { name: 'zonegroup_id', label: '所属 Zonegroup', type: 'select', required: true, optionsLoader: async (_clusterId, row) => zonePlacementGroups(row) },
        { name: 'placement_id', label: '已有放置目标', type: 'select', required: true, optionsLoader: async (_clusterId, row) => zonePlacementOptions(row) },
        { name: 'storage_class', label: '新存储类（须已在组目标声明，且此 Zone 尚未配置）', required: true },
        { name: 'data_pool', label: '数据池引用（可含命名空间）', required: true },
        { name: 'compression', label: '压缩算法', type: 'select', required: true, options: zonePlacementCompressions.map(value => ({value,label:value})) },
        { name: 'confirm_placement', label: '风险确认', type: 'select', required: true, options: [{value:'acknowledged',label:'已备份并核对组声明和池配置，了解不迁移数据或发布 Period'}] }
      ],
      buildBody: (values, clusterId, row) => ({cluster_id:clusterId,...zoneStorageClassInput(values,row)})
    }, {
      title: '编辑放置池与压缩', path: '/rgw/zone/placement', method: 'PATCH',
      successMessage: 'Zone 放置配置已回读核验（未迁移数据或重启网关）',
      disabledWhen: zonePlacementBlocked, confirmation: zonePlacementConfirmation,
      changedValues: zonePlacementChanged,
      fields: [
        { name: 'zonegroup_id', label: '所属 Zonegroup', type: 'select', required: true, optionsLoader: async (_clusterId, row) => zonePlacementGroups(row) },
        { name: 'placement_id', label: '已有放置目标', type: 'select', required: true, optionsLoader: async (_clusterId, row) => zonePlacementOptions(row) },
        { name: 'storage_class', label: '已有存储类', type: 'select', required: true, optionsDependencies: ['placement_id'], optionsLoader: async (_clusterId, row, values) => zonePlacementClasses(row, values?.placement_id) },
        { name: 'index_pool', label: '索引池（影响整个放置目标）', required: true },
        { name: 'data_pool', label: '所选存储类数据池', required: true },
        { name: 'data_extra_pool', label: '额外数据池（留空使用原生回退；影响整个放置目标）' },
        { name: 'compression', label: '压缩算法', type: 'select', required: true, options: zonePlacementCompressions.map(value => ({value,label:value})) },
        { name: 'confirm_placement', label: '风险确认', type: 'select', required: true, options: [{value:'acknowledged',label:'已备份并评估更换池风险，了解不迁移数据及 Period 发布范围'}] }
      ],
      buildBody: (values, clusterId, row) => ({cluster_id:clusterId,...zonePlacementInput(values,row)})
    }],
    deleteAction: {
      title: '删除 Zone（保留池）', path: '/rgw/zone', action: 'rgw_zone.delete', resourceKind: 'rgw_zone', risk: 'high',
      disabledWhen: zoneDeleteBlocked, confirmation: zoneDeleteConfirmation,
      resourceKey: row => String(row.name),
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...zoneDeleteInput(row) }),
      successMessage: 'Zone 删除、成员关系及适用的 Period 发布已核验；池和对象未删除'
    },
    title: 'RGW Zones',
    path: '/rgw/zones',
    requiredCapabilities: ['rgw_admin'],
    detailContent: (row, clusterId) => <RgwSyncStatus key={`${clusterId}:${row.id}:${row.name}`} row={row} clusterId={clusterId} />,
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
      { key: 'placement_pools', title: '放置池与存储类别', ellipsis: false, render: value => <RgwPlacementClasses value={value} zone /> },
      { key: 'pool_references', title: '精确池引用明细', ellipsis: false, render: (_value, row) => <RgwZonePoolReferences row={row} /> },
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
    title: 'Bucket 配置文档',
    path: '/rgw/bucket/policy',
    requiredCapabilities: ['rgw_admin'],
    requiredEndpoints: ['s3'],
    rowKeyCandidates: ['bucket_id', 'name', 'kind'],
    body: { kind: 'policy' },
    buildQuery: (body) => new URLSearchParams({ kind: String(body.kind ?? 'policy') }),
    filterFields: [
      { name: 'bucket_id', label: 'Bucket ID', required: true },
      { name: 'kind', label: '配置类型', type: 'select', required: true, options: rgwBucketConfigurationReadOptions }
    ],
    createAction: {
      title: '更新 Bucket 配置文档',
      buttonLabel: '更新配置',
      path: '/rgw/bucket/policy',
      method: 'PATCH',
      successMessage: 'Bucket 配置文档提交成功',
      confirmation: rgwBucketConfigurationUpdateConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID', required: true },
        {
          name: 'kind',
          label: '配置类型',
          type: 'select',
          required: true,
          options: rgwBucketConfigurationOptions
        },
        { name: 'document', label: '完整配置文档（Policy 为 JSON，其它为 XML；提交将替换该配置）', type: 'textarea', required: true }
      ],
      initialValues: { kind: 'policy' },
      buildBody: (values, clusterId) => ({ cluster_id: clusterId, ...rgwBucketConfigurationInput(values) })
    },
    updateAction: {
      title: '编辑当前 Bucket 配置文档',
      path: '/rgw/bucket/policy',
      method: 'PATCH',
      successMessage: 'Bucket 配置文档提交成功',
      disabledWhen: rgwBucketConfigurationEditBlocked,
      initialValues: rgwBucketConfigurationEditInitial,
      confirmation: rgwBucketConfigurationUpdateConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'kind', label: '配置类型（不可更改）', readOnly: true },
        { name: 'document', label: '完整配置文档（Policy 为 JSON，其它为 XML；提交将替换该配置）', type: 'textarea', required: true }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...rgwBucketConfigurationEditInput(values, row) })
    },
    deleteAction: {
      title: '删除 Bucket 配置',
      path: '/rgw/bucket/policy',
      action: 'rgw_bucket_policy.delete',
      resourceKind: 'rgw_bucket_policy',
      risk: 'high',
      successMessage: 'Bucket 配置删除已回读核验（复制配置允许返回空规则）',
      disabledWhen: rgwBucketConfigurationDeleteBlocked,
      confirmation: rgwBucketConfigurationDeleteConfirmation,
      resourceKey: (row) => `${rgwBucketConfigurationDeleteInput(row).bucket_id} / ${row.kind}`,
      buildBody: (row, clusterId) => ({ cluster_id: clusterId, ...rgwBucketConfigurationDeleteInput(row) })
    },
    extraActions: [{
      title: '设置 MFA Delete 与版本控制', buttonLabel: '设置 MFA Delete', path: '/rgw/bucket/mfa', method: 'PATCH',
      successMessage: '版本控制与 MFA Delete 已回读核验', visibleWhen: row => row.kind === 'versioning',
      disabledWhen: bucketMFABlocked, initialValues: bucketMFAInitial, confirmation: bucketMFAConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'status', label: '版本控制', type: 'select', required: true, options: [{ value: 'Enabled', label: '启用' }, { value: 'Suspended', label: '暂停（不是关闭）' }] },
        { name: 'mfa_delete', label: 'MFA Delete', type: 'select', required: true, options: [{ value: 'Enabled', label: '启用 MFA 删除保护' }, { value: 'Disabled', label: '停用 MFA 删除保护' }] },
        { name: 'mfa_serial_secret', label: '当前 S3 用户已绑定的 MFA 设备序列号', type: 'password', required: true },
        { name: 'mfa_token', label: '当前验证码（保留前导零）', type: 'password', required: true },
        { name: 'confirm_mfa', label: '保护及版本控制确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '确认 MFA 保护变化与验证码过期风险，失败先核对再重试' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketMFAInput(values, row) })
    }, {
      title: '创建或编辑 Bucket 通知', buttonLabel: '创建 / 编辑通知', path: '/rgw/bucket/notification', method: 'POST',
      successMessage: '通知配置已回读核验（不代表消息投递成功）',
      visibleWhen: row => row.kind === 'notification', disabledWhen: notificationFormBlocked, initialValues: notificationFormInitial, confirmation: notificationFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'notification_draft', label: '通知配置', required: true, renderControl: disabled => <RgwBucketNotificationEditor disabled={disabled} /> },
        { name: 'confirm_notification', label: '非原子修改确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份配置，接受更换 Topic 的先删后建、通知空窗及失败无自动回滚' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...notificationFormInput(values, row) })
    }, {
      title: '删除 Bucket 通知规则', buttonLabel: '删除通知',
      path: '/rgw/bucket/notification', method: 'DELETE',
      successMessage: '通知删除已回读核验（不代表队列或 Topic 映射已清理）',
      visibleWhen: (row) => row.kind === 'notification',
      disabledWhen: bucketNotificationDeleteBlocked, initialValues: bucketNotificationDeleteInitial, confirmation: bucketNotificationDeleteConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'mode', label: '删除范围', type: 'select', required: true, options: [{ value: 'single', label: '指定唯一通知 ID' }, { value: 'all', label: '全部通知规则' }] },
        { name: 'notification_id', label: '通知 ID（精确匹配；全部删除时必须留空）' },
        { name: 'confirm_notification', label: '停止通知确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已备份配置，确认删除规则且没有自动撤销' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketNotificationDeleteInput(values, row) })
    }, {
      title: '设置 Dashboard 桶复制规则', buttonLabel: '设置同名桶复制',
      path: '/rgw/bucket/replication', method: 'POST',
      successMessage: 'S3 复制规则已回读核验（不代表复制运行或完成）',
      visibleWhen: (row) => row.kind === 'replication',
      disabledWhen: bucketReplicationFormBlocked, initialValues: bucketReplicationFormInitial, confirmation: bucketReplicationFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'confirm_replication', label: '上层策略与覆盖确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '已准备上层策略并备份规则，确认以全部 Zone 的同名桶规则替换全部 S3 复制规则' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketReplicationFormInput(values, row) })
    }, {
      title: '修改对象锁默认保留', buttonLabel: '修改默认保留',
      path: '/rgw/bucket/policy', method: 'PATCH',
      successMessage: '对象锁默认保留配置已回读核验',
      visibleWhen: (row) => row.kind === 'object-lock',
      disabledWhen: objectLockFormBlocked, initialValues: objectLockFormInitial, confirmation: objectLockFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'kind', label: '配置类型（不可更改）', readOnly: true },
        { name: 'retention_action', label: '默认保留策略操作', type: 'select', required: true, options: [{ value: 'set', label: '设置默认保留期' }, { value: 'clear', label: '清除默认保留期（不关闭对象锁）' }] },
        { name: 'mode', label: '模式（仅设置操作使用）', type: 'select', options: [{ value: 'GOVERNANCE', label: 'GOVERNANCE 治理模式' }, { value: 'COMPLIANCE', label: 'COMPLIANCE 合规模式' }] },
        { name: 'unit', label: '单位（仅设置操作使用）', type: 'select', options: [{ value: 'Days', label: '天' }, { value: 'Years', label: '年' }] },
        { name: 'period', label: '正整数保留期（仅设置操作使用）' },
        { name: 'confirm_lock', label: '不可逆影响确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认对象锁无法关闭，保留策略可能阻止对象删除' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...objectLockFormInput(values, row) })
    }, {
      title: '替换 Bucket ACL', buttonLabel: '修改 ACL 预设',
      path: '/rgw/bucket/acl', method: 'PATCH',
      successMessage: 'Bucket ACL 已提交并回读核验',
      visibleWhen: (row) => row.kind === 'acl',
      disabledWhen: bucketAclFormBlocked, initialValues: bucketAclFormInitial, confirmation: bucketAclFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'acl', label: '完整 ACL 预设（覆盖现有授权）', type: 'select', required: true, options: bucketAclOptions },
        { name: 'confirm_replace', label: '访问权限风险确认', type: 'select', required: true, options: [{ value: 'acknowledged', label: '我确认覆盖所有现有 ACL 授权，并理解所选预设的公开访问风险' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketAclFormInput(values, row) })
    }, {
      title: '编辑生命周期规则',
      buttonLabel: '编辑生命周期',
      path: '/rgw/bucket/policy', method: 'PATCH',
      successMessage: '生命周期已提交并回读核验（不代表对象处理完成）',
      visibleWhen: (row) => row.kind === 'lifecycle',
      disabledWhen: lifecycleFormBlocked, initialValues: lifecycleFormInitial, confirmation: lifecycleFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'kind', label: '配置类型（不可更改）', readOnly: true },
        { name: 'lifecycle_draft', label: '完整生命周期规则', required: true, renderControl: (disabled) => <RgwBucketLifecycleEditor disabled={disabled} /> }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...lifecycleFormInput(values, row) })
    }, {
      title: '编辑 CORS 规则',
      buttonLabel: '编辑 CORS',
      path: '/rgw/bucket/policy', method: 'PATCH',
      successMessage: 'CORS 已提交并回读核验',
      visibleWhen: (row) => row.kind === 'cors',
      disabledWhen: corsFormBlocked, initialValues: corsFormInitial, confirmation: corsFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'kind', label: '配置类型（不可更改）', readOnly: true },
        { name: 'cors_draft', label: '完整 CORS 规则（保留顺序）', required: true, renderControl: (disabled) => <RgwBucketCorsEditor disabled={disabled} /> }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...corsFormInput(values, row) })
    }, {
      title: '编辑 Bucket 默认加密',
      buttonLabel: '编辑加密',
      path: '/rgw/bucket/policy',
      method: 'PATCH',
      successMessage: '默认加密配置提交成功并已回读核验',
      visibleWhen: (row) => row.kind === 'encryption',
      disabledWhen: bucketEncryptionFormBlocked,
      initialValues: bucketEncryptionFormInitial,
      confirmation: bucketEncryptionFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'kind', label: '配置类型（不可更改）', readOnly: true },
        { name: 'algorithm', label: '默认加密算法（需先配置 RGW 加密服务）', type: 'select', required: true, options: [{ label: 'SSE-S3 / AES256', value: 'AES256' }, { label: 'SSE-KMS / aws:kms', value: 'aws:kms' }] },
        { name: 'kms_master_key_id', label: 'KMS Key ID（KMS 必填，AES256 请清空）', type: 'textarea' },
        { name: 'bucket_key_enabled', label: 'Bucket Key 配置值', type: 'select', required: true, options: [{ label: '启用', value: 'true' }, { label: '不启用', value: 'false' }] }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketEncryptionFormInput(values, row) })
    }, {
      title: '逐条编辑 Bucket 标签',
      buttonLabel: '编辑标签',
      path: '/rgw/bucket/policy',
      method: 'PATCH',
      successMessage: 'Bucket 标签提交成功并已回读核验',
      visibleWhen: (row) => row.kind === 'tagging',
      disabledWhen: bucketTagFormBlocked,
      initialValues: bucketTagFormInitial,
      confirmation: bucketTagFormConfirmation,
      fields: [
        { name: 'bucket_id', label: 'Bucket ID（不可更改）', readOnly: true },
        { name: 'kind', label: '配置类型（不可更改）', readOnly: true },
        { name: 'tag_set', label: '完整标签集合', required: true, renderControl: (disabled) => <RgwBucketTagEditor disabled={disabled} /> }
      ],
      buildBody: (values, clusterId, row) => ({ cluster_id: clusterId, ...bucketTagFormInput(values, row) })
    }],
    columns: [
      { key: 'bucket_id', title: 'Bucket ID' },
      { key: 'kind', title: '配置类型' },
      { key: 'configured', title: '配置状态', render: (value) => value === true ? '已配置' : value === false ? '未配置' : '状态不可用' },
      { key: 'versioning', title: '版本控制 / MFA', ellipsis: false, render: bucketVersioningSummary },
      { key: 'tags', title: '标签条目', ellipsis: false, render: (value, row) => row.kind === 'tagging' ? <RgwBucketTagEntries value={value} /> : '—' },
      { key: 'encryption', title: '默认加密', ellipsis: false, render: rgwBucketEncryptionSummary },
      { key: 'object_lock', title: '对象锁默认保留', ellipsis: false, render: rgwBucketObjectLockSummary },
      { key: 'acl', title: 'ACL 访问控制', ellipsis: false, render: (value, row) => row.kind === 'acl' ? <RgwBucketAcl value={value} configured={row.configured} /> : '—' },
      { key: 'replication', title: 'S3 复制规则', ellipsis: false, render: (value, row) => row.kind === 'replication' ? <RgwBucketReplication value={value} configured={row.configured} /> : '—' },
      { key: 'notifications', title: '事件通知规则', ellipsis: false, render: (value, row) => row.kind === 'notification' ? <RgwBucketNotifications value={value} configured={row.configured} /> : '—' },
      { key: 'cors_rules', title: 'CORS 规则', ellipsis: false, render: (value, row) => row.kind === 'cors' ? <RgwBucketCorsRules value={value} configured={row.configured} /> : '—' },
      { key: 'lifecycle_rules', title: '生命周期规则', ellipsis: false, render: (value, row) => row.kind === 'lifecycle' ? <RgwBucketLifecycleRules value={value} configured={row.configured} /> : '—' },
      { key: 'content_type', title: '响应类型' },
      { key: 'document', title: '原始配置文档', ellipsis: false, render: (value, row) => <pre style={{ whiteSpace: 'pre-wrap', maxHeight: 240, overflow: 'auto' }}>{row.configured === false ? '未配置，无配置文档' : typeof value === 'string' ? value : '配置文档不可用'}</pre> }
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

function text(value: unknown) {
  return value === null || value === undefined ? '' : String(value)
}

function numberOrUndefined(value: unknown) {
  return typeof value === 'number' && Number.isInteger(value) && value >= -2147483648 && value <= 2147483647
    ? value : undefined
}
