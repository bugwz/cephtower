export function rgwAccountTextPatch(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const patch: Record<string, string> = {}
  for (const key of ['account_name', 'email']) {
    const value = values[key]
    if (value === undefined || value === row?.[key]) continue
    // Empty initial values for unavailable or empty metadata are not edits.
    if (value === '' && (row?.[key] == null || row?.[key] === '')) continue
    if (typeof value !== 'string' || value.trim() === '' || /[\0\r\n]/.test(value)) {
      throw new Error(`${key} 必须是非空单行文本；原生命令不支持清空`)
    }
    patch[key] = value
  }
  return patch
}
