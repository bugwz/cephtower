export function rgwAccountLimitPatch(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const patch: Record<string, number> = {}
  for (const key of ['max_users', 'max_roles', 'max_groups', 'max_buckets', 'max_access_keys']) {
    const value = values[key]
    if (value == null || value === '' || value === row?.[key]) continue
    if (typeof value !== 'number' || !Number.isInteger(value) || value < -1 || value > 2147483647) {
      throw new Error(`${key} 必须是 -1 到 2147483647 之间的整数`)
    }
    patch[key] = value
  }
  return patch
}

// IAM resource limits differ from max_buckets: zero prevents creation.
export function rgwAccountLimit(value: unknown) {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < -2147483648 || value > 2147483647) {
    return '资源上限未返回或无效'
  }
  if (value < 0) return '无限制'
  if (value === 0) return '0（禁止新增）'
  return String(value)
}
