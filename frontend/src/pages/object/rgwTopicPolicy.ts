export function topicIdentityBlocked(row: Record<string, unknown>) {
  if (row.stale === true) return 'Topic 库存已过期，请刷新'
  if (typeof row.scope !== 'string' || typeof row.name !== 'string' || !row.name || row.metadata_key !== `${row.scope}:${row.name}`) return 'Topic 身份不完整'
  const id = btoa(Array.from(new TextEncoder().encode(String(row.metadata_key)), byte => String.fromCharCode(byte)).join('')).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  if (row.natural_key !== id) return 'Topic 编码身份不一致'
  const arn = typeof row.arn === 'string' ? row.arn.split(':') : []
  if (arn.length !== 6 || arn[0] !== 'arn' || arn[1] !== 'aws' || arn[2] !== 'sns' || !arn[3] || arn[4] !== row.scope || arn[5] !== row.name) return 'Topic ARN 与身份不一致'
  return undefined
}
export function topicPolicyBlocked(row: Record<string, unknown>) {
  return topicIdentityBlocked(row) ?? (typeof row.policy !== 'string' ? '当前 Policy 未返回，请刷新' : undefined)
}
export function topicPolicyInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择 Topic')
  const blocked = topicPolicyBlocked(row)
  if (blocked) throw new Error(blocked)
  return { topic_id: row.natural_key as string, topic_arn: row.arn as string, policy: row.policy as string }
}
export function topicPolicyInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择 Topic')
  const blocked = topicPolicyBlocked(row)
  if (blocked) throw new Error(blocked)
  if (values.topic_id !== row.natural_key || values.topic_arn !== row.arn) throw new Error('不可更改 Topic 身份')
  if (values.policy_mode !== 'set' && values.policy_mode !== 'clear') throw new Error('请选择设置或清除 Policy')
  const policy = values.policy_mode === 'clear' ? '' : values.policy
  if (typeof policy !== 'string') throw new Error('Policy 必须是 JSON 对象')
  if (values.policy_mode === 'set') {
    let parsed: unknown
    try { parsed = JSON.parse(policy) } catch { throw new Error('Policy 必须是 JSON 对象') }
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('Policy 必须是 JSON 对象')
  }
  if (policy === row.policy) throw new Error('Policy 未发生变化')
  const input = { topic_id: row.natural_key, topic_arn: row.arn, expected_policy: row.policy, policy }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('Policy 请求超过大小限制')
  return input
}
export function topicPolicyConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  topicPolicyInput(values, row)
  return '确认替换或清除完整 Topic Policy？访问权限可能立即改变，清除不保证私有访问。需要 HTTPS RGW S3 端点及 SNS 读取、修改权限；策略可能导致回读权限丢失，此时操作可能已生效。不会修改推送端点、队列或桶通知规则；快照核验不是原子锁，不会自动回滚或重试。'
}
