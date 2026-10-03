export function rgwRolePolicies(value: unknown) {
  if (!Array.isArray(value)) return undefined
  if (value.some(policy => !policy || typeof policy !== 'object' || Array.isArray(policy)
    || typeof policy.PolicyName !== 'string' || policy.PolicyName === '' || typeof policy.PolicyValue !== 'string')) return undefined
  return value.map((policy, index) => ({ id: index, name: policy.PolicyName as string, document: policy.PolicyValue as string }))
}
