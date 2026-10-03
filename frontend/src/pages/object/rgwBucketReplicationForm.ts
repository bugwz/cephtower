export function bucketReplicationFormBlocked(row: Record<string, unknown>) {
  if (row.kind !== 'replication' || typeof row.bucket_id !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.bucket_id)) return '请先读取目标 Bucket 复制配置'
  if (row.configured !== true && row.configured !== false) return '复制配置状态未知，请刷新'
  if (row.configured === true && (typeof row.document !== 'string' || !row.document.trim())) return '当前复制文档不可用'
  if (row.configured === false && row.document !== null) return '复制配置状态与文档不一致'
  return undefined
}
export function bucketReplicationFormInitial(row?: Record<string, unknown>) {
  if (!row) throw new Error('请选择复制配置记录')
  const blocked = bucketReplicationFormBlocked(row)
  if (blocked) throw new Error(blocked)
  return { bucket_id: row.bucket_id as string, confirm_replication: undefined }
}
export function bucketReplicationFormInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const initial = bucketReplicationFormInitial(row)
  if (values.bucket_id !== initial.bucket_id) throw new Error('Bucket ID 不可更改')
  if (values.confirm_replication !== 'acknowledged') throw new Error('请确认上层策略及替换完整复制规则的影响')
  const input = { bucket_id: initial.bucket_id, expected_document: row!.configured ? row!.document as string : '' }
  if (new TextEncoder().encode(JSON.stringify(input)).length > 1024 * 1024 - 128) throw new Error('原始复制快照超出请求上限')
  return input
}
export function bucketReplicationFormConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const { bucket_id } = bucketReplicationFormInput(values, row)
  return `确认替换 Bucket ID ${bucket_id} 的全部 S3 复制规则为 Dashboard 同名桶复制规则 dashboard_admin_pipe（Enabled、优先级 0、全部对象、全部 Zone）？请先在目标 Zonegroup 准备并发布上层复制策略；本操作不会代替上层配置，也不证明复制正在运行或已完成。目标租户必须与 S3 永久用户凭据一致，需要 ListBuckets 和复制配置读写权限；Account/IAM 身份受 Ceph 原生限制。原有过滤、目标及其他 S3 复制规则将被移除，不修改其他本地同步组，不删除已复制对象。请先备份原文并避免跨页面或外部并发修改；快照检查不是原子锁，失败可能已写入，没有自动回滚或重试。`
}
