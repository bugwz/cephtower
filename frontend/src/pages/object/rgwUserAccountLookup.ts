import { listAllResources } from '../../api/resource'

export async function loadRgwUserAccount(clusterId: number, accountId: string) {
  if (!Number.isSafeInteger(clusterId) || clusterId <= 0 || !accountId || accountId !== accountId.trim()) throw new Error('集群或账户 ID 无效')
  const inventory = await listAllResources('/rgw/accounts', clusterId)
  const matches = inventory.items.filter(row => row.account_id === accountId)
  if (matches.length > 1) throw new Error('账户库存存在重复 ID，无法确认关联账户')
  const account = matches[0]
  return { account, stale: inventory.stale !== false || (account !== undefined && account.stale !== false), observedAt: account?.observed_at ?? inventory.observedAt }
}
