export function rgwBucketIndexText(value: unknown) {
  return typeof value === 'string' ? (value === '' ? '（空字符串）' : value) : '未返回或格式无效'
}

export function rgwBucketIndexCount(value: unknown) {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0 ? String(value) : '未返回或超出安全整数范围'
}
