export function rbdMirrorRoleReason(action: unknown, row: Record<string, unknown>): string | undefined {
  if (!['mirror-promote', 'mirror-demote', 'mirror-resync', 'mirror-snapshot'].includes(String(action))) return undefined
  if (row.mirror_state !== 'enabled') return '未确认镜像同步已启用，请刷新镜像信息后再执行角色操作'
  if (typeof row.primary !== 'boolean') return '主从角色未知，请刷新镜像信息'
  if (action === 'mirror-snapshot') {
    if (row.mirror_mode !== 'snapshot') return '创建同步快照要求启用快照同步模式'
    if (!row.primary) return '仅主镜像可以创建同步快照'
  }
  if (action === 'mirror-promote' && row.primary) return '镜像已是主镜像，无需提升'
  if (action === 'mirror-demote' && !row.primary) return '镜像已是非主镜像，无需降级'
  if (action === 'mirror-resync' && row.primary) return '主镜像不能从远端重新同步，请先确认主从角色'
  return undefined
}
