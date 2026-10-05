const keys = ['max_read_ops', 'max_write_ops', 'max_read_bytes', 'max_write_bytes'] as const
const valid = (value: unknown): value is number => typeof value === 'number' && Number.isSafeInteger(value) && value >= 0

export function rgwRateLimitInitial(value: unknown) {
  const limits = value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : {}
  return { enabled: limits.enabled === true ? 'enable' : limits.enabled === false ? 'disable' : undefined, ...Object.fromEntries(keys.map(key => [key, valid(limits[key]) ? limits[key] : undefined])) }
}

export function rgwRateLimitInput(values: Record<string, unknown>) {
  if (values.enabled !== 'enable' && values.enabled !== 'disable') throw new Error('请明确选择启用或关闭限流')
  const limits: Record<string, number> = {}
  for (const key of keys) {
    const value = values[key]
    if (!valid(value)) throw new Error(`${key} 必须明确填写非负安全整数，0 表示无限制`)
    limits[key] = value
  }
  return { ...limits, enabled: values.enabled === 'enable' }
}

export function rgwRateLimitConfirmation(values: Record<string, unknown>, target: string) {
  const limits = rgwRateLimitInput(values)
  const labels = ['读请求数', '写请求数', '读取字节数', '写入字节数']
  const summary = keys.map((key, index) => `${labels[index]}：${values[key] === 0 ? '无限制（0）' : values[key]}`).join('；')
  return `确认整体更新 ${target} 的限流设置？状态：${limits.enabled ? '启用' : '关闭'}；${summary}。限制按每 RGW 每分钟计算，不是集群总限额。关闭时仍保存上述限制值，其他作用域限流仍独立生效。`
}
