export function bucketNotificationDeleteBlocked(row: Record<string, unknown>) {
  if (row.kind !== 'notification' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) return '请先读取目标 Bucket 通知配置'
  if (row.configured !== true || typeof row.document !== 'string' || !row.document.trim() || !Array.isArray(row.notifications) || row.notifications.length === 0 || row.notifications.some(rule => !rule || typeof rule.id !== 'string')) return '没有已确认的通知规则，请刷新'
  return undefined
}
export function bucketNotificationDeleteInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择通知配置记录')
  const blocked = bucketNotificationDeleteBlocked(row)
  if (blocked) throw new Error(blocked)
  return { bucket_id: row.bucket_id as string, mode: undefined, notification_id: '', confirm_notification: undefined }
}
export function bucketNotificationDeleteInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = bucketNotificationDeleteInitial(row)
  if (values.bucket_id !== initial.bucket_id) throw new Error('Bucket ID 不可更改')
  if (values.confirm_notification !== 'acknowledged') throw new Error('请确认停止通知及不可撤销影响')
  const id = values.notification_id ?? ''
  if (typeof id !== 'string') throw new Error('通知 ID 必须为原始字符串')
  if (values.mode === 'single') {
    if (!id || /[\0\r\n]/.test(id) || (row!.notifications as { id: string }[]).filter(rule => rule.id === id).length !== 1) throw new Error('指定 ID 必须精确匹配唯一通知，不得为空；重复 ID 不能单条删除')
  } else if (values.mode !== 'all' || id !== '') throw new Error('请选择删除范围；全部删除时必须清空通知 ID')
  const input = { bucket_id: initial.bucket_id, mode: values.mode, notification_id: id, expected_document: row!.document as string }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('通知快照超出请求上限')
  return input
}
export function bucketNotificationDeleteConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketNotificationDeleteInput(values, row)
  return `确认删除 Bucket ID ${input.bucket_id} 的${input.mode === 'all' ? '全部通知规则' : `指定通知 ${JSON.stringify(input.notification_id)}`}？将停止这些规则的后续事件订阅，不删除 Bucket、对象或独立 Topic，不保证已排队消息停止投递或 Topic/Bucket 映射已清理。请备份完整 XML；快照检查不是原子锁，外部并发修改可能被删除。失败可能已部分生效，不自动回滚或重试。`
}
