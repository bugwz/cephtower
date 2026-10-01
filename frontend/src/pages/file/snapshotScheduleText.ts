const intervals: Record<string, string> = { m: '分钟', h: '小时', d: '天', w: '周', M: '月', y: '年' }
const retentionUnits: Record<string, string> = { m: '每分钟', h: '每小时', d: '每日', w: '每周', M: '每月', y: '每年', n: '最近' }

export function scheduleActiveText(value: unknown): string {
  return value === true ? '启用' : value === false ? '停用' : '未知'
}

export function scheduleToggleAction(value: unknown): 'activate' | 'deactivate' | undefined {
  return value === true ? 'deactivate' : value === false ? 'activate' : undefined
}

export function scheduleIntervalText(value: unknown): string {
  if (typeof value !== 'string' || !value) return '未知'
  const match = /^([1-9][0-9]*)([mhdwMy])$/.exec(value)
  return match ? `每 ${match[1]} ${intervals[match[2]]}（${value}）` : value
}

export const scheduleFrequencyOptions = [
  { label: '小时', value: 'h' }, { label: '天', value: 'd' },
  { label: '周', value: 'w' }, { label: '月（30 天）', value: 'M' },
  { label: '年（365 天）', value: 'y' }
]

export function buildScheduleInterval(interval: unknown, frequency: unknown): string {
  if (typeof interval !== 'string' || !/^[1-9][0-9]*$/.test(interval)) throw new Error('周期间隔必须为正整数')
  if (!scheduleFrequencyOptions.some((option) => option.value === frequency)) throw new Error('请选择周期单位')
  return `${interval}${frequency}`
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
