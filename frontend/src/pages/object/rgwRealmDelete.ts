export function realmDeleteBlocked(row: Record<string, unknown>) {
  if (row.stale === true) return 'Realm 库存已过期，请重新采集'
  if (row.is_default !== false) return '请先将其他 Realm 设为默认，并重新采集确认；默认状态未知时不可删除'
  for (const key of ['id', 'name', 'current_period']) {
    if (typeof row[key] !== 'string' || !row[key] || row[key] !== row[key].trim() || row[key].startsWith('-') || /[\s\x00]/.test(row[key])) return 'Realm 身份或当前 Period 不完整'
  }
  return undefined
}
export function realmDeleteInput(row: Record<string, unknown>) {
  const blocked = realmDeleteBlocked(row)
  if (blocked) throw new Error(blocked)
  return { realm_id: row.id, name: row.name, expected_current_period: row.current_period, confirm_delete: true }
}
export function realmDeleteConfirmation(row: Record<string, unknown>) {
  realmDeleteInput(row)
  return `确认删除 Realm ${row.name}（ID ${row.id}，当前 Period ${row.current_period}）？请先备份配置并停止依赖此 Realm 的业务。此操作仅移除 Realm 配置，不删除 Zonegroup、Zone、Period、网关服务、用户、桶或对象，不自动清理它们的 Realm 引用，也不传播为远端 Realm 删除。残留引用可能导致网关或复制不可用。默认 Realm 必须先切换。身份及 Period 检查不是原子锁，外部并发仍可能改变配置；失败可能部分生效，不自动重试或回滚。成功只证明 Realm ID 与名称索引不再存在，不证明所有关联引用已清理。`
}
