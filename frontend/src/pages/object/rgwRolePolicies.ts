export function rgwRolePolicies(value: unknown) {
  if (!Array.isArray(value)) return undefined
  if (value.some(policy => !policy || typeof policy !== 'object' || Array.isArray(policy)
    || typeof policy.PolicyName !== 'string' || policy.PolicyName === '' || typeof policy.PolicyValue !== 'string')) return undefined
  return value.map((policy, index) => ({ id: index, name: policy.PolicyName as string, document: policy.PolicyValue as string }))
}

export function rgwPolicyDeleteOptions(row?: Record<string, unknown>) {
  const policies = rgwRolePolicies(row?.PermissionPolicies)
  if (!policies) throw new Error('内联策略信息不可用，请重新采集')
  const names = policies.map(policy => policy.name)
  if (new Set(names).size !== names.length) throw new Error('内联策略名称重复，请重新采集')
  return names.map(name => ({ label: name, value: name }))
}

export function rgwPolicyMutation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  if (values.action === 'delete') {
    const name = values.existing_policy
    if (typeof name !== 'string' || !rgwPolicyDeleteOptions(row).some(option => option.value === name)) throw new Error('请选择当前角色已采集的内联策略')
    return { action: 'delete', policy_name: name }
  }
  if (values.action !== 'put') throw new Error('请选择有效的策略操作')
  if (typeof values.policy_name !== 'string' || values.policy_name.trim() === '') throw new Error('请输入策略名称')
  if (typeof values.policy_document !== 'string') throw new Error('请输入策略 JSON 对象')
  const doc: unknown = JSON.parse(values.policy_document)
  if (!doc || typeof doc !== 'object' || Array.isArray(doc)) throw new Error('请输入策略 JSON 对象')
  return { action: 'put', policy_name: values.policy_name, policy_document: values.policy_document }
}
