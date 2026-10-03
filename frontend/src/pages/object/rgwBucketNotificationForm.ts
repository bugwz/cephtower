export type NotificationRule = { id: string; topic: string; events: string[]; filters: { kind: string; name: string; value: string }[] }
export type NotificationDraft = { mode: string; selected: string; rule: NotificationRule; existing: NotificationRule[] }
export const notificationEvents = [
  's3:ObjectCreated:*', 's3:ObjectCreated:Put', 's3:ObjectCreated:Post', 's3:ObjectCreated:Copy', 's3:ObjectCreated:CompleteMultipartUpload',
  's3:ObjectRemoved:*', 's3:ObjectRemoved:Delete', 's3:ObjectRemoved:DeleteMarkerCreated',
  's3:ObjectLifecycle:*', 's3:ObjectLifecycle:Expiration:*', 's3:ObjectLifecycle:Expiration:Current', 's3:ObjectLifecycle:Expiration:NonCurrent', 's3:ObjectLifecycle:Expiration:DeleteMarker', 's3:ObjectLifecycle:Expiration:AbortMultipartUpload',
  's3:ObjectLifecycle:Transition:*', 's3:ObjectLifecycle:Transition:Current', 's3:ObjectLifecycle:Transition:NonCurrent',
  's3:ObjectSynced:*', 's3:ObjectSynced:Create', 's3:ObjectSynced:Delete', 's3:ObjectSynced:DeletionMarkerCreated',
  's3:LifecycleExpiration:*', 's3:LifecycleExpiration:Delete', 's3:LifecycleExpiration:DeleteMarkerCreated', 's3:LifecycleTransition',
  's3:Replication:*', 's3:Replication:Create', 's3:Replication:Delete', 's3:Replication:DeletionMarkerCreated'
]
export function notificationRuleShape(value: unknown): value is NotificationRule {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false
  const r = value as NotificationRule
  return typeof r.id === 'string' && typeof r.topic === 'string' && Array.isArray(r.events) && r.events.every(e => typeof e === 'string') && Array.isArray(r.filters) && r.filters.every(f => f && ['S3Key', 'S3Metadata', 'S3Tags'].includes(f.kind) && typeof f.name === 'string' && typeof f.value === 'string')
}
export function notificationFormBlocked(row: Record<string, unknown>) {
  if (row.kind !== 'notification' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id) || row.configured !== true || typeof row.document !== 'string' || !row.document.trim() || !Array.isArray(row.notifications) || !row.notifications.every(notificationRuleShape)) return '请先成功读取完整通知配置（允许空规则）'
  return undefined
}
export function notificationFormInitial(row?: Record<string, unknown>) {
  if (!row || notificationFormBlocked(row)) throw new Error('通知配置不可用，请刷新')
  return { bucket_id: row.bucket_id as string, notification_draft: { mode: '', selected: '', rule: { id: '', topic: '', events: [], filters: [] }, existing: structuredClone(row.notifications) } as NotificationDraft, confirm_notification: undefined }
}
export function notificationEditableRule(rule: NotificationRule): NotificationRule {
  return { ...structuredClone(rule), events: rule.events.map(event => event === 's3:ObjectLifecycle:Expiration:AbortMPU' ? 's3:ObjectLifecycle:Expiration:AbortMultipartUpload' : event) }
}
export function notificationFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = notificationFormInitial(row)
  if (values.bucket_id !== initial.bucket_id) throw new Error('Bucket ID 不可更改')
  if (values.confirm_notification !== 'acknowledged') throw new Error('请确认通知替换及非原子操作风险')
  const draft = values.notification_draft as NotificationDraft | undefined
  if (!draft || !notificationRuleShape(draft.rule) || !['create', 'edit'].includes(draft.mode)) throw new Error('请选择创建或编辑并填写通知规则')
  const rule = structuredClone(draft.rule)
  if (!rule.id || /[\r\n]/.test(rule.id) || !/^arn:aws:sns:[A-Za-z0-9_.-]+:[A-Za-z0-9_.-]*:[A-Za-z0-9_-]{1,256}$/.test(rule.topic)) throw new Error('通知 ID 不能为空或包含换行，Topic 必须为完整 ARN')
  const matching = (row!.notifications as NotificationRule[]).filter(item => item.id === rule.id)
  if (draft.mode === 'create' ? matching.length !== 0 : matching.length !== 1 || draft.selected !== rule.id) throw new Error('新 ID 必须不存在；编辑必须选择唯一既有 ID，且不能改名')
  if (rule.events.some(event => !notificationEvents.includes(event))) throw new Error('存在原生不支持的事件，请明确调整；ObjectRestore 不被此版本后端识别')
  const seen = new Set<string>()
  for (const filter of rule.filters) {
    const key = `${filter.kind}\0${filter.name}`
    if (seen.has(key) || filter.kind === 'S3Key' && !['prefix', 'suffix', 'regex'].includes(filter.name)) throw new Error('过滤名称重复或 S3Key 名称无效')
    seen.add(key)
  }
  for (const value of [rule.id, ...rule.filters.flatMap(filter => [filter.name, filter.value])]) if (/[\u0000-\u0008\u000b\u000c\u000e-\u001f\ufffe\uffff]/u.test(value) || /[\ud800-\udfff]/u.test(value)) throw new Error('通知字段包含非法 XML 字符')
  const input = { bucket_id: initial.bucket_id, mode: draft.mode, rule, expected_document: row!.document as string }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('通知配置及快照超出请求上限')
  return input
}
export function notificationFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = notificationFormInput(values, row)
  return `确认${input.mode === 'create' ? '创建' : '完整替换'} Bucket ID ${input.bucket_id} 的通知 ${JSON.stringify(input.rule.id)}，目标 ${JSON.stringify(input.rule.topic)}？此规则的事件和过滤条件将整体提交，其他规则应保持不变。空事件列表使用 Ceph 默认 ObjectCreated:* 与 ObjectRemoved:*，不代表订阅所有事件。更换 Topic 名称需要先删除旧规则再写入，可能产生通知空窗或只完成删除；新 Topic 需要 GetTopicAttributes 与 Publish 权限，Bucket 需要通知读写权限。快照不是原子锁，外部并发修改可能被覆盖；失败不自动重试或回滚。正则由 Ceph 使用 C++ 语法运行，本界面不验证其正确性或开销；配置回读不证明投递成功或 Topic 映射已更新。请先备份原始文档。`
}
