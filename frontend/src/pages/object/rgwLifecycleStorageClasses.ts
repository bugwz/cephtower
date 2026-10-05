import { jsonInit, request, type ApiRecord } from '../../api/client'
import { readStorageClassInventory, type StorageClassInventory } from './rgwStorageClassInventory'

const record=(v:unknown):v is ApiRecord=>!!v&&typeof v==='object'&&!Array.isArray(v)
export type LifecycleStorageClasses={groupId:string;placement:string;observed:string;options:Array<{value:string;label:string}>}

export function lifecycleStorageClasses(bucket:unknown,bucketId:string,inventory:StorageClassInventory):LifecycleStorageClasses {
  if(!record(bucket)||bucket.kind!=='rgw_bucket'||bucket.natural_key!==bucketId||bucket.stale!==false||!record(bucket.data)||inventory.stale)throw new Error('Bucket identity or fresh inventory unavailable')
  const groupId=bucket.data.zonegroup,rule=bucket.data.placement_rule
  if(typeof groupId!=='string'||!groupId||typeof rule!=='string'||!rule)throw new Error('Bucket placement identity unavailable')
  // Native rgw_placement_rule::from_str splits at the first slash.
  const placement=rule.split('/',1)[0]
  if(!placement)throw new Error('Bucket placement name unavailable')
  const scoped=inventory.rows.filter(row=>row.groupId===groupId&&row.placement===placement)
  if(!scoped.length)throw new Error('Bucket placement not found in inventory')
  const options=scoped.filter(row=>row.stale===false&&row.declared==='已声明'&&['本地','云 S3','云 S3 Glacier'].includes(row.type)&&row.storageClass!=='STANDARD'&&(row.type==='本地'||row.fields.find(f=>f.name==='分层内部类名')?.value===row.storageClass)).map(row=>({value:row.storageClass,label:`${row.storageClass} — ${row.type} / ${row.groupName} (${row.groupId}) / ${row.placement}`}))
  if(new Set(options.map(o=>o.value)).size!==options.length)throw new Error('Ambiguous scoped storage class')
  return {groupId,placement,observed:typeof bucket.observed_at==='string'?bucket.observed_at:'未返回',options}
}

export async function readLifecycleStorageClasses(clusterId:number,bucketId:string,signal?:AbortSignal):Promise<LifecycleStorageClasses> {
  if(!Number.isSafeInteger(clusterId)||clusterId<=0||!bucketId||!/^[A-Za-z0-9_-]+$/.test(bucketId))throw new Error('Invalid scope')
  if(signal?.aborted)throw new Error('Read cancelled')
  const bucket=await request<unknown>('/rgw/bucket',jsonInit('GET',{cluster_id:clusterId,bucket_id:bucketId},{signal,cache:'no-store',suppressErrorNotification:true}))
  if(signal?.aborted)throw new Error('Read cancelled')
  const inventory=await readStorageClassInventory(clusterId,signal)
  if(signal?.aborted)throw new Error('Read cancelled')
  return lifecycleStorageClasses(bucket,bucketId,inventory)
}
