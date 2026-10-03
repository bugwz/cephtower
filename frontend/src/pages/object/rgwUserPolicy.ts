import { rgwIdentityList } from './rgwUserIdentity'

export function rgwUserPolicyBlocked(row?: Record<string, unknown>) {
  if (typeof row?.account_id !== 'string' || row.account_id === '' || typeof row.type !== 'string' || row.type === '' || row.type === 'root') return '仅支持已采集到账户信息的非 root 用户'
  return undefined
}

export function rgwUserPolicyOptions(row?: Record<string, unknown>) {
  const policies = rgwIdentityList(row?.managed_user_policies)
  if (!policies) throw new Error('托管策略列表不可用，请重新采集')
  return policies.map(arn => ({ label: arn, value: arn }))
}

export function rgwUserPolicyInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const reason = rgwUserPolicyBlocked(row)
  if (reason) throw new Error(reason)
  const action = values.action
  if (action !== 'attach' && action !== 'detach') throw new Error('请选择关联或解除关联')
  const arn = action === 'detach' ? values.existing_policy : values.policy_arn
  if (typeof arn !== 'string' || /\s/.test(arn) || !/^arn:[a-z0-9-]+:iam::[A-Za-z0-9-]+:policy\/[A-Za-z0-9+=,.@_/-]+$/.test(arn)) throw new Error('请输入有效的托管 IAM 策略 ARN')
  if (action === 'detach' && !rgwUserPolicyOptions(row).some(option => option.value === arn)) throw new Error('请选择已采集的关联策略')
  if (action === 'attach' && rgwIdentityList(row?.managed_user_policies)?.includes(arn)) throw new Error('该策略已关联')
  return { action, policy_arn: arn }
}
