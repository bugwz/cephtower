import { groupStorageClassOptions } from './rgwZonegroupStorageClass'
type Row=Record<string,any>
const object=(v:unknown):v is Row=>!!v&&typeof v==='object'&&!Array.isArray(v)
function targets(row?:Row) {
  groupStorageClassOptions(row)
  if(typeof row?.realm_id!=='string'||!row.realm_id)throw new Error('需要已配置的 Realm；此操作不会自动创建 Realm')
  const targets=row.placement_targets as Row[]
  for(const t of targets) {
    const tiers=t.tier_targets??[]
    if(t.tier_targets===null||!Array.isArray(tiers)||tiers.some(v=>!object(v)||typeof v.key!=='string'||!v.key||!object(v.val))||new Set(tiers.map(v=>v.key)).size!==tiers.length)throw new Error('分层清单缺失或存在歧义')
  }
  return targets
}
export function cloudRestoreClasses(row?:Row,placement?:unknown) {
  const target=targets(row).find(t=>t.name===placement)
  return (target?.tier_targets??[]).filter((t:Row)=>t.key!=='STANDARD'&&target!.storage_classes.includes(t.key)&&t.val.storage_class===t.key&&['cloud-s3','cloud-s3-glacier'].includes(t.val.tier_type)).map((t:Row)=>({value:t.key,label:`${t.key} (${t.val.tier_type})`}))
}
export function cloudRestoreTargets(row?:Row) {return targets(row).filter(t=>!t.name.includes('/')&&cloudRestoreClasses(row,t.name).length).map(t=>({value:t.name,label:t.name}))}
export function cloudRestoreBlocked(row:Row) {try {if(!cloudRestoreTargets(row).length)return '没有可编辑的已声明云分层';return undefined}catch(e){return (e as Error).message}}
export function cloudRestoreLocalClasses(row?:Row,placement?:unknown) {
  const target=targets(row).find(t=>t.name===placement)
  return (target?.storage_classes??[]).filter((c:string)=>!/[ ,=/\\]/.test(c)&&!(target?.tier_targets??[]).some((t:Row)=>t.key===c)).map((c:string)=>({value:c,label:c}))
}
export function cloudRestoreChanged(changed:Row,values:Row,row?:Row) {
  const reset={retain_head_object:undefined,allow_read_through:undefined,read_through_restore_days:undefined,restore_storage_class:undefined,confirm_restore:undefined}
  if(Object.prototype.hasOwnProperty.call(changed,'placement_id'))return {...reset,storage_class:undefined}
  if(Object.prototype.hasOwnProperty.call(changed,'storage_class')) {
    const tier=targets(row).find(t=>t.name===values.placement_id)?.tier_targets?.find((t:Row)=>t.key===values.storage_class)?.val
    if(!tier)return reset
    return {...reset,retain_head_object:typeof tier.retain_head_object==='boolean'?String(tier.retain_head_object):undefined,allow_read_through:typeof tier.allow_read_through==='boolean'?String(tier.allow_read_through):undefined,read_through_restore_days:Number.isSafeInteger(tier.read_through_restore_days)&&tier.read_through_restore_days>=0?tier.read_through_restore_days:undefined,restore_storage_class:cloudRestoreLocalClasses(row,values.placement_id).some((c:Row)=>c.value===tier.restore_storage_class)?tier.restore_storage_class:undefined}
  }
  return Object.keys(changed).some(k=>k!=='confirm_restore')?{confirm_restore:undefined}:{}
}
export function cloudRestoreInput(values:Row,row?:Row) {
  if(!cloudRestoreTargets(row).some(t=>t.value===values.placement_id)||!cloudRestoreClasses(row,values.placement_id).some((c:Row)=>c.value===values.storage_class))throw new Error('请选择已有云分层')
  if(!['true','false'].includes(values.retain_head_object)||!['true','false'].includes(values.allow_read_through)||!Number.isSafeInteger(values.read_through_restore_days)||values.read_through_restore_days<0||!cloudRestoreLocalClasses(row,values.placement_id).some((c:Row)=>c.value===values.restore_storage_class))throw new Error('必须明确选择布尔值、非负精确天数及目标中的本地恢复类')
  if(values.confirm_restore!=='acknowledged')throw new Error('请确认恢复配置和 Period 发布风险')
  const tier=targets(row).find(t=>t.name===values.placement_id)!.tier_targets.find((t:Row)=>t.key===values.storage_class).val
  return {name:row!.name,zonegroup_id:row!.id,realm_id:row!.realm_id,expected_default_placement:row!.default_placement,placement_id:values.placement_id,storage_class:values.storage_class,tier_type:tier.tier_type,retain_head_object:values.retain_head_object==='true',allow_read_through:values.allow_read_through==='true',read_through_restore_days:values.read_through_restore_days,restore_storage_class:values.restore_storage_class,confirm_restore:true}
}
export function cloudRestoreConfirmation(values:Row,row?:Row) {
  const p=cloudRestoreInput(values,row)
  return `修改组 ${p.name}（${p.zonegroup_id}）目标 ${p.placement_id} 的云分层 ${p.storage_class}：保留头对象=${p.retain_head_object}，读穿透=${p.allow_read_through}，恢复天数=${p.read_through_restore_days}，恢复类=${p.restore_storage_class}。改变头对象保留可能影响后续恢复，读穿透可能触发远端请求和费用；请核对各 Zone 的恢复类池映射。本次不直接恢复任何对象，不修改端点、凭据、ACL 或 Glacier 专有参数。${p.expected_default_placement===''?'原生操作会初始化默认目标为当前目标 STANDARD。':''}随后提交 Realm ${p.realm_id} 的 Period，可能一并发布其他待提交变更。请先备份；非原子事务，失败可能部分生效，不自动重试或回滚，不隐式新建 Realm 或重启服务。`
}
