const token = (value: unknown): value is string => typeof value === 'string' && !!value && !value.startsWith('-') && !/[\x00-\x1f\x7f]/.test(value)
export function zoneDeleteBlocked(row: Record<string, unknown>) {
  if (row.stale === true) return 'Zone 库存已过期，请重新采集'
  if (!token(row.id) || !token(row.name) || typeof row.realm_id !== 'string' || row.realm_id !== '' && !token(row.realm_id)) return 'Zone 身份或 Realm 归属不完整'
  if (row.is_default !== false) return '请先调整当前默认 Zone；默认状态未知时不能删除'
  return undefined
}
export function zoneDeleteInput(row: Record<string, unknown>) {
  const blocked = zoneDeleteBlocked(row)
  if (blocked) throw new Error(blocked)
  return { zone_id: row.id, name: row.name, realm_id: row.realm_id, confirm_delete: true }
}
export function zoneDeleteConfirmation(row: Record<string, unknown>) {
  zoneDeleteInput(row)
  return `确认删除 Zone ${row.name}（${row.id}）及所有 Zonegroup 中对应的成员？保留所有池、桶、对象和网关服务，不自动重启网关。请先备份并停止依赖此 Zone 的业务；主 Zone、当前默认 Zone 和唯一 Zone 不可删除。成员移除会按原生规则调整剩余成员 log_data。默认状态只覆盖当前默认 Realm 上下文，其他默认引用可能残留，请事先核对。${row.realm_id ? `将发布 Realm ${row.realm_id} 的 Period，可能同时发布其他待提交变更；成功不代表远端同步完成。` : '无 Realm，不发布 Period。'}核验不是跨进程原子锁，删除与发布非事务；失败可能部分生效，不自动回滚或重试。此入口不包含可选池清理。`
}
