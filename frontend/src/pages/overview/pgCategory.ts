// Mirrors Dashboard PgCategory states; warning takes precedence over unknown.
export function pgCategory(value: unknown) {
  const clean = new Set(['active', 'clean'])
  const working = new Set(['activating', 'backfill_wait', 'backfilling', 'creating', 'deep', 'degraded', 'forced_backfill', 'forced_recovery', 'peering', 'peered', 'recovering', 'recovery_wait', 'repair', 'scrubbing', 'snaptrim', 'snaptrim_wait'])
  const warning = new Set(['backfill_toofull', 'backfill_unfound', 'down', 'incomplete', 'inconsistent', 'recovery_toofull', 'recovery_unfound', 'remapped', 'snaptrim_error', 'stale', 'undersized'])
  const states = typeof value === 'string' ? [...new Set(value.trim().split(/[+\s]+/).filter(Boolean))] : []
  if (states.some((state) => warning.has(state))) return { key: 'warning', label: '告警', color: 'warning' }
  if (!states.length || states.some((state) => !clean.has(state) && !working.has(state))) return { key: 'unknown', label: '未知', color: 'default' }
  if (states.some((state) => working.has(state))) return { key: 'working', label: '处理中', color: 'processing' }
  return { key: 'clean', label: '正常', color: 'success' }
}

export function pgSummary(value: unknown) {
  if (!Array.isArray(value)) return null
  const counts: Record<string, number> = { clean: 0, working: 0, warning: 0, unknown: 0 }
  const seen = new Set<string>()
  let total = 0
  for (const row of value) {
    if (!row || typeof row !== 'object' || typeof row.name !== 'string' || !row.name.trim() || seen.has(row.name) || !Number.isSafeInteger(row.count) || row.count < 0 || !Number.isSafeInteger(total + row.count)) return null
    seen.add(row.name)
    total += row.count
    counts[pgCategory(row.name).key] += row.count
  }
  return { total, categories: ['active+clean', 'scrubbing', 'inconsistent', 'unknown'].map((state) => {
    const category = pgCategory(state)
    return { ...category, count: counts[category.key] }
  }) }
}

export function poolPGDistribution(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const rows = Object.entries(value).map(([name, count]) => ({ name, count }))
  const summary = pgSummary(rows)
  if (!summary) return null
  return { ...summary, rows: rows.map((row) => ({ ...row, count: row.count as number, category: pgCategory(row.name) })) }
}
