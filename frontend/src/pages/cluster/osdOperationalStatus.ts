import { isRecord, type ApiRecord } from '../../api/client'

export function osdOperationalStatus(record: ApiRecord, osdId: string, removals: unknown, queueStale: unknown, blocked: boolean): string {
  if (blocked || record.stale !== false || queueStale !== false || !Array.isArray(removals) || !/^(0|[1-9][0-9]*)$/.test(osdId) || BigInt(osdId) > 2147483647n) return '未知（需刷新库存）'
  const seen = new Set<number>()
  let selected: ApiRecord | undefined
  for (const removal of removals) {
    if (!isRecord(removal) || removal.stale !== false || typeof removal.osd_id !== 'number' || !Number.isInteger(removal.osd_id) || removal.osd_id < 0 || removal.osd_id > 2147483647 || seen.has(removal.osd_id)) return '未知（移除队列无效）'
    seen.add(removal.osd_id)
    if (String(removal.osd_id) === osdId) selected = removal
  }
  if (!selected) return '未在移除队列中'
  return selected.replace === true ? '移除队列中（保留 ID 替换）' : '移除队列中'
}
