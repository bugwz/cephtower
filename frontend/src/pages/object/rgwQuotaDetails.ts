export function rgwQuotaDetails(value: unknown) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return { state: '配额信息不可用' }
  const quota = value as Record<string, unknown>
  if (quota.enabled === false) return { state: '未启用' }
  if (quota.enabled !== true) return { state: '配额启用状态未知' }
  const limit = (raw: unknown) => typeof raw === 'number' && Number.isSafeInteger(raw)
    ? raw < 0 ? '无限制' : String(raw) : '未返回或超出精确显示范围'
  return { state: '已启用', size: limit(quota.max_size), objects: limit(quota.max_objects) }
}
