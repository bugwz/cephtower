export function rgwUserFlagPatch(values: Record<string, unknown>) {
  const patch: Record<string, boolean> = {}
  for (const key of ['suspended', 'system']) {
    const value = values[key]
    if (value === undefined || value === null || value === 'keep') continue
    if (value !== 'enable' && value !== 'disable') throw new Error('请选择保持不变、启用或关闭')
    patch[key] = value === 'enable'
  }
  return patch
}
