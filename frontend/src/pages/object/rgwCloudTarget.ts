import { cloudRestoreClasses, cloudRestoreTargets } from './rgwCloudRestore'
type Row=Record<string,any>
export const cloudTargetTextFields=[['region','目标区域'],['target_path','目标路径'],['target_storage_class','远端目标存储类']] as const
const keys=['region','host_style','target_path','target_storage_class','multipart_sync_threshold','multipart_min_part_size'] as const
export function cloudTargetValues(value:unknown):Row {
  if(!value||typeof value!=='object'||Array.isArray(value)||Object.keys(value).length!==6)throw new Error('必须提供完整且仅包含六项的目标参数')
  const v=value as Row,result:Row={}
  for(const key of keys) {
    if(key.startsWith('multipart_')) {if(!Number.isSafeInteger(v[key])||v[key]<0)throw new Error('分段大小必须为非负精确整数（字节）')}
    else if(typeof v[key]!=='string'||new TextEncoder().encode(v[key]).length>4096||/[\x00-\x1f\x7f-\x9f]/.test(v[key]))throw new Error('目标文本字段缺失、过长或包含控制字符')
    result[key]=v[key]
  }
  if(!['path','virtual'].includes(result.host_style))throw new Error('请选择原生 path 或 virtual 寻址方式')
  return result
}
function selected(row:Row|undefined,values:Row) {
  if(!cloudRestoreTargets(row).some(t=>t.value===values.placement_id)||!cloudRestoreClasses(row,values.placement_id).some((c:Row)=>c.value===values.storage_class))throw new Error('请选择已有 Realm 中的云分层')
  const tier=row!.placement_targets.find((t:Row)=>t.name===values.placement_id).tier_targets.find((t:Row)=>t.key===values.storage_class).val
  return {tier,old:cloudTargetValues(Object.fromEntries(keys.map(key=>[key,tier.s3?.[key]])))}
}
export function cloudTargetChanged(changed:Row,values:Row,row?:Row) {
  const reset:Row={confirm_target:undefined}
  for(const key of keys)reset[key]=undefined
  for(const [key] of cloudTargetTextFields)reset[`${key}_mode`]=undefined
  if(Object.prototype.hasOwnProperty.call(changed,'placement_id'))return {...reset,storage_class:undefined}
  if(Object.prototype.hasOwnProperty.call(changed,'storage_class')) {
    try {
      const {old}=selected(row,values)
      return {...reset,...old,...Object.fromEntries(cloudTargetTextFields.map(([key])=>[`${key}_mode`,old[key]===''?'clear':'set']))}
    }catch{return reset}
  }
  return Object.keys(changed).some(k=>k!=='confirm_target')?{confirm_target:undefined}:{}
}
export function cloudTargetInput(values:Row,row?:Row) {
  const {tier,old}=selected(row,values),raw:Row=Object.fromEntries(keys.map(key=>[key,values[key]]))
  for(const [key] of cloudTargetTextFields) {
    if(values[`${key}_mode`]==='clear')raw[key]=''
    else if(values[`${key}_mode`]!=='set'||typeof values[key]!=='string'||values[key]==='')throw new Error('文本字段必须明确设置非空值或清空')
  }
  const target=cloudTargetValues(raw)
  if(JSON.stringify(old)===JSON.stringify(target))throw new Error('目标配置没有变化')
  if(values.confirm_target!=='acknowledged')throw new Error('请确认目标参数、远端访问及 Period 发布风险')
  return {name:row!.name,zonegroup_id:row!.id,realm_id:row!.realm_id,placement_id:values.placement_id,storage_class:values.storage_class,tier_type:tier.tier_type,expected_default_placement:row!.default_placement,expected_target:old,target,confirm_target:true}
}
export function cloudTargetConfirmation(values:Row,row?:Row) {
  const p=cloudTargetInput(values,row)
  return `修改组 ${p.name}（${p.zonegroup_id}）目标 ${p.placement_id} 云类 ${p.storage_class} 的区域、寻址方式、目标路径/存储类及分段参数。目标值：${JSON.stringify(p.target)}。空文本表示显式清空，0 是原生数值而不是省略。改变目标路径/类可能影响后续写入和已有数据访问；必须核对远端支持、分段限制和费用，本操作不迁移对象或验证远端连通性。保留端点、凭据、ACL 和恢复参数。${p.expected_default_placement===''?'原生将初始化当前目标 STANDARD 为默认规则。':''}随后提交 Realm ${p.realm_id} 的 Period，可能一并发布其他待提交变更。请备份；失败可能部分生效，不自动重试或回滚。`
}
