import { Descriptions } from 'antd'
import type { ApiRecord } from '../../api/client'
import { rgwIdentityText } from './rgwUserIdentity'

export function RgwRoleSummary({ row }: { row: ApiRecord }) {
  const duration = row.MaxSessionDuration
  return <Descriptions size="small" column={1} items={[
    { key: 'name', label: '角色名称（含租户前缀）', children: rgwIdentityText(row.RoleName, '空名称') },
    { key: 'id', label: '原生 Role ID', children: rgwIdentityText(row.RoleId, '空 ID') },
    { key: 'account', label: '账户 ID', children: rgwIdentityText(row.AccountId, '未关联账户（租户作用域角色）') },
    { key: 'path', label: 'Path', children: rgwIdentityText(row.Path, '空路径') },
    { key: 'arn', label: 'ARN', children: rgwIdentityText(row.Arn, '空 ARN') },
    { key: 'description', label: '描述', children: rgwIdentityText(row.Description, '未设置描述') },
    { key: 'duration', label: '最大会话时长（秒，非当前会话剩余时间）', children: typeof duration === 'number' && Number.isSafeInteger(duration) && duration >= 3600 && duration <= 43200 ? String(duration) : '未返回或格式无效' },
    { key: 'created', label: '创建时间（命令原值）', children: rgwIdentityText(row.CreateDate, '未提供时间') }
  ]} />
}
