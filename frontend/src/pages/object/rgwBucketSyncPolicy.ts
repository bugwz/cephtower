export function rgwBucketSyncPolicy(value: unknown): string {
  const unavailable = '桶本地同步策略不可用（不推断已停用）'
  if (!value || typeof value !== 'object' || Array.isArray(value)) return unavailable
  const groups = (value as Record<string, unknown>).groups
  if (!Array.isArray(groups)) return unavailable
  const seen = new Set<string>()
  const lines: string[] = []
  for (const group of groups) {
    if (!group || typeof group !== 'object' || Array.isArray(group) || typeof group.id !== 'string' || typeof group.status !== 'string' || !group.status
      || !Array.isArray(group.pipes) || !group.data_flow || typeof group.data_flow !== 'object' || Array.isArray(group.data_flow) || seen.has(group.id)) return unavailable
    seen.add(group.id)
    const state = group.status === 'enabled' ? '已启用'
      : group.status === 'allowed' ? '允许（未启用）'
      : group.status === 'forbidden' ? '禁止'
      : `未知状态 ${JSON.stringify(group.status)}`
    lines.push(`${JSON.stringify(group.id)}：${state}`)
  }
  return `${lines.length ? lines.join('；') : '无桶本地同步组'}。仅为采集时的配置，不代表 Zonegroup 继承策略、有效复制链路或同步进度。`
}
