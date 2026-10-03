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
