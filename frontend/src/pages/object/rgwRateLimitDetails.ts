export function rgwRateLimitDetails(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return { state: '限流信息不可用' }
  const limits = value as Record<string, unknown>
  if (limits.enabled === false) return { state: '未启用' }
  if (limits.enabled !== true) return { state: '限流启用状态未知' }
  return {
    state: '已启用',
    limits: ['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'].map(key => {
      const raw = limits[key]
      return typeof raw === 'number' && Number.isSafeInteger(raw)
        ? raw <= 0 ? '无限制' : String(raw)
        : '未返回或超出精确显示范围'
    })
  }
}
