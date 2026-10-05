type Row = Record<string, any>
const token = (value: unknown): value is string => typeof value === 'string' && !!value && !value.startsWith('-') && !/[\x00-\x1f\x7f]/.test(value)
export const zonePlacementCompressions = ['none', 'lz4', 'zlib', 'snappy', 'zstd', 'brotli']
function placements(row?: Row) {
  if (!row || row.stale === true || !token(row.id) || !token(row.name) || typeof row.realm_id !== 'string') throw new Error('Zone 身份不完整或库存过期')
  const entries = row.placement_pools
  if (!Array.isArray(entries) || !entries.length || entries.some(p => !token(p?.key) || !p.val || typeof p.val.storage_classes !== 'object' || Array.isArray(p.val.storage_classes) || !p.val.storage_classes) || new Set(entries.map(p => p.key)).size !== entries.length) throw new Error('放置配置不完整或有重复')
  return entries as Array<{key: string; val: Row}>
}
export function zonePlacementBlocked(row: Row) { try { placements(row); if (!zonePlacementGroups(row).length) return '缺少同 Realm 的 Zonegroup 成员信息'; return undefined } catch (error) { return (error as Error).message } }
export function zonePlacementGroups(row?: Row) {
  return Array.isArray(row?.zonegroup_memberships) ? row.zonegroup_memberships.filter((g: Row) => token(g.zonegroup_id) && g.realm_id === row.realm_id).map((g: Row) => ({value:g.zonegroup_id,label:String(g.zonegroup_name || g.zonegroup_id)})) : []
}
export function zonePlacementOptions(row?: Row) { return placements(row).map(p => ({value:p.key,label:p.key})) }
export function zonePlacementClasses(row: Row | undefined, placement: unknown) {
  return Object.keys(placements(row).find(p => p.key === placement)?.val.storage_classes || {}).map(key => ({value:key,label:key}))
}
export function zonePlacementChanged(changed: Row, values: Row, row?: Row) {
  if (!('placement_id' in changed) && !('storage_class' in changed)) return {}
  const placement = placements(row).find(p => p.key === values.placement_id)
  const storage = 'placement_id' in changed ? undefined : values.storage_class
  const sc = placement?.val.storage_classes[storage]
  return { storage_class:storage, index_pool:placement?.val.index_pool, data_extra_pool:placement?.val.data_extra_pool, data_pool:sc?.data_pool, compression:sc?.compression_type }
}
export function zonePlacementInput(values: Row, row?: Row) {
  placements(row)
  if (!zonePlacementGroups(row).some((g: Row) => g.value === values.zonegroup_id)) throw new Error('必须选择所属 Zonegroup')
  if (!zonePlacementClasses(row,values.placement_id).some(c => c.value === values.storage_class)) throw new Error('必须选择已有放置目标和存储类')
  for (const key of ['index_pool','data_pool','data_extra_pool']) if (typeof values[key] !== 'string' || key !== 'data_extra_pool' && !values[key] || /[\x00-\x1f\x7f]/.test(values[key])) throw new Error('池引用不能为空或包含控制字符；额外数据池可显式留空')
  if (!zonePlacementCompressions.includes(values.compression) || values.confirm_placement !== 'acknowledged') throw new Error('请选择压缩算法并确认影响')
  return { zone_id:row!.id, name:row!.name, realm_id:row!.realm_id, zonegroup_id:values.zonegroup_id, placement_id:values.placement_id, storage_class:values.storage_class, index_pool:values.index_pool, data_pool:values.data_pool, data_extra_pool:values.data_extra_pool, compression:values.compression, confirm_placement:true }
}
export function zonePlacementConfirmation(values: Row, row?: Row) {
  const p = zonePlacementInput(values,row)
  return `修改 Zone ${p.name}（${p.zone_id}）放置目标 ${p.placement_id} 的 ${p.storage_class} 存储类：索引池 ${p.index_pool}、数据池 ${p.data_pool}、额外数据池 ${p.data_extra_pool || '空（原生回退）'}、压缩 ${p.compression}。索引池和额外数据池影响整个放置目标，而非仅此存储类。此操作不搬迁现有数据；更换池可能使现有桶或对象不可访问，必须先备份并评估业务影响。配置写入成功不证明池存在、可用或支持 OMAP，请事先核对。压缩配置不重写已有对象，算法可用性取决于网关部署。${p.realm_id ? `将提交 Realm ${p.realm_id} 的 Period，可能发布其他待提交配置；不代表远端同步完成。` : '无 Realm，不发布 Period。'}不自动重启网关，不自动重试或回滚；分步执行可能部分生效，写前核验不是跨进程原子锁。`
}
