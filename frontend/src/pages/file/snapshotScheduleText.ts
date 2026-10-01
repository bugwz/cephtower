const intervals: Record<string, string> = { m: '分钟', h: '小时', d: '天', w: '周', M: '月', y: '年' }
const retentionUnits: Record<string, string> = { m: '每分钟', h: '每小时', d: '每日', w: '每周', M: '每月', y: '每年', n: '最近' }

export function scheduleIntervalText(value: unknown): string {
  if (typeof value !== 'string' || !value) return '未知'
  const match = /^([1-9][0-9]*)([mhdwMy])$/.exec(value)
  return match ? `每 ${match[1]} ${intervals[match[2]]}（${value}）` : value
}

export function scheduleRetentionText(value: unknown): string {
  if (value == null) return '未知'
  let entries: Array<[string, unknown]>
  if (typeof value === 'string') {
    if (!value) return '未设置'
    if (!/^([1-9][0-9]*[nmhdwMy])+$/.test(value)) return value
    entries = Array.from(value.matchAll(/([1-9][0-9]*)([nmhdwMy])/g), (match) => [match[2], match[1]])
  } else if (typeof value === 'object' && !Array.isArray(value)) {
    entries = Object.entries(value)
  } else return '未知'
  if (!entries.length) return '未设置'
  return entries.map(([unit, count]) => {
    const text = typeof count === 'string' ? count : typeof count === 'number' && Number.isSafeInteger(count) ? String(count) : ''
    if (!/^[1-9][0-9]*$/.test(text) || !Object.prototype.hasOwnProperty.call(retentionUnits, unit)) return `${unit}：未知规则`
    return `${retentionUnits[unit]}保留 ${text} 个快照`
  }).join('；')
}
