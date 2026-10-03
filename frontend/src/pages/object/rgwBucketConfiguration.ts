export const rgwBucketConfigurationOptions = [
  { label: 'Policy（JSON）', value: 'policy' },
  { label: 'CORS（XML）', value: 'cors' },
  { label: 'Lifecycle（XML）', value: 'lifecycle' },
  { label: 'Encryption（XML）', value: 'encryption' }
]

export function rgwBucketConfigurationDeleteBlocked(row: Record<string, unknown>) {
  if (row.configured !== true) return '只有已确认存在的配置可以删除，请先刷新'
  if (typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) return 'Bucket ID 缺失或无效'
  if (!rgwBucketConfigurationOptions.some(option => option.value === row.kind)) return '配置类型缺失或无效'
  return undefined
}

export function rgwBucketConfigurationDeleteInput(row: Record<string, unknown>) {
  const blocked = rgwBucketConfigurationDeleteBlocked(row)
  if (blocked) throw new Error(blocked)
  return { bucket_id: row.bucket_id as string, kind: row.kind as string }
}

export function rgwBucketConfigurationDeleteConfirmation(row: Record<string, unknown>) {
  const { bucket_id, kind } = rgwBucketConfigurationDeleteInput(row)
  const impact: Record<string, string> = {
    policy: '将移除 Bucket Policy 中的允许和拒绝规则，访问权限可能变化。',
    cors: '将移除全部 CORS 规则，浏览器跨域访问可能失败。',
    lifecycle: '将移除全部生命周期规则，不会恢复已过期或已迁移的对象。',
    encryption: '将移除 Bucket 默认加密配置，不会解密已有对象；后续写入的加密行为由请求与集群设置决定。'
  }
  return `确认删除 Bucket ID ${bucket_id} 的 ${kind} 完整配置？${impact[kind]}不会删除 Bucket 或对象。请先保存原始文档，删除后没有自动撤销；外部并发修改也可能被删除。`
}

export function rgwBucketConfigurationInput(values: Record<string, unknown>) {
  const { kind, document, bucket_id } = values
  if (typeof bucket_id !== 'string' || !bucket_id || bucket_id.trim() !== bucket_id) throw new Error('请输入准确的 Bucket ID')
  if (!rgwBucketConfigurationOptions.some(option => option.value === kind)) throw new Error('请选择配置类型')
  if (typeof document !== 'string' || !document.trim()) throw new Error('请输入完整配置文档')
  // API JSON bodies are capped at 1 MiB; reserve room for cluster_id and framing.
  if (new TextEncoder().encode(JSON.stringify({ bucket_id, kind, document })).length > 1024 * 1024 - 128) throw new Error('配置文档编码后的请求超出 1 MiB 上限')
  if (kind === 'policy') {
    const policy = JSON.parse(document)
    if (!policy || typeof policy !== 'object' || Array.isArray(policy)) throw new Error('Policy 必须是 JSON 对象')
  } else if (!document.trimStart().startsWith('<')) throw new Error('此配置类型必须提供 XML 文档，后端将校验其结构')
  return { bucket_id, kind, document }
}
