export function rbdImageDeleteReason(row: Record<string, unknown>): string | undefined {
  if (row.primary === false) return '非主镜像不能直接删除，请先处理镜像同步关系'
  if (row.has_snapshot_children === true) return '快照仍有克隆子镜像（含回收站），请先解除依赖'
  if (row.has_snapshot_children !== false) return '快照子镜像依赖未完整采集，请刷新后再删除'
  return undefined
}
