import type { ApiRecord } from '../../api/client'

export type HardwareHealthGroup = 'ok' | 'other' | 'unknown'

export function hardwareHealthGroup(value: unknown): HardwareHealthGroup {
  if (value === 'OK') return 'ok'
  if (typeof value !== 'string' || value.trim() === '') return 'unknown'
  return 'other'
}

export function hardwareHealthCounts(rows: ApiRecord[]) {
  const counts = { total: rows.length, ok: 0, other: 0, unknown: 0 }
  for (const row of rows) counts[hardwareHealthGroup(row.health)]++
  return counts
}
