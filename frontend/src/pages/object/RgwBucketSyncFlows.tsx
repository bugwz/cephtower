import { Table } from 'antd'

type RecordValue = Record<string, unknown>
function record(value: unknown): value is RecordValue {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}
type Flow = { index: number; kind: string; id: string; source: string; destination: string; zones: string }
type GroupFlows = { id: string; rows: Flow[]; unavailable: boolean; extensions: string[] }
const exact = (value: string) => JSON.stringify(value)
export function bucketSyncFlows(value: unknown): GroupFlows[] | undefined {
  if (!record(value) || !Array.isArray(value.groups)) return undefined
  const seen = new Set<string>()
  const result: GroupFlows[] = []
  for (const group of value.groups) {
    if (!record(group) || typeof group.id !== 'string' || seen.has(group.id)) return undefined
    seen.add(group.id)
    const item: GroupFlows = { id: group.id, rows: [], unavailable: false, extensions: [] }
    result.push(item)
    if (!record(group.data_flow)) { item.unavailable = true; continue }
    const data = group.data_flow
    item.extensions = Object.keys(data).filter(key => key !== 'symmetrical' && key !== 'directional')
    for (const kind of ['symmetrical', 'directional']) {
      if (!Object.prototype.hasOwnProperty.call(data, kind)) continue
      const entries = data[kind]
      if (!Array.isArray(entries)) { item.unavailable = true; continue }
      for (const entry of entries) {
        if (!record(entry)) { item.unavailable = true; continue }
        if (kind === 'symmetrical') {
          if (typeof entry.id !== 'string' || !Array.isArray(entry.zones) || entry.zones.some(zone => typeof zone !== 'string')) {
            item.unavailable = true; continue
          }
          item.rows.push({ index: item.rows.length, kind: '对称', id: exact(entry.id), source: '不适用', destination: '不适用', zones: entry.zones.length ? entry.zones.map(exact).join(' / ') : 'Zone 列表为空' })
        } else {
          if (typeof entry.source_zone !== 'string' || typeof entry.dest_zone !== 'string') { item.unavailable = true; continue }
          item.rows.push({ index: item.rows.length, kind: '定向', id: '原生数据无 ID', source: exact(entry.source_zone), destination: exact(entry.dest_zone), zones: '不适用' })
        }
      }
    }
  }
  return result
}

export function RgwBucketSyncFlows({ value }: { value: unknown }) {
  const groups = bucketSyncFlows(value)
  if (!groups) return <span>桶本地数据流不可用</span>
  return <div>
    <p>采集时的桶本地数据流配置；Zone 使用原生标识，不推断名称。是否实际复制仍取决于组状态、管道和上层策略，不代表运行状态或同步进度。</p>
    {!groups.length && <p>无桶本地同步组（不代表不存在继承的复制策略）</p>}
    {groups.map(group => <section key={group.id}>
      <h4>同步组 {exact(group.id)}</h4>
      {group.unavailable && <p>部分数据流格式不可用；以下仅展示可识别项，请检查原生策略。</p>}
      {group.extensions.length > 0 && <p>存在未识别的数据流字段：{group.extensions.map(exact).join(' / ')}；完整内容见原生策略。</p>}
      <Table size="small" rowKey="index" dataSource={group.rows} pagination={group.rows.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 650 }}
        locale={{ emptyText: group.unavailable || group.extensions.length ? '无可识别数据流（不推断为空配置）' : '此组无桶本地数据流' }}
        columns={[
          { title: '类型', dataIndex: 'kind' }, { title: '数据流 ID', dataIndex: 'id' },
          { title: '对称 Zone 列表', dataIndex: 'zones' }, { title: '源 Zone', dataIndex: 'source' },
          { title: '目标 Zone', dataIndex: 'destination' }
        ]} />
    </section>)}
  </div>
}
