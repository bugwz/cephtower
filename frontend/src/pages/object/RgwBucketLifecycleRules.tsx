import { Table } from 'antd'

const actionNames: Record<string, string> = {
  Expiration: '当前版本过期', NoncurrentVersionExpiration: '非当前版本过期',
  AbortIncompleteMultipartUpload: '终止未完成分段上传', Transition: '当前版本转换',
  NoncurrentVersionTransition: '非当前版本转换'
}
const fieldNames: Record<string, string> = {
  Days: '天数', NoncurrentDays: '成为非当前版本后的天数', DaysAfterInitiation: '上传发起后的天数',
  NewerNoncurrentVersions: '保留较新非当前版本数量', Date: 'UTC 日期', StorageClass: '目标存储类别',
  ExpiredObjectDeleteMarker: '清除过期删除标记'
}
type Selector = { kind: string; and: boolean; prefix: string | null; tags: { key: string; value: string }[]; object_size_greater_than: string | null; object_size_less_than: string | null; archive_zone: boolean }
type Action = { type: string; fields: Record<string, string> }
type Rule = { id: string; status: string; selector: Selector; actions: Action[] }
const record = (value: unknown): value is Record<string, unknown> => !!value && typeof value === 'object' && !Array.isArray(value)

export function bucketLifecycleRows(value: unknown): (Rule & { index: number })[] | undefined {
  if (!Array.isArray(value)) return undefined
  for (const rule of value) {
    if (!record(rule) || typeof rule.id !== 'string' || typeof rule.status !== 'string' || !['Enabled', 'Disabled'].includes(rule.status)) return undefined
    const selector = rule.selector
    if (!record(selector) || typeof selector.kind !== 'string' || !['Prefix', 'Filter'].includes(selector.kind) || typeof selector.and !== 'boolean' || typeof selector.archive_zone !== 'boolean') return undefined
    if (['prefix', 'object_size_greater_than', 'object_size_less_than'].some(key => selector[key] !== null && typeof selector[key] !== 'string')) return undefined
    if (!Array.isArray(selector.tags) || selector.tags.some(tag => !record(tag) || typeof tag.key !== 'string' || typeof tag.value !== 'string')) return undefined
    if (!Array.isArray(rule.actions) || !rule.actions.length || rule.actions.some(action => !record(action) || typeof action.type !== 'string' || !Object.prototype.hasOwnProperty.call(actionNames, action.type) || !record(action.fields) || !Object.keys(action.fields).length || Object.entries(action.fields).some(([key, item]) => !Object.prototype.hasOwnProperty.call(fieldNames, key) || typeof item !== 'string'))) return undefined
  }
  return (value as Rule[]).map((rule, index) => ({ ...rule, index }))
}

export function lifecycleSelectorText(selector: Selector): string {
  const conditions: string[] = []
  if (selector.prefix !== null) conditions.push(`前缀：${JSON.stringify(selector.prefix)}`)
  for (const tag of selector.tags) conditions.push(`标签：${JSON.stringify(tag.key)} = ${JSON.stringify(tag.value)}`)
  if (selector.object_size_greater_than !== null) conditions.push(`对象字节数 > ${selector.object_size_greater_than === '' ? '（未设置）' : selector.object_size_greater_than}`)
  if (selector.object_size_less_than !== null) conditions.push(`对象字节数 < ${selector.object_size_less_than === '' ? '（未设置）' : selector.object_size_less_than}`)
  if (selector.archive_zone) conditions.push('归档区域条件：存在')
  return `${selector.kind === 'Prefix' ? '旧式 Prefix' : selector.and ? 'Filter / And' : 'Filter'}\n${conditions.length ? conditions.join('\n') : '无显式过滤条件'}`
}

export function lifecycleActionsText(actions: Action[]): string {
  return actions.map(action => `${actionNames[action.type]}\n${Object.entries(action.fields).map(([key, value]) => `${fieldNames[key]}：${JSON.stringify(value)}`).join('\n')}`).join('\n\n')
}

export function RgwBucketLifecycleRules({ value, configured }: { value: unknown; configured: unknown }) {
  const rules = bucketLifecycleRows(value)
  if (!rules || typeof configured !== 'boolean' || (configured && !rules.length) || (!configured && rules.length > 0)) return <span>生命周期数据不可用</span>
  if (!configured) return <span>未配置生命周期</span>
  return <Table size="small" rowKey="index" dataSource={rules} pagination={rules.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 750 }} columns={[
    { title: '序号', dataIndex: 'index', render: (index: number) => index + 1 },
    { title: '规则 ID', dataIndex: 'id', render: (id: string) => JSON.stringify(id) },
    { title: '状态', dataIndex: 'status', render: (status: string) => status === 'Enabled' ? '启用' : '禁用' },
    { title: '过滤条件', dataIndex: 'selector', render: (selector: Selector) => <span style={{ whiteSpace: 'pre-wrap' }}>{lifecycleSelectorText(selector)}</span> },
    { title: '动作', dataIndex: 'actions', render: (actions: Action[]) => <span style={{ whiteSpace: 'pre-wrap' }}>{lifecycleActionsText(actions)}</span> }
  ]} />
}
