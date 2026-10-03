export function groupSnapshotOptions(row: Record<string, unknown> | undefined, action: unknown) {
  if (action === 'create' || !action) return []
  if (!Array.isArray(row?.snapshots)) throw new Error('组快照信息不可用，请刷新后再操作')
  const names = new Set<string>()
  const options: { label: string; value: string }[] = []
  for (const snapshot of row.snapshots) {
    if (!snapshot || typeof snapshot !== 'object' || typeof snapshot.snapshot !== 'string' || !snapshot.snapshot || names.has(snapshot.snapshot)) throw new Error('组快照信息不完整或名称重复，请重新采集')
    names.add(snapshot.snapshot)
    if (action === 'rollback' && snapshot.state !== 'complete') continue
    options.push({ label: snapshot.snapshot, value: snapshot.snapshot })
  }
  return options
}

export function groupSnapshotName(values: Record<string, unknown>, row?: Record<string, unknown>): string {
  if (values.action === 'create') return String(values.name ?? '')
  if (!['remove', 'rename', 'rollback'].includes(String(values.action))) throw new Error('请选择有效的组快照操作')
  if (typeof values.existing_name !== 'string' || !groupSnapshotOptions(row, values.action).some((option) => option.value === values.existing_name)) throw new Error('请选择当前组内可操作的快照；回滚仅支持已完成快照')
  return values.existing_name
}
