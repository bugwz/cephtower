export function rgwBucketObjectLockSummary(value: unknown, row: Record<string, unknown>): string {
  if (row.kind !== 'object-lock') return '—'
  if (row.configured === false && value === null) return '未启用 Bucket 对象锁（原生配置不存在）'
  if (row.configured !== true || !value || typeof value !== 'object' || Array.isArray(value)) return '对象锁配置不可用'
  const configuration = value as Record<string, unknown>
  if (configuration.enabled !== true) return '对象锁配置不可用'
  const retention = configuration.default_retention
  if (retention === null) return '对象锁已启用；未设置默认保留期（对象仍可能有独立保留或 Legal Hold）'
  if (!retention || typeof retention !== 'object' || Array.isArray(retention)) return '默认保留策略不可用'
  const fields = retention as Record<string, unknown>
  if (typeof fields.mode !== 'string' || !fields.mode || (fields.days !== null && typeof fields.days !== 'string') || (fields.years !== null && typeof fields.years !== 'string') || (fields.days === null) === (fields.years === null)) return '默认保留策略不可用'
  const mode = fields.mode === 'GOVERNANCE' ? '治理模式 GOVERNANCE' : fields.mode === 'COMPLIANCE' ? '合规模式 COMPLIANCE' : `未知模式 ${JSON.stringify(fields.mode)}`
  const amount = fields.days ?? fields.years
  const period = typeof amount === 'string' && /^[0-9]+$/.test(amount) && BigInt(amount) > 0n ? `${amount} ${fields.days !== null ? '天' : '年'}` : `无效期限 ${JSON.stringify(amount)}`
  return `对象锁已启用；${mode}；默认保留 ${period}。默认策略不代表已有对象的实际保留状态。`
}
