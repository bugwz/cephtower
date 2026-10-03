import { Table } from 'antd'

function record(value: unknown): value is Record<string, unknown> { return !!value && typeof value === 'object' && !Array.isArray(value) }
const raw = (v: unknown) => v === undefined ? '未返回' : JSON.stringify(v)
const epoch = (v: unknown) => typeof v === 'number' && Number.isSafeInteger(v) && v >= 0 ? String(v) : '不可用（非精确非负整数）'
export function currentPeriodSummary(value: unknown, realm: unknown, current: unknown) {
  if (!record(value) || typeof realm !== 'string' || !realm || typeof current !== 'string' || !current || value.realm_id !== realm || value.id !== current) return '当前 Period 数据不可用或归属不匹配'
  return `Period ${raw(value.id)}；epoch ${epoch(value.epoch)}；Realm epoch ${epoch(value.realm_epoch)}；前驱 ${raw(value.predecessor_uuid)}；主 Zonegroup ${raw(value.master_zonegroup)}；主 Zone ${raw(value.master_zone)}`
}
export function RgwCurrentPeriod({ value, realm, current }: { value: unknown; realm: unknown; current: unknown }) {
  const summary = currentPeriodSummary(value, realm, current)
  if (!record(value) || value.realm_id !== realm || value.id !== current || summary.startsWith('当前 Period 数据不可用')) return <span>{summary}</span>
  const map = record(value.period_map) ? value.period_map : undefined
  const groups = map?.zonegroups
  const valid = Array.isArray(groups) && groups.every(group => record(group) && typeof group.id === 'string' && typeof group.name === 'string') && new Set(groups.map(group => group.id)).size === groups.length
  return <div>
    <p>{summary}</p>
    <p>采集时 Realm 指向的 Period 快照，不是待提交配置差异，不代表各远端已同步；与 Zonegroup 当前本地配置可能不同。</p>
    {!valid ? <p>Period Zonegroup 列表不可用（不推断为空）</p> : <Table size="small" rowKey="id" dataSource={groups} pagination={groups.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 600 }} locale={{ emptyText: '此 Period 的 Zonegroup 列表为空' }} columns={[
      { title: 'Zonegroup ID', dataIndex: 'id', render: raw },
      { title: '名称', dataIndex: 'name', render: raw },
      { title: '主 Zone', dataIndex: 'master_zone', render: raw },
      { title: '成员 Zone', dataIndex: 'zones', render: raw },
      { title: '同步策略', dataIndex: 'sync_policy', render: (policy: unknown) => <pre>{raw(policy)}</pre> }
    ]} />}
    <details><summary>查看原生 Period（含配置与同步标记）</summary><pre>{JSON.stringify(value, null, 2)}</pre></details>
  </div>
}
