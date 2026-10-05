const valid = (value: unknown): value is number => typeof value === 'number' && Number.isSafeInteger(value) && value >= -1

export function rgwQuotaInitial(value: unknown) {
  const quota = value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : {}
  return {
    enabled: quota.enabled === true ? 'enable' : quota.enabled === false ? 'disable' : undefined,
    max_size: valid(quota.max_size) ? quota.max_size : undefined,
    max_objects: valid(quota.max_objects) ? quota.max_objects : undefined
  }
}

export function rgwQuotaInput(values: Record<string, unknown>) {
  if (values.enabled !== 'enable' && values.enabled !== 'disable') throw new Error('请明确选择启用或关闭配额')
  if (!valid(values.max_size) || !valid(values.max_objects)) throw new Error('配额必须明确填写 -1 或非负安全整数，-1 表示无限制')
  return { enabled: values.enabled === 'enable', max_size: values.max_size, max_objects: values.max_objects }
}

export function rgwQuotaConfirmation(values: Record<string, unknown>, target: string, scope: string) {
  const quota = rgwQuotaInput(values)
  const size = quota.max_size === -1 ? '无限制（-1）' : `${quota.max_size} 字节（原生向上取整后 ${(BigInt(quota.max_size) + 1023n) / 1024n * 1024n} 字节）`
  const objects = quota.max_objects === -1 ? '无限制（-1）' : `${quota.max_objects}`
  return `确认整体更新 ${target} 的${scope}？状态：${quota.enabled ? '启用' : '关闭'}；容量上限：${size}；对象上限：${objects}。0 是零配额，不是无限制。关闭时仍保存上述限制值，但不启用本项配额检查；其他作用域配额仍独立生效。`
}
