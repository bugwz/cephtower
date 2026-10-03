export function rgwBucketLimitPatch(value: unknown, current: unknown) {
  if (value === current) return {}
  return rgwBucketLimitInput(value)
}

export function rgwBucketLimitInput(value: unknown) {
  if (value === undefined || value === null || value === '') return {}
  if (typeof value !== 'number' || !Number.isInteger(value) || value < -1 || value > 2147483647) {
    throw new Error('最大 Bucket 数必须是 -1 到 2147483647 之间的整数')
  }
  return { max_buckets: value }
}

export function rgwBucketLimit(value: unknown) {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < -2147483648 || value > 2147483647) {
    return 'Bucket 上限未返回或无效'
  }
  if (value < 0) return '禁止创建 Bucket'
  if (value === 0) return '无限制'
  return String(value)
}
