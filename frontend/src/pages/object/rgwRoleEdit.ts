export function rgwRoleInitial(row?: Record<string, unknown>) {
  const policy = row?.AssumeRolePolicyDocument
  const duration = row?.MaxSessionDuration
  return {
    assume_role_policy: typeof policy === 'string' ? policy : policy && typeof policy === 'object' && !Array.isArray(policy) ? JSON.stringify(policy, null, 2) : undefined,
    max_session_duration: typeof duration === 'number' && Number.isInteger(duration) && duration >= 3600 && duration <= 43200 ? duration : undefined
  }
}

export function rgwRolePatch(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = rgwRoleInitial(row)
  const patch: Record<string, string | number> = {}
  const policy = values.assume_role_policy
  if (policy != null && policy !== '' && policy !== initial.assume_role_policy) {
    if (typeof policy !== 'string') throw new Error('信任策略必须是 JSON 对象')
    const parsed: unknown = JSON.parse(policy)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('信任策略必须是 JSON 对象')
    patch.assume_role_policy = policy
  }
  const duration = values.max_session_duration
  if (duration != null && duration !== '' && duration !== initial.max_session_duration) {
    if (typeof duration !== 'number' || !Number.isInteger(duration) || duration < 3600 || duration > 43200) throw new Error('会话时长必须是 3600 到 43200 秒之间的整数')
    patch.max_session_duration = duration
  }
  if (Object.keys(patch).length === 0) throw new Error('请修改至少一个角色字段')
  return patch
}
