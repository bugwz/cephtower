export function rgwBucketLimit(value: unknown) {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < -2147483648 || value > 2147483647) {
    return 'Bucket 上限未返回或无效'
  }
  if (value < 0) return '禁止创建 Bucket'
  if (value === 0) return '无限制'
  return String(value)
}
