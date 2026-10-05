import { jsonInit, request, type ApiRecord } from '../../api/client'
import { groupPlacementClassRows, zonePlacementClassRows } from './rgwPlacementClassData'

const object=(v:unknown):v is ApiRecord=>!!v&&typeof v==='object'&&!Array.isArray(v)
const identity=(v:unknown):v is string=>typeof v==='string'&&v.length>0
export type LocalClassMapping={key:string;placement:string;storageClass:string;zone:string;zoneId:string;status:string;pool:string;compression:string;observed:string}

export async function readLocalClassZones(clusterId:number,signal?:AbortSignal) {
  if(!Number.isSafeInteger(clusterId)||clusterId<=0)throw new Error('Invalid cluster identity')
  const rows:ApiRecord[]=[],seen=new Set<string>()
  let cursor='',stale=false
  for(let page=0;page<100;page++) {
    if(signal?.aborted)throw new Error('Zone inventory read cancelled')
    const query=new URLSearchParams({limit:'500'});if(cursor)query.set('cursor',cursor)
    const result=await request<ApiRecord>(`/rgw/zones?${query}`,jsonInit('GET',{cluster_id:clusterId},{signal,suppressErrorNotification:true,cache:'no-store'}))
    if(signal?.aborted)throw new Error('Zone inventory read cancelled')
    if(!object(result)||!Array.isArray(result.items)||!object(result.meta)||typeof result.meta.stale!=='boolean'||!object(result.pagination))throw new Error('Incomplete zone inventory')
    stale ||= result.meta.stale
    for(const item of result.items) {
      if(!object(item)||!object(item.data)||typeof item.stale!=='boolean')throw new Error('Invalid zone inventory')
      rows.push({...item.data,stale:item.stale,observed_at:item.observed_at});stale ||= item.stale
    }
    const next=result.pagination.next_cursor
    if(next===undefined||next===null||next==='')return {rows,stale}
    if(typeof next!=='string'||seen.has(next))throw new Error('Invalid zone cursor')
    seen.add(next);cursor=next
  }
  throw new Error('Zone page limit exceeded')
}

export function localClassMappings(group:ApiRecord,zones:ApiRecord[]) {
  if(!identity(group.id)||typeof group.realm_id!=='string'||!Array.isArray(group.zones))throw new Error('Invalid group identity or members')
  const members=new Map<string,{name:string}>(),inventory=new Map<string,ApiRecord>()
  for(const member of group.zones) {
    if(!object(member)||!identity(member.id)||!identity(member.name)||members.has(member.id))throw new Error('Invalid member identity')
    members.set(member.id,{name:member.name})
  }
  for(const zone of zones) {
    if(!object(zone)||!identity(zone.id)||!identity(zone.name)||inventory.has(zone.id))throw new Error('Invalid zone identity')
    inventory.set(zone.id,zone)
  }
  const classes=groupPlacementClassRows(group.placement_targets),rows:LocalClassMapping[]=[],issues=[...classes.issues]
  if(!members.size)issues.push('组未列出成员 Zone，无法关联本地池映射')
  for(const cls of classes.rows.filter(c=>c.type==='本地'&&c.declared==='已声明'))for(const [id,member] of members) {
    const zone=inventory.get(id)
    const row:LocalClassMapping={key:JSON.stringify([cls.placement,cls.storageClass,id]),placement:cls.placement,storageClass:cls.storageClass,zone:member.name,zoneId:id,status:'未找到成员库存，不能判定未配置',pool:'未知',compression:'未知',observed:typeof zone?.observed_at==='string'?zone.observed_at:'未返回'}
    if(zone) {
      if(zone.name!==member.name||zone.realm_id!==group.realm_id)row.status='成员名称或 Realm 不一致，未关联'
      else {
        const mappings=zonePlacementClassRows(zone.placement_pools)
        const match=mappings.rows.find(m=>m.placement===cls.placement&&m.storageClass===cls.storageClass)
        if(mappings.issues.length)row.status='Zone 映射不完整或存在歧义，未关联'
        else if(!match)row.status='本次 Zone 配置未列出此目标/类映射'
        else {
          row.status='已匹配配置（不证明池存在或可访问）'
          row.pool=match.fields.find(f=>f.name==='数据池（原生引用）')?.value??'未知'
          row.compression=match.fields.find(f=>f.name==='压缩配置')?.value??'未知'
        }
      }
      if(zone.stale===true)row.status+='；Zone 库存过期'
    }
    rows.push(row)
  }
  return {rows,issues}
}
