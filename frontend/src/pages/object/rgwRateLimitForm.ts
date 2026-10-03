const keys = ['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'] as const
const valid = (value: unknown): value is number => typeof value === 'number' && Number.isSafeInteger(value) && value >= 0

export function rgwRateLimitInitial(value: unknown) {
  const limits = value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : {}
  return { enabled: limits.enabled === true, ...Object.fromEntries(keys.map(key => [key, valid(limits[key]) ? limits[key] : undefined])) }
}

export function rgwRateLimitInput(values: Record<string, unknown>) {
  const limits: Record<string, number> = {}
  for (const key of keys) {
    const value = values[key]
    if (!valid(value)) throw new Error(`${key} 必须明确填写非负安全整数，0 表示无限制`)
    limits[key] = value
  }
  return limits
}
