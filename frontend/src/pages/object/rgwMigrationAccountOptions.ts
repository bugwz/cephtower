import type { ApiRecord } from '../../api/client'
import { listAllResources, type ResourceListResult } from '../../api/resource'

export function rgwMigrationAccountOptions(inventory: ResourceListResult, tenant: string) {
  const counts = new Map<string, number>()
  for (const row of inventory.items) {
    if (typeof row.account_id === 'string') counts.set(row.account_id, (counts.get(row.account_id) ?? 0) + 1)
  }
  return inventory.items.flatMap(row => {
    const id = row.account_id
    if (typeof id !== 'string' || !/^RGW[0-9]{17}$/.test(id) || counts.get(id) !== 1 || row.tenant !== tenant) return []
    const name = typeof row.account_name === 'string' && row.account_name !== '' ? row.account_name : '账户名称未返回'
    const stale = inventory.stale !== false || row.stale !== false
    return [{ value: id, label: `${name} · ${id}${stale ? '（库存过期或新鲜度未知，提交时重新核验）' : ''}` }]
  })
}

export async function loadRgwMigrationAccountOptions(clusterId: number, row?: ApiRecord) {
  if (!Number.isSafeInteger(clusterId) || clusterId < 1 || typeof row?.tenant !== 'string') throw new Error('集群或用户租户未知，请重新采集')
  const inventory = await listAllResources('/rgw/accounts', clusterId)
  return rgwMigrationAccountOptions(inventory, row.tenant)
}
