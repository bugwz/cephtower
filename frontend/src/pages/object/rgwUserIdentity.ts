export function rgwIdentityText(value: unknown, empty: string) {
  return typeof value === 'string' ? value === '' ? empty : value : '未返回或格式无效'
}

export function rgwIdentityList(value: unknown) {
  return Array.isArray(value) && value.every(item => typeof item === 'string' && item.length > 0) ? value as string[] : undefined
}
