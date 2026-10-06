import type { ApiRecord } from '../../api/client'

export function hostStorageCapacity(devices: ApiRecord[]): string {
  let total = 0n
  for (const device of devices) {
    const value = device.size_bytes
    const raw = typeof value === 'string' ? value : typeof value === 'number' && Number.isSafeInteger(value) ? String(value) : ''
    if (!/^(0|[1-9][0-9]*)$/.test(raw) || BigInt(raw) > 18446744073709551615n) return '未知（部分设备容量缺失或无效）'
    total += BigInt(raw)
  }
  if (total < 1024n) return `${total} B`
  const units = ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB']
  let divisor = 1024n, index = 0
  while (index < units.length - 1 && total >= divisor * 1024n) { divisor *= 1024n; index++ }
  const hundredths = total * 100n / divisor
  return `约 ${hundredths / 100n}.${String(hundredths % 100n).padStart(2, '0')} ${units[index]}（${total} B）`
}
