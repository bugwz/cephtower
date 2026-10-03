export function rgwUserCreateFlags(values: Record<string, unknown>) {
  const flags: Record<string, boolean> = {}
  for (const field of ['system', 'suspended']) {
    if (values[field] !== 'enable' && values[field] !== 'disable') throw new Error('请选择新用户的系统标志和暂停状态')
    flags[field] = values[field] === 'enable'
  }
  return flags
}
