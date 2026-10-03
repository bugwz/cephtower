export const rgwBucketConfigurationOptions = [
  { label: 'Policy（JSON）', value: 'policy' },
  { label: 'CORS（XML）', value: 'cors' },
  { label: 'Lifecycle（XML）', value: 'lifecycle' },
  { label: 'Encryption（XML）', value: 'encryption' },
  { label: 'Tags 标签（XML，最多 50 条）', value: 'tagging' }
]
export const rgwBucketConfigurationReadOptions = [...rgwBucketConfigurationOptions, { label: 'Object Lock 默认保留', value: 'object-lock' }, { label: 'ACL 访问控制列表', value: 'acl' }]

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
    encryption: '将移除 Bucket 默认加密配置，不会解密已有对象；后续写入的加密行为由请求与集群设置决定。',
    tagging: '将移除 Bucket 的全部标签属性，可能影响依赖 Bucket 标签条件的访问权限；不改变对象标签。'
  }
  return `确认删除 Bucket ID ${bucket_id} 的 ${kind} 完整配置？${impact[kind]}不会删除 Bucket 或对象。请先保存原始文档，删除后没有自动撤销；外部并发修改也可能被删除。`
}

export function rgwBucketConfigurationEditBlocked(row: Record<string, unknown>) {
  if (typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) return 'Bucket ID 缺失或无效'
  if (!rgwBucketConfigurationOptions.some(option => option.value === row.kind)) return '配置类型缺失或无效'
  if (row.configured !== true && row.configured !== false) return '配置状态未知，请刷新后重试'
  if (row.configured && (typeof row.document !== 'string' || !row.document.trim())) return '当前配置文档不可用，请刷新后重试'
  if (!row.configured && row.document !== null) return '未配置状态与文档不一致，请刷新后重试'
  return undefined
}

export function rgwBucketConfigurationEditInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择配置记录')
  const blocked = rgwBucketConfigurationEditBlocked(row)
  if (blocked) throw new Error(blocked)
  return { bucket_id: row.bucket_id as string, kind: row.kind as string, document: row.configured ? row.document as string : '' }
}

export function rgwBucketConfigurationEditInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = rgwBucketConfigurationEditInitial(row)
  if (values.bucket_id !== initial.bucket_id || values.kind !== initial.kind) throw new Error('不能更改编辑记录的 Bucket ID 或配置类型')
  return rgwBucketConfigurationInput(values)
}

export function rgwBucketConfigurationUpdateConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { bucket_id, kind } = row ? rgwBucketConfigurationEditInput(values, row) : rgwBucketConfigurationInput(values)
  return `确认向 Bucket ID ${bucket_id} 提交 ${kind} 完整配置？这会整体替换该类型的已有配置，不是合并修改；未保留的规则将被移除。Policy 影响访问权限，CORS 影响跨域访问，Lifecycle 可能导致对象过期删除或迁移，Encryption 影响后续写入的默认加密，Tags 可能影响依赖 Bucket 标签条件的访问权限（不修改对象标签）。请先备份原始文档；外部并发修改可能被覆盖，提交不会自动回滚。`
}

export function rgwBucketConfigurationInput(values: Record<string, unknown>) {
  const { kind, document, bucket_id } = values
  if (typeof bucket_id !== 'string' || !bucket_id || bucket_id.trim() !== bucket_id) throw new Error('请输入准确的 Bucket ID')
  if (!rgwBucketConfigurationOptions.some(option => option.value === kind)) throw new Error('请选择配置类型')
  if (typeof document !== 'string' || !document.trim()) throw new Error('请输入完整配置文档')
  // API JSON bodies are capped at 1 MiB; reserve room for cluster_id and framing.
  if (new TextEncoder().encode(JSON.stringify({ bucket_id, kind, document })).length > 1024 * 1024 - 128) throw new Error('配置文档编码后的请求超出 1 MiB 上限')
  if (kind === 'policy') {
    let policy: unknown
    try { policy = JSON.parse(document) } catch { throw new Error('Policy 必须是有效的 JSON 文档') }
    if (!policy || typeof policy !== 'object' || Array.isArray(policy)) throw new Error('Policy 必须是 JSON 对象')
  } else if (!document.trimStart().startsWith('<')) throw new Error('此配置类型必须提供 XML 文档，后端将校验其结构')
  return { bucket_id, kind, document }
}
