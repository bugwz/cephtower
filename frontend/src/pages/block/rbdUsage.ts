export function rbdUsageText(value: unknown, features: unknown): string {
  if (!Array.isArray(features) || !features.every((feature) => typeof feature === 'string')) return '特性信息不可用，无法确认用量采集条件'
  if (!features.includes('fast-diff')) return '不可用：未启用 fast-diff'
  if (typeof value === 'number' && Number.isSafeInteger(value) && value >= 0) return String(value)
  return '用量未返回或无效，请重新采集'
}
