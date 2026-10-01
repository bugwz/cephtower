export interface SnapshotDependencies {
  pending: 'yes' | 'no' | 'unknown'
  clones: Array<{ name: string; group: string }> | undefined
  orphans: number | undefined
}

export function snapshotDependencies(row: Record<string, unknown>): SnapshotDependencies {
  const pending = row.has_pending_clones === 'yes' ? 'yes' : row.has_pending_clones === 'no' ? 'no' : 'unknown'
  let clones: SnapshotDependencies['clones']
  if (Array.isArray(row.pending_clones)) {
    const parsed = row.pending_clones.map((value) => {
      if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
      const item = value as Record<string, unknown>
      if (typeof item.name !== 'string' || !item.name || (item.target_group !== undefined && (typeof item.target_group !== 'string' || !item.target_group))) return undefined
      return { name: item.name, group: item.target_group === undefined ? '_nogroup' : item.target_group as string }
    })
    if (parsed.every((value) => value !== undefined)) clones = parsed as SnapshotDependencies['clones']
  } else if (row.pending_clones === undefined && pending === 'no') clones = []
  const count = row.orphan_clones_count
  const orphans = typeof count === 'number' && Number.isSafeInteger(count) && count >= 0 ? count : count === undefined && pending !== 'unknown' ? 0 : undefined
  return { pending, clones, orphans }
}

export function snapshotDeleteReason(row: Record<string, unknown>): string | undefined {
  const dependencies = snapshotDependencies(row)
  if (dependencies.pending === 'yes' || (dependencies.clones?.length ?? 0) > 0) return '快照存在待处理克隆，请等待克隆完成或先取消克隆并刷新状态'
  if (dependencies.orphans !== undefined && dependencies.orphans > 0) return '快照存在孤儿克隆记录，请检查集群克隆索引并刷新状态'
  return undefined
}

export function snapshotPendingText(value: unknown): string {
  return value === 'yes' ? '存在待处理克隆' : value === 'no' ? '无待处理克隆' : '未知'
}
