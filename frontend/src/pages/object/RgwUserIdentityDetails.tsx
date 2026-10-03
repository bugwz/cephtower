import { Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'
import { rgwIdentityList, rgwIdentityText } from './rgwUserIdentity'
import { RgwUserTagsTable } from './RgwUserTagsTable'

function Identifiers({ value }: { value: unknown }) {
  const ids = rgwIdentityList(value)
  if (!ids) return <span>未返回或格式无效</span>
  if (ids.length === 0) return <span>未配置</span>
  return <ul style={{ margin: 0, paddingLeft: 18 }}>{ids.map((id, index) => <li key={index}>{id}</li>)}</ul>
}

export function RgwUserIdentityDetails({ row }: { row: ApiRecord }) {
  return <Descriptions size="small" column={1} items={[
    { key: 'full-uid', label: '完整 UID（命令原值）', children: rgwIdentityText(row.full_user_id, '空 UID') },
    { key: 'local-id', label: '本地用户 ID', children: rgwIdentityText(row.user_id, '空用户 ID') },
    { key: 'tenant', label: '租户', children: rgwIdentityText(row.tenant, '默认租户') },
    { key: 'namespace', label: '用户命名空间', children: row.namespace === undefined ? '未返回（原生命令在命名空间为空时省略）' : rgwIdentityText(row.namespace, '空命名空间') },
    { key: 'account', label: '账户 ID', children: rgwIdentityText(row.account_id, '未关联账户') },
    { key: 'type', label: '用户类型（命令原值）', children: rgwIdentityText(row.type, '空类型') },
    { key: 'path', label: '用户路径', children: rgwIdentityText(row.path, '空路径') },
    { key: 'created', label: '创建时间（命令原值）', children: rgwIdentityText(row.create_date, '未提供时间') },
    { key: 'tags', label: '用户标签', children: <RgwUserTagsTable value={row.tags} /> },
    { key: 'policies', label: '直接关联的托管策略 ARN（非完整有效权限）', children: row.account_id === '' || row.type === 'root' ? '不适用' : <Identifiers value={row.managed_user_policies} /> },
    { key: 'mask', label: '操作掩码（命令原值）', children: rgwIdentityText(row.op_mask, '空操作掩码') },
    { key: 'mfa', label: 'MFA 标识', children: <Identifiers value={row.mfa_ids} /> },
    { key: 'groups', label: '用户组 ID', children: <Identifiers value={row.group_ids} /> }
  ]} />
}

export function RgwUserPlacementDetails({ row }: { row: ApiRecord }) {
  return <Descriptions size="small" column={1} items={[
    { key: 'placement', label: '默认放置规则', children: rgwIdentityText(row.default_placement, '未显式设置') },
    { key: 'storage', label: '默认存储类', children: rgwIdentityText(row.default_storage_class, '未显式设置') },
    { key: 'tags', label: '放置标签', children: <Identifiers value={row.placement_tags} /> }
  ]} />
}
