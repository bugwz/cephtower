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
