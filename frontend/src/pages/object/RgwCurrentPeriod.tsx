import { Table } from 'antd'

function record(value: unknown): value is Record<string, unknown> { return !!value && typeof value === 'object' && !Array.isArray(value) }
const raw = (v: unknown) => v === undefined ? '未返回' : JSON.stringify(v)
const epoch = (v: unknown) => typeof v === 'number' && Number.isSafeInteger(v) && v >= 0 ? String(v) : '不可用（非精确非负整数）'
const flag = (value: unknown) => value === true ? '是' : value === false ? '否' : '未返回或类型异常'
const names = (value: unknown) => Array.isArray(value) && value.every(item => typeof item === 'string') ? value.length ? value.join('、') : '空列表' : '未返回或类型异常'

export function periodZoneRole(id: unknown, master: unknown) {
  return typeof master === 'string' && master && typeof id === 'string' && id ? id === master ? '主 Zone' : '非主 Zone' : '主 Zone 信息不可用'
}

export function RgwPeriodZones({ value, master }: { value: unknown; master: unknown }) {
  if (!Array.isArray(value) || !value.every(zone => record(zone) && typeof zone.id === 'string' && !!zone.id && typeof zone.name === 'string' && !!zone.name) || new Set(value.map(zone => zone.id)).size !== value.length) return <p>成员 Zone 列表不可用（不推断为空）</p>
  return <div>
    <p>主 Zone 按本 Zonegroup 的 master_zone ID 判断；非主不代表远端，也不代表同步正常。sync_from_all 为是时，原生允许从所有来源同步，sync_from 不是排除列表。</p>
    {typeof master !== 'string' || !master || !value.some(zone => zone.id === master) ? <p>未找到有效主 Zone 成员，请核对原生配置。</p> : null}
    <Table size="small" rowKey="id" dataSource={value} pagination={value.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 1800 }} locale={{ emptyText: '此 Zonegroup 的成员 Zone 列表为空' }} columns={[
      { title: 'Zone ID', dataIndex: 'id', render: raw },
      { title: '名称', dataIndex: 'name', render: raw },
      { title: '组内角色', dataIndex: 'id', render: (id: unknown) => periodZoneRole(id, master) },
      { title: '端点', dataIndex: 'endpoints', render: names },
      { title: '只读', dataIndex: 'read_only', render: flag },
      { title: '类型', dataIndex: 'tier_type', render: (value: unknown) => value === '' ? '普通 Zone' : value === 'archive' ? '归档 Zone' : raw(value) },
      { title: '从全部来源同步', dataIndex: 'sync_from_all', render: flag },
      { title: '原生 sync_from', dataIndex: 'sync_from', render: names },
      { title: '重定向 Zone', dataIndex: 'redirect_zone', render: (value: unknown) => value === '' ? '未设置' : raw(value) },
      { title: '元数据日志', dataIndex: 'log_meta', render: flag },
      { title: '数据日志', dataIndex: 'log_data', render: flag },
      { title: '桶索引最大分片', dataIndex: 'bucket_index_max_shards', render: epoch },
      { title: '支持特性', dataIndex: 'supported_features', render: names }
    ]} />
  </div>
}
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
    {!valid ? <p>Period Zonegroup 列表不可用（不推断为空）</p> : <Table size="small" rowKey="id" dataSource={groups} pagination={groups.length > 5 ? { pageSize: 5 } : false} scroll={{ x: 1800 }} locale={{ emptyText: '此 Period 的 Zonegroup 列表为空' }} columns={[
      { title: 'Zonegroup ID', dataIndex: 'id', render: raw },
      { title: '名称', dataIndex: 'name', render: raw },
      { title: 'API 名称', dataIndex: 'api_name', render: raw },
      { title: '主组配置标志', dataIndex: 'is_master', render: flag },
      { title: '端点', dataIndex: 'endpoints', render: names },
      { title: 'S3 访问域名', dataIndex: 'hostnames', render: names },
      { title: '静态网站域名', dataIndex: 'hostnames_s3website', render: names },
      { title: '默认放置规则（原值）', dataIndex: 'default_placement', render: raw },
      { title: '启用特性', dataIndex: 'enabled_features', render: names },
      { title: '主 Zone', dataIndex: 'master_zone', render: raw },
      { title: '成员 Zone', dataIndex: 'zones', render: (zones: unknown, group: Record<string, unknown>) => <details><summary>查看成员配置</summary><RgwPeriodZones value={zones} master={group.master_zone} /></details> },
      { title: '同步策略', dataIndex: 'sync_policy', render: (policy: unknown) => <pre>{raw(policy)}</pre> }
    ]} />}
    <details><summary>查看原生 Period（含配置与同步标记）</summary><pre>{JSON.stringify(value, null, 2)}</pre></details>
  </div>
}
