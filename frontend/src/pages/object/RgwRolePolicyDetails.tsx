import { rgwRolePolicies } from './rgwRolePolicies'
import { rgwIdentityList } from './rgwUserIdentity'

export function RgwRoleManagedPolicies({ value }: { value: unknown }) {
  if (value === undefined) return <span>未返回托管策略字段（原生命令在无关联时省略）；不能据此判断完整有效权限</span>
  const policies = rgwIdentityList(value)
  if (!policies) return <span>托管策略格式无效</span>
  if (policies.length === 0) return <span>托管策略列表为空</span>
  return <ul style={{ margin: 0, paddingLeft: 18, overflowWrap: 'anywhere' }}>{policies.map((arn, index) => <li key={index}>{arn}</li>)}</ul>
}

export function RgwPolicyDocument({ value }: { value: unknown }) {
  if (typeof value !== 'string') return <span>策略文档未返回或格式无效</span>
  return <details><summary>查看策略原文</summary>
    <pre style={{ margin: 0, maxHeight: 320, maxWidth: 600, overflow: 'auto', whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' }}>{value === '' ? '（空字符串）' : value}</pre>
  </details>
}

export function RgwRolePolicyDetails({ value }: { value: unknown }) {
  if (value === undefined) return <span>未返回内联策略字段（原生命令在无内联策略时省略）；不能据此判断完整有效权限</span>
  const policies = rgwRolePolicies(value)
  if (!policies) return <span>内联策略格式无效</span>
  if (policies.length === 0) return <span>内联策略列表为空</span>
  return <div>{policies.map(policy => <details key={policy.id}>
    <summary>{policy.name}</summary><RgwPolicyDocument value={policy.document} />
  </details>)}</div>
}
