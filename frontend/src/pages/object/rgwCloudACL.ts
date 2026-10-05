import { cloudRestoreClasses, cloudRestoreTargets } from './rgwCloudRestore'
type Row=Record<string,any>
type ACL={source_id:string;dest_id:string;type:string}
const object=(v:unknown):v is Row=>!!v&&typeof v==='object'&&!Array.isArray(v)
export function cloudACLList(value:unknown):ACL[] {
  if(!Array.isArray(value)||value.length>256)throw new Error('ACL 必须是最多 256 条记录的 JSON 数组')
  const seen=new Set<string>()
  const entries=value.map(v=>{
    if(!object(v)||Object.keys(v).length!==3||typeof v.source_id!=='string'||!v.source_id||typeof v.dest_id!=='string'||!['id','email','uri'].includes(v.type)||seen.has(v.source_id))throw new Error('ACL 需要唯一非空 source_id、字符串 dest_id 和 id/email/uri 类型，不能含额外字段')
    for(const text of [v.source_id,v.dest_id])if(new TextEncoder().encode(text).length>4096||/[,{}\x00-\x1f\x7f-\x9f]/.test(text))throw new Error('身份值不能包含逗号、花括号、控制字符或超过 4096 字节')
    seen.add(v.source_id)
    return {source_id:v.source_id,dest_id:v.dest_id,type:v.type}
  })
  return entries.sort((a,b)=>a.source_id<b.source_id?-1:a.source_id>b.source_id?1:0)
}
function selected(row:Row|undefined,values:Row) {
  if(!cloudRestoreTargets(row).some(t=>t.value===values.placement_id)||!cloudRestoreClasses(row,values.placement_id).some((c:Row)=>c.value===values.storage_class))throw new Error('请选择已有 Realm 中的云分层')
  const tier=row!.placement_targets.find((t:Row)=>t.name===values.placement_id).tier_targets.find((t:Row)=>t.key===values.storage_class).val
  if(!Array.isArray(tier.s3?.acl_mappings))throw new Error('ACL 库存未知，不能按空清单编辑')
  const old=cloudACLList(tier.s3.acl_mappings.map((a:unknown)=>{
    if(!object(a)||Object.keys(a).length!==2||!object(a.val)||a.key!==a.val.source_id)throw new Error('原生 ACL 身份不一致，不能猜测修复')
    return a.val
  }))
  return {tier,old}
}
export function cloudACLChanged(changed:Row,values:Row,row?:Row) {
  const reset={acls_json:undefined,confirm_acl:undefined,confirm_clear:undefined}
  if(Object.prototype.hasOwnProperty.call(changed,'placement_id'))return {...reset,storage_class:undefined}
  if(Object.prototype.hasOwnProperty.call(changed,'storage_class')) {
    try {return {...reset,acls_json:JSON.stringify(selected(row,values).old,null,2),confirm_clear:'no'}}catch{return reset}
  }
  return Object.keys(changed).some(k=>k!=='confirm_acl')?{confirm_acl:undefined,...(Object.prototype.hasOwnProperty.call(changed,'acls_json')?{confirm_clear:undefined}:{})}:{}
}
export function cloudACLInput(values:Row,row?:Row) {
  const {tier,old}=selected(row,values)
  let parsed:unknown;try {parsed=JSON.parse(values.acls_json)}catch{throw new Error('请输入 ACL JSON 数组')}
  const acls=cloudACLList(parsed)
  if(JSON.stringify(old)===JSON.stringify(acls))throw new Error('ACL 没有变化')
  if(values.confirm_acl!=='acknowledged'||!['no','acknowledged'].includes(values.confirm_clear)||acls.length===0&&values.confirm_clear!=='acknowledged')throw new Error('请明确确认变更、发布及清空范围')
  return {name:row!.name,zonegroup_id:row!.id,realm_id:row!.realm_id,placement_id:values.placement_id,storage_class:values.storage_class,tier_type:tier.tier_type,expected_default_placement:row!.default_placement,expected_acls:old,acls,confirm_acl:true,confirm_clear:values.confirm_clear==='acknowledged'}
}
export function cloudACLConfirmation(values:Row,row?:Row) {
  const p=cloudACLInput(values,row),sources=new Set(p.acls.map(a=>a.source_id)),removed=p.expected_acls.filter(a=>!sources.has(a.source_id))
  return `修改组 ${p.name}（${p.zonegroup_id}）目标 ${p.placement_id} 云类 ${p.storage_class} 的 ACL 身份映射：${p.expected_acls.length} 条变为 ${p.acls.length} 条，移除 ${removed.length} 个旧来源。${p.acls.length===0?'将清空全部显式映射；这不代表远端对象不可访问。':''}空 dest_id 是原生空目标身份，不是删除。保留端点、凭据和恢复参数，不直接重写已有远端对象 ACL；后续云分层的访问权限可能改变。${p.expected_default_placement===''?'原生将初始化当前目标 STANDARD 为默认放置规则。':''}随后发布 Realm ${p.realm_id} 的 Period，可能包含其他待提交变更。请备份并核对来源/远端身份；非原子操作，失败可能部分生效，不自动重试或回滚。`
}
