type RecordValue = Record<string, unknown>
export function rgwStorageScope(scope: unknown, accountId: unknown) {
  if (scope === 'user') return '用户汇总'
  if (scope === 'account') {
    const identity = typeof accountId === 'string' && accountId !== '' ? ` ${accountId}` : '（账户 ID 未返回）'
    return `所属账户${identity} 的汇总，非单用户用量；同账户用户之间不可相加`
  }
  return '统计范围未知，请重新采集'
}
export function rgwStorageTimes(value: unknown, account = false) {
  const stats = record(value) ? value : {}
  const display = (value: unknown) => typeof value === 'string' && value !== '' ? value : '未返回或格式无效'
  return {
    synced: display(stats[account ? 'last_synced' : 'last_stats_sync']),
    updated: display(stats[account ? 'last_updated' : 'last_stats_update'])
  }
}
function record(value: unknown): value is RecordValue {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

export function rgwStorageRows(value: unknown, categorized = false) {
  if (!record(value)) return undefined
  const entries = categorized ? Object.entries(value) : [['汇总', value.stats] as const]
  if (entries.some(([, stats]) => !record(stats))) return undefined
  const count = (value: unknown) => typeof value === 'number' && Number.isSafeInteger(value) && value >= 0
    ? String(value) : '未返回或超出精确显示范围'
  return entries.map(([category, raw]) => {
    const stats = raw as RecordValue
    return {
      category,
      size: count(stats.size),
      actual: count(stats.size_actual),
      utilized: count(stats.size_utilized),
      objects: count(stats.num_objects)
    }
  })
}
