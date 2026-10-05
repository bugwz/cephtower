import { groupStorageClassInput, groupStorageClassOptions } from './rgwZonegroupStorageClass'
import { cloudRestoreLocalClasses } from './rgwCloudRestore'
import { cloudConnectionValues } from './rgwCloudConnection'
import { cloudTargetTextFields, cloudTargetValues } from './rgwCloudTarget'
import { cloudACLList } from './rgwCloudACL'
type Row=Record<string,any>
export function cloudCreateTargets(row?:Row) {
  const options=groupStorageClassOptions(row)
  if(!row?.realm_id)throw new Error('需要已有 Realm，不会自动初始化 Realm')
  return options.filter(o=>Array.isArray(row.placement_targets.find((t:Row)=>t.name===o.value).tier_targets)&&cloudRestoreLocalClasses(row,o.value).length)
}
export function cloudCreateBlocked(row:Row) {try {return cloudCreateTargets(row).length?undefined:'没有含本地恢复类和完整分层清单的已有目标'}catch(e){return (e as Error).message}}
export function cloudCreateChanged(changed:Row) {
  const reset:Row=Object.keys(changed).some(k=>!['confirm_create','credentials_saved'].includes(k))?{confirm_create:undefined,credentials_saved:undefined}:{}
  if(Object.prototype.hasOwnProperty.call(changed,'placement_id'))Object.assign(reset,{storage_class:undefined,restore_storage_class:undefined,endpoint:undefined,access_key:undefined,secret:undefined})
  if(Object.prototype.hasOwnProperty.call(changed,'tier_type'))Object.assign(reset,{glacier_restore_days:undefined,glacier_restore_tier_type:undefined})
  return reset
}
export function cloudCreateInput(values:Row,row?:Row) {
  const identity=groupStorageClassInput(values,row)
  if(identity.storage_class==='STANDARD'||!cloudCreateTargets(row).some(o=>o.value===values.placement_id)||!['cloud-s3','cloud-s3-glacier'].includes(values.tier_type))throw new Error('请选择已有目标和支持的云类型，输入未声明的非 STANDARD 类')
  const connection=cloudConnectionValues(values)
  if(values.credentials_saved!=='acknowledged')throw new Error('请确认远端凭据已配置并安全保存')
  const raw:Row={host_style:values.host_style,multipart_sync_threshold:values.multipart_sync_threshold,multipart_min_part_size:values.multipart_min_part_size}
  for(const [key] of cloudTargetTextFields) {
    if(values[`${key}_mode`]==='clear')raw[key]=''
    else if(values[`${key}_mode`]==='set'&&typeof values[key]==='string'&&values[key]!=='')raw[key]=values[key]
    else throw new Error('请明确设置非空目标文本或选择空值')
  }
  const target=cloudTargetValues(raw)
  let acls:ReturnType<typeof cloudACLList>
  try {acls=cloudACLList(JSON.parse(values.acls_json))}catch{throw new Error('请填写有效 ACL JSON 数组，无映射明确填 []')}
  if(!['true','false'].includes(values.retain_head_object)||!['true','false'].includes(values.allow_read_through)||!Number.isSafeInteger(values.read_through_restore_days)||values.read_through_restore_days<0||!cloudRestoreLocalClasses(row,values.placement_id).some((c:Row)=>c.value===values.restore_storage_class))throw new Error('请选择明确布尔值、精确非负天数及已有本地恢复类')
  const glacier:{glacier_restore_days?:number;glacier_restore_tier_type?:string}={}
  if(values.tier_type==='cloud-s3-glacier') {
    if(!Number.isSafeInteger(values.glacier_restore_days)||values.glacier_restore_days<0||!['Standard','Expedited'].includes(values.glacier_restore_tier_type))throw new Error('请填写 Glacier 非负精确天数和恢复等级')
    glacier.glacier_restore_days=values.glacier_restore_days;glacier.glacier_restore_tier_type=values.glacier_restore_tier_type
  } else if(values.glacier_restore_days!==undefined||values.glacier_restore_tier_type!==undefined)throw new Error('普通 S3 不能提交 Glacier 参数')
  return {...identity,...connection,tier_type:values.tier_type as string,target,acls,retain_head_object:values.retain_head_object==='true',allow_read_through:values.allow_read_through==='true',read_through_restore_days:values.read_through_restore_days as number,restore_storage_class:values.restore_storage_class as string,...glacier,credentials_saved:true}
}
export function cloudCreateConfirmation(values:Row,row?:Row) {
  const p=cloudCreateInput(values,row)
  return `在组 ${p.name}（${p.zonegroup_id}）目标 ${p.placement_id} 新建 ${p.tier_type} 存储类 ${p.storage_class}，端点 ${p.endpoint}，目标配置 ${JSON.stringify(p.target)}，ACL 映射 ${p.acls.length} 条，保留头对象=${p.retain_head_object}、读穿透=${p.allow_read_through}、恢复类=${p.restore_storage_class}、读穿透天数=${p.read_through_restore_days}。${p.tier_type==='cloud-s3-glacier'?`Glacier 恢复天数=${p.glacier_restore_days}、等级=${p.glacier_restore_tier_type}。`:''}${p.endpoint.startsWith('http:')?'警告：HTTP 不提供 TLS。':''}凭据必须已在远端配置并安全保存，本机原生命令进程参数可能被有权限的用户/监控读取。需核对恢复类各 Zone 池映射、远端权限、分段限制与恢复费用；此操作不配置 Zone 池、不创建远端资源、不迁移对象或建立生命周期规则，不验证远端连通性。${p.expected_default_placement===''?'原生会将当前目标 STANDARD 初始化为默认规则。':''}随后发布 Realm ${p.realm_id} Period，可能包含其他待提交变更；不隐式创建 Realm 或重启网关。请备份，类型不能在此重设；非原子操作，失败可能部分生效，不自动重试或回滚。`
}
