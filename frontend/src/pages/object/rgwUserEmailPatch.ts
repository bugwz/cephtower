export function rgwUserEmailPatch(values: Record<string, unknown>) {
  const action = values.email_action
  if (action == null || action === 'keep') return {}
  if (action === 'clear') return { email: '' }
  if (action !== 'set') throw new Error('请选择邮箱操作')
  if (typeof values.email !== 'string' || values.email.trim() === '' || /[\0\r\n]/.test(values.email)) {
    throw new Error('请输入非空单行邮箱；清空邮箱请使用清空操作')
  }
  return { email: values.email }
}
