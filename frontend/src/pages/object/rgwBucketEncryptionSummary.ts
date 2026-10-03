export function rgwBucketEncryptionSummary(value: unknown, row: Record<string, unknown>): string {
  if (row.kind !== 'encryption') return '—'
  if (row.configured === false && value === null) return '未设置 Bucket 默认加密配置'
  if (row.configured !== true || !value || typeof value !== 'object') return '加密数据不可用'
  const config = value as Record<string, unknown>
  if (typeof config.rule_exists !== 'boolean' || typeof config.algorithm !== 'string' || typeof config.kms_master_key_id !== 'string' || typeof config.bucket_key_enabled !== 'boolean') return '加密数据不可用'
  if (!config.rule_exists) return '配置已保存，但没有默认加密规则'
  return `算法：${config.algorithm || '（空）'}；KMS Key ID：${JSON.stringify(config.kms_master_key_id)}；Bucket Key：${config.bucket_key_enabled ? '启用' : '未启用'}。仅描述 Bucket 默认配置，不代表已有对象的加密状态。`
}
