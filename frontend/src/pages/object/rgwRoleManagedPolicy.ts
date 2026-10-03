function rolePolicyIdentity(row?: Record<string, unknown>) {
  if (!row || row.stale === true || typeof row.AccountId !== 'string' || !/^RGW[0-9]{17}$/.test(row.AccountId) || typeof row.RoleName !== 'string' || !/^[A-Za-z0-9_+=,.@-]{1,64}$/.test(row.RoleName) || typeof row.RoleId !== 'string' || !row.RoleId || typeof row.Arn !== 'string' || !row.Arn.startsWith(`arn:aws:iam::${row.AccountId}:role/`) || !row.Arn.endsWith(`/${row.RoleName}`)) throw new Error('仅支持身份完整且未过期的 Account 角色')
  // Native role dump omits this field when its set is empty. IAM rechecks the
  // complete list before writing; omission alone is never proof of permission.
  const policies = row.ManagedPermissionPolicies === undefined ? [] : row.ManagedPermissionPolicies
  if (!Array.isArray(policies) || policies.some(item => typeof item !== 'string' || !item) || new Set(policies).size !== policies.length) throw new Error('托管策略快照不完整或有重复项，请刷新')
  return {account_id:row.AccountId,name:row.RoleName,expected_role_id:row.RoleId,expected_role_arn:row.Arn,expected_policies:[...policies] as string[]}
}
export function roleManagedPolicyBlocked(row: Record<string, unknown>) {
  try {rolePolicyIdentity(row);return undefined} catch (error) {return (error as Error).message}
}
export function roleManagedPolicyInitial(row?: Record<string, unknown>) {
  const identity = rolePolicyIdentity(row)
  return {account_id:identity.account_id,name:identity.name}
}
export function roleManagedPolicyInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const identity = rolePolicyIdentity(row)
  if (values.account_id !== identity.account_id || values.name !== identity.name) throw new Error('不可更改角色身份')
  if (values.mode !== 'attach' && values.mode !== 'detach') throw new Error('请选择关联或解除')
  if (typeof values.owner_uid !== 'string' || !/^[A-Za-z0-9_.$@-]{1,512}$/.test(values.owner_uid)) throw new Error('请填写已配置 S3 永久密钥所属完整 UID')
  const policy = values.policy_arn
  if (typeof policy !== 'string' || !policy.startsWith('arn:aws:iam::aws:policy/') || policy.length > 2048 || /\s/.test(policy) || policy === 'arn:aws:iam::aws:policy/') throw new Error('请填写 Ceph 支持的完整托管策略 ARN')
  const present = identity.expected_policies.includes(policy)
  if ((values.mode === 'attach' && present) || (values.mode === 'detach' && !present)) throw new Error('所选操作不会改变当前托管策略集合')
  if (values.confirm_managed_policy !== 'acknowledged') throw new Error('请确认角色权限变更影响')
  return {...identity,owner_uid:values.owner_uid,mode:values.mode,policy_arn:policy}
}
export function roleManagedPolicyConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = roleManagedPolicyInput(values,row)
  return `确认${input.mode === 'attach' ? '关联' : '解除'} ${input.policy_arn} 到 Account ${input.account_id} 的角色 ${input.name}（ID ${input.expected_role_id}）？这将改变角色的托管权限，不修改信任策略或内联策略，不保证现有会话立即失效。后端核对永久 S3 密钥所属 UID 与 Account，要求 HTTPS RGW 端点；不支持临时会话或子用户密钥。原生命令省略空策略字段时提交空快照，仍须 IAM 验证。策略是否受目标版本支持及 IAM 权限由 Ceph 判定。前后读取不是原子锁，外部并发删除/重建角色仍有竞态；失败可能已经生效，不自动重试或回滚。`
}
