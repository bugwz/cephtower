export function rgwDefaultRealm(value: unknown): string {
  if (typeof value !== 'string') return '默认 Realm ID 未返回或不可用'
  if (value === '') return '原生返回空 ID：可能未设置，也可能读取失败'
  return value
}
