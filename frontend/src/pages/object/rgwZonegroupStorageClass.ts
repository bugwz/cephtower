type Row = Record<string, any>
const token = (value: unknown): value is string => typeof value === 'string' && !!value && !value.startsWith('-') && !/[\x00-\x1f\x7f]/.test(value)
function targets(row?: Row, allowEmpty = false) {
  if (!row || row.stale === true || !token(row.id) || !token(row.name) || typeof row.realm_id !== 'string' || typeof row.default_placement !== 'string') throw new Error('组身份、默认放置目标未知或库存已过期')
  const values = row.placement_targets
  if (!Array.isArray(values) || !allowEmpty && !values.length || values.some(p => !token(p?.name) || !Array.isArray(p.storage_classes) || p.storage_classes.some((c: unknown) => !token(c)) || new Set(p.storage_classes).size !== p.storage_classes.length) || new Set(values.map(p => p.name)).size !== values.length) throw new Error('已有放置目标或存储类清单不完整')
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

export function groupPlacementCreateBlocked(row: Row) { try { targets(row,true); return undefined } catch (error) { return (error as Error).message } }
export function groupPlacementCreateInput(values: Row,row?: Row) {
  const entries = targets(row,true)
  if (!token(values.placement_id) || values.placement_id.includes('/') || entries.some(p => p.name === values.placement_id)) throw new Error('必须输入尚不存在且不含 / 的放置目标名称')
  let tags: unknown
  try { tags = JSON.parse(String(values.tags_json ?? '')) } catch { throw new Error('标签必须是 JSON 字符串数组') }
  if (!Array.isArray(tags) || tags.some(t => !token(t) || t.includes(',')) || new Set(tags).size !== tags.length) throw new Error('标签必须唯一、非空，且不含逗号或控制字符')
  if (values.confirm_create !== 'acknowledged') throw new Error('请确认新建范围与风险')
  return {zonegroup_id:row!.id,name:row!.name,realm_id:row!.realm_id,placement_id:values.placement_id,tags,expected_default_placement:row!.default_placement,confirm_create:true}
}
export function groupPlacementCreateConfirmation(values: Row,row?: Row) {
  const p = groupPlacementCreateInput(values,row)
  return `在 Zonegroup ${p.name}（${p.zonegroup_id}）新建放置目标 ${p.placement_id}，初始存储类 STANDARD，标签 ${JSON.stringify(p.tags)}。${p.expected_default_placement === '' ? '当前没有默认目标，原生操作会把新目标设为默认；请先评估新桶创建行为。' : `保留当前默认目标 ${p.expected_default_placement}。`}不覆盖已有目标，不配置任何 Zone 池，不迁移数据或自动发布 Period/重启网关。请先备份，并在发布前完成相关 Zone 的放置池配置。检查不是跨进程原子锁，分步核验失败可能已写入，不自动重试或回滚。`
}

export function groupPlacementDefaultClasses(row?: Row, placement?: unknown) {
  return (targets(row).find(p => p.name === placement)?.storage_classes || []).map((value: string) => ({value,label:value}))
}
export function groupPlacementDefaultInput(values: Row,row?: Row) {
  if (!groupStorageClassOptions(row).some(p => p.value === values.placement_id) || !groupPlacementDefaultClasses(row,values.placement_id).some((p: Row) => p.value === values.storage_class)) throw new Error('请选择已声明的目标与存储类')
  if (values.confirm_default !== 'acknowledged') throw new Error('请确认默认规则变更范围')
  return {zonegroup_id:row!.id,name:row!.name,realm_id:row!.realm_id,placement_id:values.placement_id,storage_class:values.storage_class,expected_default_placement:row!.default_placement,confirm_default:true}
}
export function groupPlacementDefaultConfirmation(values: Row,row?: Row) {
  const p = groupPlacementDefaultInput(values,row)
  return `将 Zonegroup ${p.name}（${p.zonegroup_id}）默认放置规则由 ${p.expected_default_placement || '未设置'} 改为目标 ${p.placement_id}、存储类 ${p.storage_class}。这会影响采用组默认值的新请求；不覆盖显式放置规则、不迁移已有桶或对象。保留目标、标签、全部存储类及 Zone 池配置。不自动发布 Period 或重启网关；请先核对相关 Zone 池及存储类用途（包括云分层限制），再单独评估发布。请备份；检查不是原子锁，失败可能已写入，不自动重试或回滚。`
}
