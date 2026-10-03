function groups(row?: Record<string, unknown>) {
  const policy = row?.bucket_sync_policy as { groups?: unknown } | undefined
  if (!row || row.stale === true || typeof row.natural_key !== 'string' || !/^[A-Za-z0-9_-]+$/.test(row.natural_key) || !policy || !Array.isArray(policy.groups) || policy.groups.length === 0) throw new Error('请刷新并选择含同步组的 Bucket')
  const seen = new Set<string>()
  for (const group of policy.groups) {
    if (!group || typeof group !== 'object' || typeof group.id !== 'string' || !group.id || typeof group.status !== 'string' || !group.status || seen.has(group.id)) throw new Error('同步组数据不可用')
    seen.add(group.id)
  }
  return policy.groups as { id: string; status: string }[]
}
export function bucketSyncGroupBlocked(row: Record<string, unknown>) {
  try { groups(row); return undefined } catch (error) { return (error as Error).message }
}
export function bucketSyncGroupInitial(row?: Record<string, unknown>) {
  groups(row)
  return { bucket_id: row!.natural_key as string, group_id: undefined, status: undefined, confirm_change: undefined }
}
export function bucketSyncGroupInput(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const available = groups(row)
  if (values.bucket_id !== row!.natural_key) throw new Error('Bucket ID 不可更改')
  const group = available.find(group => group.id === values.group_id)
  if (!group) throw new Error('请输入当前策略中准确的同步组 ID')
  if (!['enabled', 'allowed', 'forbidden'].includes(values.status as string)) throw new Error('请明确选择目标状态')
  if (values.status === group.status) throw new Error('同步组状态未变化')
  if (values.confirm_change !== 'acknowledged') throw new Error('请确认复制影响')
  return { bucket_id: row!.natural_key as string, group_id: group.id, expected_status: group.status, status: values.status as string }
}
export function bucketSyncGroupConfirmation(values: Record<string, unknown>, row?: Record<string, unknown>) {
  const input = bucketSyncGroupInput(values, row)
  return `确认修改 Bucket ID ${input.bucket_id} 的同步组 ${JSON.stringify(input.group_id)}：${input.expected_status} → ${input.status}？enabled 启用、allowed 仅允许但不启用、forbidden 禁止；可能改变后续复制行为。不会新建数据流或管道、不修改 Zonegroup 策略或提交 period。最终同步效果仍取决于上层策略和拓扑；已有副本不会删除。请备份策略，并避免外部并发修改；核验失败不代表未生效，操作不会自动回滚。`
}
