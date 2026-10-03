export const rgwBucketConfigurationOptions = [
  { label: 'Policy（JSON）', value: 'policy' },
  { label: 'CORS（XML）', value: 'cors' },
  { label: 'Lifecycle（XML）', value: 'lifecycle' },
  { label: 'Encryption（XML）', value: 'encryption' }
]

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
