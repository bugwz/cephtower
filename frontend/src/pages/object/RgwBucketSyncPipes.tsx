import { Table } from 'antd'

type ObjectValue = Record<string, unknown>
function record(value: unknown): value is ObjectValue {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
const raw = (value: unknown) => value === undefined ? '未返回' : JSON.stringify(value)
function zones(value: unknown): string {
  if (value === undefined) return '未返回 Zone 选择（不推断全部 Zone）'
  if (!Array.isArray(value) || value.some(zone => typeof zone !== 'string')) return 'Zone 列表格式不可用'
  if (!value.length) return 'Zone 列表为空'
  if (value.length === 1 && value[0] === '*') return '*（所有 Zone，仍受数据流约束）'
  return value.map(raw).join(' / ')
}
type PipeRow = {
  index: number; id: string; sourceBucket: string; destBucket: string; sourceZones: string; destZones: string;
  params: unknown
}
type GroupPipes = { id: string; rows: PipeRow[]; unavailable: boolean }
export function bucketSyncPipes(value: unknown): GroupPipes[] | undefined {
  if (!record(value) || !Array.isArray(value.groups)) return undefined
  const seen = new Set<string>()
  const result: GroupPipes[] = []
  for (const group of value.groups) {
    if (!record(group) || typeof group.id !== 'string' || seen.has(group.id)) return undefined
    seen.add(group.id)
    const item: GroupPipes = { id: group.id, rows: [], unavailable: false }
    result.push(item)
    if (!Array.isArray(group.pipes)) { item.unavailable = true; continue }
    const pipeIDs = new Set<string>()
    for (const pipe of group.pipes) {
      if (!record(pipe) || typeof pipe.id !== 'string' || !record(pipe.source) || !record(pipe.dest)
        || typeof pipe.source.bucket !== 'string' || typeof pipe.dest.bucket !== 'string') { item.unavailable = true; continue }
      if (pipeIDs.has(pipe.id)) item.unavailable = true
      pipeIDs.add(pipe.id)
      item.rows.push({
        index: item.rows.length, id: raw(pipe.id), sourceBucket: raw(pipe.source.bucket), destBucket: raw(pipe.dest.bucket),
        sourceZones: zones(pipe.source.zones), destZones: zones(pipe.dest.zones), params: pipe.params
      })
    }
  }
  return result
}
export function bucketPipeParameters(value: unknown): string {
  if (!record(value)) return '管道参数不可用'
  const source = record(value.source) ? value.source : {}
  const filter = record(source.filter) ? source.filter : {}
  const dest = record(value.dest) ? value.dest : {}
  const mode = value.mode === 'system' ? 'system（系统模式）' : value.mode === 'user' ? 'user（用户模式）' : `未知/未返回：${raw(value.mode)}`
  const priority = typeof value.priority === 'number' && Number.isSafeInteger(value.priority) ? String(value.priority) : `不可用或非精确整数：${raw(value.priority)}`
  return [
    `模式：${mode}`, `用户：${raw(value.user)}`, `优先级：${priority}`,
    `源前缀：${raw(filter.prefix)}`, `源标签：${raw(filter.tags)}`,
    `目标 ACL 转换：${raw(dest.acl_translation)}`, `目标存储类：${raw(dest.storage_class)}`
  ].join('\n')
}
export function RgwBucketSyncPipes({ value }: { value: unknown }) {
  const groups = bucketSyncPipes(value)
  if (!groups) return <span>桶本地管道配置不可用</span>
  return <div>
    <p>采集时的桶本地管道配置，不代表有效复制链路、访问权限或同步进度。桶键按原生值显示，* 为通配选择；Zone 显示值可能是名称或未解析的 ID。</p>
    {!groups.length && <p>无桶本地同步组（不推断继承策略）</p>}
    {groups.map(group => <section key={group.id}>
      <h4>同步组 {raw(group.id)}</h4>
      {group.unavailable && <p>部分管道数据不可用或 ID 重复；以下仅展示可识别项，请检查原生策略。</p>}
      <Table size="small" rowKey="index" dataSource={group.rows} pagination={group.rows.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 900 }}
        locale={{ emptyText: group.unavailable ? '无可识别管道（不推断为空配置）' : '此组无桶本地管道' }}
        columns={[
          { title: '管道 ID', dataIndex: 'id' }, { title: '源 Zone', dataIndex: 'sourceZones' },
          { title: '目标 Zone', dataIndex: 'destZones' }, { title: '源桶键', dataIndex: 'sourceBucket' },
          { title: '目标桶键', dataIndex: 'destBucket' },
          { title: '过滤与权限参数', dataIndex: 'params', render: (params: unknown) => <details><summary>查看参数</summary><pre>{bucketPipeParameters(params)}</pre><pre>{raw(params)}</pre></details> }
        ]} />
    </section>)}
  </div>
}
