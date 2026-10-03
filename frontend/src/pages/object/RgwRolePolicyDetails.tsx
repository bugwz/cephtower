import { rgwRolePolicies } from './rgwRolePolicies'

export function RgwPolicyDocument({ value }: { value: unknown }) {
  if (typeof value !== 'string') return <span>策略文档未返回或格式无效</span>
  return <details><summary>查看策略原文</summary>
    <pre style={{ margin: 0, maxHeight: 320, maxWidth: 600, overflow: 'auto', whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value === '' ? '（空字符串）' : value}</pre>
  </details>
}

export function RgwRolePolicyDetails({ value }: { value: unknown }) {
  const policies = rgwRolePolicies(value)
  if (!policies) return <span>内联策略未返回或格式无效</span>
  if (policies.length === 0) return <span>内联策略列表为空</span>
  return <div>{policies.map(policy => <details key={policy.id}>
    <summary>{policy.name}</summary><RgwPolicyDocument value={policy.document} />
  </details>)}</div>
}
