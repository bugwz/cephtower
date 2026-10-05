import { Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'
import { rgwIdentityText } from './rgwUserIdentity'
import { rgwBucketBooleanState, rgwBucketVersioning } from './rgwBucketState'

export function RgwBucketSummary({ row }: { row: ApiRecord }) {
  return <Descriptions size="small" column={1} items={[
    { key: 'name', label: 'Bucket 名称', children: rgwIdentityText(row.name, '空名称') },
    { key: 'tenant', label: '租户', children: rgwIdentityText(row.tenant, '默认租户') },
    { key: 'id', label: '原生 Bucket ID', children: rgwIdentityText(row.id, '空 ID') },
    { key: 'owner', label: 'Owner（原生标识）', children: rgwIdentityText(row.owner, '空 Owner') },
    { key: 'versioning', label: '版本控制', children: rgwBucketVersioning(row.versioning) },
    { key: 'mfa', label: 'MFA Delete', children: rgwBucketBooleanState(row.mfa_enabled) },
    { key: 'lock', label: '对象锁启用标记（非保留策略）', children: rgwBucketBooleanState(row.object_lock_enabled) },
    { key: 'zonegroup', label: 'Zonegroup ID', children: rgwIdentityText(row.zonegroup, '未显式设置') },
    { key: 'placement', label: '放置规则', children: rgwIdentityText(row.placement_rule, '未显式设置') },
    { key: 'created', label: '创建时间（命令原值）', children: rgwIdentityText(row.creation_time, '未提供时间') },
    { key: 'modified', label: '修改时间（命令原值）', children: rgwIdentityText(row.mtime, '未提供时间') }
  ]} />
}
