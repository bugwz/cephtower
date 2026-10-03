import type { ApiRecord } from '../../api/client'
import type { FieldColumn } from '../../components/DataTable'

export function hardwareAttributeColumns(category: string): FieldColumn[] {
  const fields = new Map<string, string[][]>([
    ['memory', [['description', '内存描述']]],
    ['storage', [['description', '存储描述'], ['model', '型号'], ['capacity_bytes', '容量（bytes）'], ['protocol', '协议'], ['serial_number', '序列号']]],
    ['processors', [['model', '处理器型号'], ['total_cores', '核心数'], ['total_threads', '线程数']]],
    ['network', [['name', '接口名称'], ['speed_mbps', '速率（Mbps）']]],
    ['power', [['name', '电源名称'], ['model', '型号'], ['manufacturer', '制造商']]],
    ['fans', [['name', '风扇名称']]],
    ['firmwares', [['name', '固件名称'], ['version', '版本（原值）'], ['release_date', '发布日期（原值）']]],
  ])
  return (fields.get(category) ?? []).map(([key, title]) => ({ key, title, render: (value) => typeof value === 'string' && value.trim() !== '' ? value : '未知（未返回）' }))
}

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
