const token = (value: unknown): value is string => typeof value === 'string' && !!value && !value.startsWith('-') && !/[\x00-\x1f\x7f]/.test(value)
export function zonegroupDeleteBlocked(row: Record<string, unknown>) {
  if (row.stale === true) return 'Zonegroup 库存已过期，请重新采集'
  if (!token(row.id) || !token(row.name) || typeof row.realm_id !== 'string' || row.realm_id !== '' && !token(row.realm_id)) return 'Zonegroup 身份或 Realm 归属不完整'
  if (row.is_default !== false || row.is_master !== false) return '请先调整主组和当前上下文默认组；状态未知时不能删除'
  if (!Array.isArray(row.zones) || row.zones.some(zone => !zone || typeof zone !== 'object' || !token(zone.id)) || new Set(row.zones.map(zone => zone.id)).size !== row.zones.length) return 'Zone 成员清单不完整或有重复'
  return undefined
}
export function zonegroupDeleteInput(row: Record<string, unknown>) {
  const blocked = zonegroupDeleteBlocked(row)
  if (blocked) throw new Error(blocked)
  return { zonegroup_id: row.id, name: row.name, realm_id: row.realm_id, expected_zones: (row.zones as Array<{id: string}>).map(zone => zone.id), confirm_delete: true }
}
export function zonegroupDeleteConfirmation(row: Record<string, unknown>) {
  const input = zonegroupDeleteInput(row)
  return `确认删除 Zonegroup ${row.name}（${row.id}），保留成员 Zone IDs ${JSON.stringify(input.expected_zones)}？不删除任何 Zone、存储池、桶或对象，不移除网关服务，也不自动重启。请先备份并停止依赖此组的业务，检查其他 Realm 的默认组引用；原生 default_info 仅覆盖当前默认 Realm 上下文，其他默认引用可能残留。${row.realm_id ? `将发布 Realm ${row.realm_id} 的 Period，可能同时发布其他待提交变更，成功不代表远端同步完成。` : '无 Realm，不发布 Period；运行中的网关不会因此自动重载。'}主组、当前上下文默认组及唯一组不可删除。身份与成员检查不是跨进程原子锁；删除和发布分步执行，失败可能部分生效，不自动回滚或重试。此入口不包含可选的 Zone/池清理。`
}
