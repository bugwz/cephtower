export function poolUsagePercent(stored?: string | null, size?: string | null): number | undefined {
  if (stored == null || size == null || !/^\d+$/.test(stored) || !/^\d+$/.test(size)) return undefined
  const total = BigInt(size), used = BigInt(stored)
  if (total === 0n || used > total) return undefined
  return Number(used * 10000n / total) / 100
}

export function formatPoolBytes(raw?: string | null): string {
  if (raw == null || !/^\d+$/.test(raw)) return '—'
  const bytes = BigInt(raw)
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB']
  let unit = 0, divisor = 1n
  while (bytes >= divisor * 1024n && unit < units.length - 1) { divisor *= 1024n; unit++ }
  if (unit === 0) return `${raw} B`
  const hundredths = bytes * 100n / divisor
  return `${hundredths / 100n}.${String(hundredths % 100n).padStart(2, '0')} ${units[unit]}`
}
