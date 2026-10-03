type RecordValue = Record<string, unknown>
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
