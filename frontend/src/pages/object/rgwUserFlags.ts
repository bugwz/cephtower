export function rgwUserSuspension(value: unknown) {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 0 || value > 255) return '暂停状态未知'
  return value === 0 ? '未暂停' : '已暂停'
}

export function rgwUserBooleanFlag(value: unknown) {
  return value === true ? '是' : value === false ? '否' : '未知'
}
