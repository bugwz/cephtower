export function topicDeleteBlocked(row: Record<string, unknown>) {
  if (row.stale === true) return 'Topic 库存已过期，请刷新'
  if (typeof row.natural_key !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.natural_key)) return 'Topic 身份不可用'
  if (typeof row.metadata_key !== 'string' || !row.metadata_key || typeof row.name !== 'string' || !row.name || typeof row.scope !== 'string' || row.metadata_key !== `${row.scope}:${row.name}`) return 'Topic 元数据身份不完整'
  const encoded = btoa(Array.from(new TextEncoder().encode(row.metadata_key), byte => String.fromCharCode(byte)).join('')).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  if (row.natural_key !== encoded) return 'Topic 编码身份与元数据键不一致'
  if (typeof row.metadata_version !== 'string' || !row.metadata_version) return '原生元数据版本不可用，请刷新'
  return undefined
}
export function topicDeleteInput(row: Record<string, unknown>) {
  const blocked = topicDeleteBlocked(row)
  if (blocked) throw new Error(blocked)
  return { topic_id: row.natural_key as string, expected_version: row.metadata_version as string }
}
export function topicDeleteConfirmation(row: Record<string, unknown>) {
  topicDeleteInput(row)
  return `确认删除通知目标 ${row.metadata_key}（${row.arn}）？原生 topic rm 会移除目标配置及其持久化队列，未投递消息可能永久丢失；不会自动清理引用此 Topic 的桶通知规则，不删除 Bucket 或对象。必须在主 Zone 执行，迁移中的 Topic 由原生拒绝。请先备份配置并检查引用；版本检查不是原子锁，删除可能部分生效或与外部修改竞争，不会自动回滚或重试。`
}
