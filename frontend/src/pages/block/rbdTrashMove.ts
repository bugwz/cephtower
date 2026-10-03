export function rbdTrashMoveReason(row: Record<string, unknown>): string | undefined {
  if (row.format === 1) return '格式 1 镜像不支持移入回收站'
  if (row.format !== 2) return '镜像格式未知，请刷新信息后再移入回收站'
  return undefined
}
