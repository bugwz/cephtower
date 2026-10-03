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

export function rgwUserUpdateConfirmation(values: Record<string, unknown>, uid: string) {
  const flags = rgwUserFlagPatch(values)
  let text = `更新 RGW 用户 ${JSON.stringify(uid)} 的已填写属性？`
  if (flags.system === true) text += '启用系统用户标志将授予 RGW 内部系统操作能力，请确认授权范围。'
  if (flags.system === false) text += '关闭系统用户标志可能中断依赖此身份的系统操作或同步。'
  if (flags.suspended === true) text += '暂停用户会阻止其客户端访问。'
  if (flags.suspended === false) text += '解除暂停会恢复该用户访问。'
  if (flags.suspended !== undefined) text += '与其它属性同时修改时按多条命令执行，失败可能已有部分变更，请检查实际状态，不要盲目重试。'
  return text
}
