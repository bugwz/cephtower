export function rgwBucketPlacement(value: unknown) {
  const placement = value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : {}
  const display = (value: unknown) => typeof value === 'string' ? (value === '' ? '未显式指定' : value) : '未返回或格式无效'
  return {
    data: display(placement.data_pool),
    extra: display(placement.data_extra_pool),
    index: display(placement.index_pool)
  }
}
