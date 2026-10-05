const unavailable = '未返回或格式无效'

export function monQuorumMembers(value: unknown): string {
  if (!Array.isArray(value) || value.some(item => typeof item !== 'string' || !item || item.trim() !== item)) return unavailable
  return value.length ? value.join('、') : '无仲裁成员'
}

export function monQuorumLeader(value: unknown): string {
  if (typeof value !== 'string' || value.trim() !== value) return unavailable
  return value || '无 Leader'
}

export function monQuorumInteger(value: unknown, seconds = false): string {
  if (typeof value !== 'string' || !/^(0|[1-9][0-9]*)$/.test(value) || value.length > 20 || BigInt(value) > 18446744073709551615n) return unavailable
  return value + (seconds ? ' 秒' : '')
}
