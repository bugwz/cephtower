import { jsonInit, request, type ApiRecord } from '../../api/client'
import { groupPlacementClassRows, type PlacementClassRow } from './rgwPlacementClassData'

const object=(v:unknown):v is ApiRecord=>!!v&&typeof v==='object'&&!Array.isArray(v)
export type StorageClassInventoryRow=PlacementClassRow&{groupId:string;groupName:string;realm:string;stale:boolean;observed:string}
export type StorageClassInventory={rows:StorageClassInventoryRow[];issues:string[];stale:boolean;groups:number}

export function storageClassInventory(groups:ApiRecord[]):StorageClassInventory {
  const rows:StorageClassInventoryRow[]=[],issues:string[]=[],seen=new Set<string>()
  let stale=false
  for(const group of groups) {
    if(!object(group)||typeof group.id!=='string'||!group.id||typeof group.name!=='string'||!group.name||seen.has(group.id)||typeof group.stale!=='boolean')throw new Error('Invalid or duplicate group identity')
    seen.add(group.id);stale ||= group.stale
    const projected=groupPlacementClassRows(group.placement_targets)
    issues.push(...projected.issues.map(issue=>`${group.name} (${group.id})：${issue}`))
    for(const row of projected.rows)rows.push({...row,key:JSON.stringify([group.id,row.placement,row.storageClass]),groupId:group.id,groupName:group.name,realm:typeof group.realm_id==='string'?group.realm_id||'未关联 Realm':'Realm 未返回',stale:group.stale,observed:typeof group.observed_at==='string'?group.observed_at:'未返回'})
  }
  return {rows,issues,stale,groups:groups.length}
}

export async function readStorageClassInventory(clusterId:number,signal?:AbortSignal):Promise<StorageClassInventory> {
  if(!Number.isSafeInteger(clusterId)||clusterId<=0)throw new Error('Invalid cluster identity')
  const groups:ApiRecord[]=[],cursors=new Set<string>()
  let cursor='',stale=false
  for(let page=0;page<100;page++) {
    if(signal?.aborted)throw new Error('Inventory read cancelled')
    const query=new URLSearchParams({limit:'500'});if(cursor)query.set('cursor',cursor)
    const result=await request<ApiRecord>(`/rgw/zonegroups?${query}`,jsonInit('GET',{cluster_id:clusterId},{signal,suppressErrorNotification:true,cache:'no-store'}))
    if(signal?.aborted)throw new Error('Inventory read cancelled')
    if(!object(result)||!Array.isArray(result.items)||!object(result.meta)||typeof result.meta.stale!=='boolean'||!object(result.pagination))throw new Error('Incomplete group inventory')
    stale ||= result.meta.stale
    for(const item of result.items) {
      if(!object(item)||!object(item.data)||typeof item.stale!=='boolean')throw new Error('Invalid group inventory')
      groups.push({...item.data,stale:item.stale,observed_at:item.observed_at})
    }
    const next=result.pagination.next_cursor
    if(next===undefined||next===null||next==='') {
      const data=storageClassInventory(groups)
      return {...data,stale:stale||data.stale}
    }
    if(typeof next!=='string'||cursors.has(next))throw new Error('Invalid group pagination')
    cursors.add(next);cursor=next
  }
  throw new Error('Group page limit exceeded')
}

export function filterStorageClasses(rows:StorageClassInventoryRow[],query:string,groupId?:string,type?:string) {
  const term=query.trim().toLocaleLowerCase()
  return rows.filter(row=>(!groupId||row.groupId===groupId)&&(!type||row.type===type)&&(!term||[row.groupName,row.groupId,row.placement,row.storageClass,row.realm,...row.fields.map(f=>f.value)].some(value=>value.toLocaleLowerCase().includes(term))))
}
