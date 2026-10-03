export function rgwUserDisplayNamePatch(value: unknown, current: unknown) {
  if (value == null || value === '' || value === current) return {}
  if (typeof value !== 'string' || value.trim() === '' || /[\0\r\n]/.test(value)) {
    throw new Error('显示名必须是非空单行文本；留空表示不修改')
  }
  return { display_name: value }
}
