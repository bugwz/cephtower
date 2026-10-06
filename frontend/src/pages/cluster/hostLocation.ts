import { isRecord } from '../../api/client'

export function hostInitialLocation(value: unknown): string {
  if (!isRecord(value)) return '未报告'
  const entries = Object.entries(value)
  if (!entries.length) return '未设置'
  if (entries.some(([key, entry]) => !key.trim() || key.trim() !== key || typeof entry !== 'string' || !entry.trim() || entry.trim() !== entry)) return '未知（位置数据无效）'
  return entries.sort(([a], [b]) => a.localeCompare(b)).map(([key, entry]) => `${key}=${entry}`).join('；')
}
