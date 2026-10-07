export function mirrorPoolIdentity(row?: Record<string, unknown>): string {
  const pool = row?.pool
  if (typeof pool !== 'string' || !pool || pool.trim() !== pool || /[\r\n\0]/.test(pool)) throw new Error('池身份不可用，请重新采集')
  return pool
}

export function mirrorPoolModeBody(values: Record<string, unknown>, clusterId: number, row?: Record<string, unknown>) {
  const pool = mirrorPoolIdentity(row)
  if (typeof values.mode !== 'string' || !['disabled', 'image', 'pool'].includes(values.mode)) throw new Error('请选择有效的同步模式')
  if (values.mode === 'disabled' && row?.mode !== 'disabled') {
    if (!Array.isArray(row?.peers)) throw new Error('Peer 库存未读取，无法确认可禁用，请重新采集')
    if (row.peers.length) throw new Error('请先删除池中的 Peer，再禁用同步')
  }
  return { cluster_id: clusterId, pool, mode: String(values.mode) }
}
