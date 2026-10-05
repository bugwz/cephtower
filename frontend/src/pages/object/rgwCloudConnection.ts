import { cloudRestoreClasses, cloudRestoreTargets } from './rgwCloudRestore'
type Row=Record<string,any>
export function cloudConnectionEndpoint(value:unknown):string {
  if(typeof value!=='string'||!value||new TextEncoder().encode(value).length>4096||/[\\?#\s\x00-\x1f\x7f-\x9f]/.test(value))throw new Error('端点必须为不含用户信息、查询或片段的 HTTP(S) URL')
  let url:URL
  try {url=new URL(value)}catch{throw new Error('端点 URL 无效')}
  const authority=value.match(/^https?:\/\/([^/]+)/)?.[1]
  if(!authority||authority.includes('@')||authority.endsWith(':')||!['http:','https:'].includes(url.protocol)||!url.hostname||url.username||url.password||url.port==='0')throw new Error('端点必须使用有效 HTTP(S) 主机及端口，不得含用户信息')
  return value
}
export function cloudConnectionChanged(changed:Row) {
  if(['placement_id','storage_class'].some(k=>Object.prototype.hasOwnProperty.call(changed,k)))return {endpoint:undefined,access_key:undefined,secret:undefined,credentials_saved:undefined,confirm_connection:undefined,...(Object.prototype.hasOwnProperty.call(changed,'placement_id')?{storage_class:undefined}:{})}
  return Object.keys(changed).some(k=>!['confirm_connection','credentials_saved'].includes(k))?{credentials_saved:undefined,confirm_connection:undefined}:{}
}
export function cloudConnectionInput(values:Row,row?:Row) {
  if(!cloudRestoreTargets(row).some(t=>t.value===values.placement_id)||!cloudRestoreClasses(row,values.placement_id).some((c:Row)=>c.value===values.storage_class))throw new Error('请选择已有 Realm 中的云分层')
  const tier=row!.placement_targets.find((t:Row)=>t.name===values.placement_id).tier_targets.find((t:Row)=>t.key===values.storage_class).val
  const endpoint=cloudConnectionEndpoint(values.endpoint)
  for(const key of ['access_key','secret'])if(typeof values[key]!=='string'||!values[key].trim()||new TextEncoder().encode(values[key]).length>4096||values[key].includes('[REDACTED]')||/[\x00-\x1f\x7f-\x9f]/.test(values[key]))throw new Error('请重新输入远端已配置的非空凭据，不可使用脱敏值')
  if(values.credentials_saved!=='acknowledged'||values.confirm_connection!=='acknowledged')throw new Error('请确认已安全保存凭据、进程参数可见性及发布风险')
  return {name:row!.name,zonegroup_id:row!.id,realm_id:row!.realm_id,placement_id:values.placement_id,storage_class:values.storage_class,tier_type:tier.tier_type,expected_default_placement:row!.default_placement,endpoint,access_key:values.access_key,secret:values.secret,credentials_saved:true,confirm_connection:true}
}
export function cloudConnectionConfirmation(values:Row,row?:Row) {
  const p=cloudConnectionInput(values,row)
  return `替换组 ${p.name}（${p.zonegroup_id}）目标 ${p.placement_id} 云类 ${p.storage_class} 的端点为 ${p.endpoint}，并替换两项凭据。${p.endpoint.startsWith('http:')?'警告：HTTP 不提供 TLS 加密。':''}不回填或展示旧凭据；新凭据必须已在远端配置并安全保存。本机原生 --tier-config 会将凭据传入进程参数，具备相应权限的本机用户/监控可能看到，日志脱敏不能消除此风险。改变连接可能中断已有数据访问；不迁移对象、不测试远端连通性或授权。保留区域、路径、ACL 和恢复参数。${p.expected_default_placement===''?'原生将初始化当前目标 STANDARD 为默认规则。':''}随后发布 Realm ${p.realm_id} 的 Period，可能包含其他待提交变更。请备份；失败可能部分生效，不自动重试或回滚。`
}
