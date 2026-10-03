// IAM resource limits differ from max_buckets: zero prevents creation.
export function rgwAccountLimit(value: unknown) {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < -2147483648 || value > 2147483647) {
    return '资源上限未返回或无效'
  }
  if (value < 0) return '无限制'
  if (value === 0) return '0（禁止新增）'
  return String(value)
}
