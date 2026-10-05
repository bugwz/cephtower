import { jsonInit, request, type ApiRecord } from '../../api/client'

export interface RgwTopologyNode { key: string; title: string; details: ApiRecord; children: RgwTopologyNode[] }
export interface RgwTopology { nodes: RgwTopologyNode[]; stale: boolean }
const record = (value: unknown): value is ApiRecord => !!value && typeof value === 'object' && !Array.isArray(value)

// Read every page without the generic inventory helper's unavailable-to-empty
// conversion. A failed topology source must not become an empty branch.
export async function readRgwTopology(clusterId: number): Promise<RgwTopology> {
  async function read(path: string) {
    const rows: ApiRecord[] = []
    const cursors = new Set<string>()
    let cursor = '', stale = false
    for (let page = 0; page < 100; page++) {
      const query = new URLSearchParams({ limit: '500' })
      if (cursor) query.set('cursor', cursor)
      const result = await request<ApiRecord>(`${path}?${query}`, jsonInit('GET', { cluster_id: clusterId }, { suppressErrorNotification: true, cache: 'no-store' }))
      if (!record(result) || !Array.isArray(result.items) || !record(result.meta) || typeof result.meta.stale !== 'boolean' || !record(result.pagination)) throw new Error('Invalid topology inventory')
      for (const item of result.items) {
        if (!record(item) || !record(item.data) || typeof item.stale !== 'boolean') throw new Error('Invalid topology resource')
        rows.push({ ...item.data, stale: item.stale, observed_at: item.observed_at })
        stale ||= item.stale
      }
      stale ||= result.meta.stale
      const next = result.pagination.next_cursor
      if (next === null || next === undefined || next === '') return { rows, stale }
      if (typeof next !== 'string' || cursors.has(next)) throw new Error('Invalid topology pagination')
      cursors.add(next)
      cursor = next
    }
    throw new Error('Topology inventory exceeds page limit')
  }
  const [realms, groups, zones] = await Promise.all(['/rgw/realms', '/rgw/zonegroups', '/rgw/zones'].map(read))
  return { nodes: buildRgwTopology(realms.rows, groups.rows, zones.rows), stale: realms.stale || groups.stale || zones.stale }
}

export function buildRgwTopology(realms: ApiRecord[], groups: ApiRecord[], zones: ApiRecord[]): RgwTopologyNode[] {
  function index(rows: ApiRecord[]) {
    const result = new Map<string, ApiRecord>()
    for (const row of rows) {
      if (!record(row) || typeof row.id !== 'string' || !row.id || typeof row.name !== 'string' || !row.name || result.has(row.id)) throw new Error('Invalid topology identity')
      result.set(row.id, row)
    }
    return result
  }
  const realmMap = index(realms), groupMap = index(groups), zoneMap = index(zones)
  const nodes: RgwTopologyNode[] = [], attached = new Set<string>()
  const realmNodes = new Map<string, RgwTopologyNode>()
  function node(kind: string, row: ApiRecord, key: string, extra: ApiRecord = {}): RgwTopologyNode {
    const details: ApiRecord = { id: row.id, name: row.name, ...extra }
    // Deliberate field allowlist: do not expose zone system_key credentials.
    for (const field of ['realm_id', 'is_default', 'is_master', 'master_zone', 'endpoints', 'read_only', 'tier_type', 'sync_from_all', 'sync_from', 'supported_features', 'stale', 'observed_at']) {
      if (row[field] !== undefined) details[field] = row[field]
    }
    return { key, title: `${kind}: ${String(row.name)}${row.is_default === true ? ' [默认]' : ''}${row.is_master === true ? ' [主组]' : ''}${row.tier_type === 'archive' ? ' [归档]' : ''}${row.stale === true ? ' [过期]' : ''}`, details, children: [] }
  }
  for (const [id, row] of realmMap) {
    const item = node('Realm', row, `realm:${id}`)
    realmNodes.set(id, item)
    nodes.push(item)
  }
  for (const [id, group] of groupMap) {
    const item = node('Zonegroup', group, `group:${id}`)
    const realm = typeof group.realm_id === 'string' ? realmNodes.get(group.realm_id) : undefined
    if (realm) realm.children.push(item)
    else {
      item.details.relationship = group.realm_id === '' ? '未关联 Realm' : 'Realm 归属缺失或未找到对应库存'
      nodes.push(item)
    }
    if (!Array.isArray(group.zones)) throw new Error('Missing zonegroup membership')
    const members = index(group.zones)
    if (typeof group.master_zone !== 'string' || !members.has(group.master_zone)) item.details.warning = '未找到有效主 Zone 成员'
    for (const [zoneID, member] of members) {
      attached.add(zoneID)
      const local = zoneMap.get(zoneID)
      const child = node('Zone', member, `group:${id}:zone:${zoneID}`, {
        role: typeof group.master_zone === 'string' && group.master_zone ? zoneID === group.master_zone ? '主 Zone' : '非主 Zone' : '角色不可用',
        local_detail: local ? '存在本地 Zone 详情' : '未找到本地 Zone 详情（不能据此认定远端）',
        ...(local ? { local_name: local.name, local_realm_id: local.realm_id, local_is_default: local.is_default, local_observed_at: local.observed_at, local_stale: local.stale } : {})
      })
      if (local && (local.name !== member.name || (typeof group.realm_id === 'string' && local.realm_id !== group.realm_id))) child.details.warning = '成员配置与本地 Zone 身份或 Realm 归属不一致'
      item.children.push(child)
    }
  }
  for (const [id, zone] of zoneMap) if (!attached.has(id)) nodes.push(node('Zone', zone, `zone:${id}`, { relationship: '本次库存未列出所属 Zonegroup' }))
  return nodes
}
