type Row = Record<string, any>
const token = (value: unknown): value is string => typeof value === 'string' && !!value && !value.startsWith('-') && !/[\x00-\x1f\x7f]/.test(value)
function targets(row?: Row) {
  if (!row || row.stale === true || !token(row.id) || !token(row.name) || typeof row.realm_id !== 'string' || typeof row.default_placement !== 'string') throw new Error('组身份、默认放置目标未知或库存已过期')
  const values = row.placement_targets
  if (!Array.isArray(values) || !values.length || values.some(p => !token(p?.name) || !Array.isArray(p.storage_classes) || p.storage_classes.some((c: unknown) => !token(c)) || new Set(p.storage_classes).size !== p.storage_classes.length) || new Set(values.map(p => p.name)).size !== values.length) throw new Error('已有放置目标或存储类清单不完整')
  return values as Row[]
}
export function groupStorageClassBlocked(row: Row) { try { targets(row); return undefined } catch (error) { return (error as Error).message } }
export function groupStorageClassOptions(row?: Row) { return targets(row).filter(p => !p.name.includes('/')).map(p => ({value:p.name,label:p.name})) }
export function groupStorageClassInput(values: Row, row?: Row) {
  const target = targets(row).find(p => p.name === values.placement_id)
  if (!target || target.name.includes('/') || !token(values.storage_class) || target.storage_classes.includes(values.storage_class)) throw new Error('必须选择已有目标和未声明的新存储类')
  if (target.tier_targets !== undefined && (!Array.isArray(target.tier_targets) || target.tier_targets.some((t: Row) => !t || !token(t.key) || t.key === values.storage_class))) throw new Error('云分层声明冲突或配置未知')
  if (values.confirm_create !== 'acknowledged') throw new Error('请确认声明范围与风险')
  return {zonegroup_id:row!.id,name:row!.name,realm_id:row!.realm_id,placement_id:values.placement_id,storage_class:values.storage_class,expected_default_placement:row!.default_placement,confirm_create:true}
}
export function groupStorageClassConfirmation(values: Row, row?: Row) {
  const p = groupStorageClassInput(values,row)
  return `在 Zonegroup ${p.name}（${p.zonegroup_id}）的已有放置目标 ${p.placement_id} 声明普通存储类 ${p.storage_class}。${p.expected_default_placement === '' ? `当前没有默认放置目标，原生操作会把 ${p.placement_id} 的 STANDARD 设为默认。` : `保留默认放置目标 ${p.expected_default_placement}。`}保留标签、已有存储类和云分层配置。此次只写组本地声明，不配置各 Zone 的池，不迁移数据、不发布 Period 或重启网关；请先完成相关 Zone 的存储类池配置，再评估并单独发布 Realm Period。写前核验不是跨进程原子锁，失败可能部分生效，不自动重试或回滚。`
}
