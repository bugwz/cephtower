export function monSessionCount(value: unknown) {
  if (typeof value !== 'string' || !/^(0|[1-9][0-9]{0,19})$/.test(value) || BigInt(value) > 18446744073709551615n) return '未返回或格式无效'
  return value
}
