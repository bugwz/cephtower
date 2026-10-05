type Row = Record<string, any>
export type PlacementClassRow = {key:string; placement:string; storageClass:string; type:string; declared:string; fields:Array<{name:string;value:string}>}
const object=(v:unknown):v is Row => !!v&&typeof v==='object'&&!Array.isArray(v)
const name=(v:unknown):v is string => typeof v==='string'&&v.length>0
export const placementText=(v:unknown) => typeof v==='string' ? v===''?'原生空值':v : typeof v==='boolean' ? v?'是':'否' : typeof v==='number'&&Number.isSafeInteger(v)?String(v):'未返回、格式异常或超出精确范围'
function endpoint(v:unknown) {
  if (v==='') return '原生空值'
  if (typeof v!=='string') return '未返回或格式异常'
  try { const u=new URL(v);if(!['http:','https:'].includes(u.protocol))return '地址已隐藏（协议异常）';return `${u.origin}${u.pathname}${u.username||u.password||u.search||u.hash?'（凭据、查询及片段已隐藏）':''}` } catch {return '地址已隐藏（无法解析）'}
}
function duplicates(values:string[]) {return new Set(values.filter((v,i)=>values.indexOf(v)!==i))}
function addFields(fields:PlacementClassRow['fields'],source:Row,labels:Record<string,string>) {
  for(const [key,label] of Object.entries(labels))fields.push({name:label,value:placementText(source[key])})
}

export function groupPlacementClassRows(value:unknown) {
  const rows:PlacementClassRow[]=[],issues:string[]=[]
  if(!Array.isArray(value))return {rows,issues:['放置目标未采集或格式异常']}
  const duplicateTargets=duplicates(value.filter(object).map(v=>v.name).filter(name))
  for(const target of value) {
    if(!object(target)||!name(target.name)||duplicateTargets.has(target.name)) {issues.push('目标身份缺失或重复，未展示歧义记录');continue}
    const validClasses=Array.isArray(target.storage_classes)&&target.storage_classes.every(name)&&new Set(target.storage_classes).size===target.storage_classes.length
    if(!validClasses)issues.push(`${target.name}：类声明清单异常`)
    const classes:string[]=validClasses?target.storage_classes:[]
    const tiers=target.tier_targets===undefined?[]:target.tier_targets
    const validTiers=Array.isArray(tiers)&&tiers.every(t=>object(t)&&name(t.key)&&object(t.val))&&new Set(tiers.map(t=>t.key)).size===tiers.length
    if(!validTiers)issues.push(`${target.name}：分层清单异常，不能认定为本地类`)
    const tierMap=new Map<string,Row>(validTiers?tiers.map((t:Row)=>[t.key,t.val]):[])
    const all=[...new Set([...classes,...tierMap.keys()])]
    for(const storageClass of all) {
      const tier=tierMap.get(storageClass)
      const fields:PlacementClassRow['fields']=[]
      const declared=validClasses?classes.includes(storageClass)?'已声明':'仅分层配置存在（声明缺失）':'声明不可用'
      if(declared!=='已声明')issues.push(`${target.name}/${storageClass}：${declared}`)
      let type=validTiers?'本地':'未知'
      if(tier) {
        type=tier.tier_type==='cloud-s3'?'云 S3':tier.tier_type==='cloud-s3-glacier'?'云 S3 Glacier':`未知类型：${placementText(tier.tier_type)}`
        if(tier.storage_class!==storageClass)issues.push(`${target.name}/${storageClass}：分层键与声明名称不一致或缺失`)
        addFields(fields,tier,{tier_type:'原生分层类型',storage_class:'分层内部类名',retain_head_object:'保留头对象',allow_read_through:'允许读穿透',read_through_restore_days:'读穿透恢复天数',restore_storage_class:'恢复目标类'})
        const s3=object(tier.s3)?tier.s3:{}
        fields.push({name:'目标端点',value:endpoint(s3.endpoint)})
        addFields(fields,s3,{region:'目标区域',host_style:'寻址方式',target_storage_class:'目标存储类',target_path:'目标路径',multipart_sync_threshold:'分段同步阈值（字节）',multipart_min_part_size:'最小分段大小（字节）'})
        if(tier.tier_type==='cloud-s3-glacier')addFields(fields,object(tier['s3-glacier'])?tier['s3-glacier']:{},{glacier_restore_days:'Glacier 恢复天数',glacier_restore_tier_type:'Glacier 恢复类型'})
        if(Array.isArray(s3.acl_mappings)&&s3.acl_mappings.every((a:unknown)=>object(a)&&typeof a.key==='string'&&object(a.val))&&new Set(s3.acl_mappings.map((a:Row)=>a.key)).size===s3.acl_mappings.length) {
          fields.push({name:'ACL 映射',value:JSON.stringify(s3.acl_mappings.map((a:Row)=>({key:a.key,type:placementText(a.val.type),source_id:placementText(a.val.source_id),dest_id:placementText(a.val.dest_id)})))})
        } else fields.push({name:'ACL 映射',value:'未返回或格式异常'})
      }
      fields.push({name:'目标标签',value:Array.isArray(target.tags)&&target.tags.every(v=>typeof v==='string')?JSON.stringify(target.tags):'未返回或格式异常'})
      rows.push({key:JSON.stringify([target.name,storageClass]),placement:target.name,storageClass,type,declared,fields})
    }
  }
  return {rows,issues}
}

export function zonePlacementClassRows(value:unknown) {
  const rows:PlacementClassRow[]=[],issues:string[]=[]
  if(!Array.isArray(value))return {rows,issues:['Zone 放置池配置未采集或格式异常']}
  const duplicateTargets=duplicates(value.filter(object).map(v=>v.key).filter(name))
  for(const entry of value) {
    if(!object(entry)||!name(entry.key)||duplicateTargets.has(entry.key)||!object(entry.val)||!object(entry.val.storage_classes)) {issues.push('Zone 放置配置身份重复、缺失或类映射异常');continue}
    const info=entry.val
    for(const [storageClass,config] of Object.entries(info.storage_classes)) {
      if(!name(storageClass)||!object(config)) {issues.push(`${entry.key}：类映射异常`);continue}
      const fields:PlacementClassRow['fields']=[]
      addFields(fields,config,{data_pool:'数据池（原生引用）'})
      fields.push({name:'压缩配置',value:config.compression_type===undefined?'未显式设置（原生空值）':placementText(config.compression_type)})
      addFields(fields,info,{index_pool:'索引池（原生引用）',inline_data:'内联数据'})
      fields.push({name:'索引类型',value:info.index_type===0?'普通（0）':info.index_type===1?'无索引（1）':`未知：${placementText(info.index_type)}`})
      fields.push({name:'额外数据池',value:info.data_extra_pool===''?`原生空值，回退 STANDARD 数据池：${placementText(object(info.storage_classes.STANDARD)?info.storage_classes.STANDARD.data_pool:undefined)}`:placementText(info.data_extra_pool)})
      rows.push({key:JSON.stringify([entry.key,storageClass]),placement:entry.key,storageClass,type:'本地映射',declared:'不能据此确认组声明',fields})
    }
  }
  return {rows,issues}
}
