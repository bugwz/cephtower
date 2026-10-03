export function rbdSnapshotLimitText(value: unknown): string {
  if (typeof value === 'string' && value.length <= 20 && /^(0|[1-9][0-9]*)$/.test(value) && BigInt(value) < 18446744073709551615n) return value
  return '未返回有效上限（原生命令在无限制时省略此字段）'
}
