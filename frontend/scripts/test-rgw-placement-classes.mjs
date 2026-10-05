import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const data={},ui={},jsx=(type,props)=>({type,props})
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,jsx:ts.JsxEmit.ReactJSX,target:ts.ScriptTarget.ES2022}}).outputText
new Function('exports',compile('rgwPlacementClassData.ts'))(data)
new Function('exports','require',compile('RgwPlacementClasses.tsx'))(ui,name=>name==='antd'?{Alert:'Alert',Table:'Table',Descriptions:'Descriptions'}:name==='./rgwPlacementClassData'?data:{jsx,jsxs:jsx})
const tier={tier_type:'cloud-s3-glacier',storage_class:'COLD',retain_head_object:false,allow_read_through:true,read_through_restore_days:0,restore_storage_class:'STANDARD',s3:{endpoint:'https://user:password@example.com/path?token=private#secret',region:'us-test-1',access_key:'sensitive-access',secret:'sensitive-secret',host_style:'path',target_storage_class:'GLACIER',target_path:'bucket/prefix',multipart_sync_threshold:0,multipart_min_part_size:5,acl_mappings:[{key:'owner',val:{type:'id',source_id:'source',dest_id:'dest',secret:'hidden-acl'}}]},'s3-glacier':{glacier_restore_days:3,glacier_restore_tier_type:'Expedited'},future:{secret:'hidden-future'}}
const targets=[{name:'p',tags:['tag'],storage_classes:['STANDARD','COLD','LOCAL'],tier_targets:[{key:'COLD',val:tier}]}]
const result=data.groupPlacementClassRows(targets)
assert.equal(result.issues.length,0)
assert.deepEqual(result.rows.map(r=>[r.storageClass,r.type]),[['STANDARD','本地'],['COLD','云 S3 Glacier'],['LOCAL','本地']])
const cold=result.rows[1],field=(row,name)=>row.fields.find(f=>f.name===name)?.value
assert.equal(field(cold,'保留头对象'),'否')
assert.equal(field(cold,'读穿透恢复天数'),'0')
assert.equal(field(cold,'Glacier 恢复类型'),'Expedited')
assert.match(field(cold,'目标端点'),/^https:\/\/example.com\/path（/)
assert.deepEqual(JSON.parse(field(cold,'ACL 映射')),[{key:'owner',type:'id',source_id:'source',dest_id:'dest'}])
assert.doesNotMatch(JSON.stringify(result),/password|private|sensitive-|hidden-/)
for(const bad of ['javascript:private-secret','malformed-private-secret']) {
 const badTarget=[{...targets[0],tier_targets:[{key:'COLD',val:{...tier,s3:{...tier.s3,endpoint:bad}}}]}]
 const badRows=data.groupPlacementClassRows(badTarget)
 assert.doesNotMatch(JSON.stringify(badRows),/private-secret/)
 assert.match(field(badRows.rows[1],'目标端点'),/隐藏/)
}
const orphan=data.groupPlacementClassRows([{...targets[0],storage_classes:['STANDARD']}])
assert.match(orphan.rows.find(r=>r.storageClass==='COLD').declared,/声明缺失/)
for(const tiers of [null,{},[{key:'COLD',val:tier},{key:'COLD',val:tier}]]) {
 const invalid=data.groupPlacementClassRows([{...targets[0],tier_targets:tiers}])
 assert.ok(invalid.issues.length)
 assert.ok(invalid.rows.every(r=>r.type==='未知'))
}
assert.equal(data.groupPlacementClassRows([targets[0],targets[0]]).rows.length,0)
assert.ok(data.groupPlacementClassRows(undefined).issues.length)
assert.equal(data.groupPlacementClassRows([]).issues.length,0)
const zones=[{key:'p',val:{index_pool:'index',data_extra_pool:'',index_type:0,inline_data:false,storage_classes:{STANDARD:{data_pool:'std'},COLD:{data_pool:'data:ns',compression_type:'none'}}}}]
const zr=data.zonePlacementClassRows(zones)
assert.equal(zr.issues.length,0)
assert.equal(field(zr.rows[1],'数据池（原生引用）'),'data:ns')
assert.equal(field(zr.rows[0],'内联数据'),'否')
assert.match(field(zr.rows[0],'压缩配置'),/未显式设置/)
assert.equal(field(zr.rows[1],'压缩配置'),'none')
assert.match(field(zr.rows[1],'额外数据池'),/STANDARD 数据池：std/)
assert.ok(data.zonePlacementClassRows([{key:'p',val:{storage_classes:null}}]).issues.length)
assert.match(data.placementText(Number.MAX_SAFE_INTEGER+1),/精确范围/)
function nodes(node){return Array.isArray(node)?node.flatMap(nodes):node&&typeof node==='object'?[node,...nodes(node.props?.children)]:[]}
const view=ui.RgwPlacementClasses({value:targets})
const table=nodes(view).find(n=>n.type==='Table')
assert.equal(table.props.dataSource.length,3)
assert.ok(table.props.expandable.expandedRowRender(cold))
assert.ok(nodes(ui.RgwPlacementClasses({value:null})).some(n=>n.type==='Alert'))
const unknownView=ui.RgwPlacementClasses({value:[{...targets[0],tier_targets:null}]})
const unknownTable=nodes(unknownView).find(n=>n.type==='Table')
assert.equal(unknownTable.props.columns.find(c=>c.title==='目标端点').render(null,unknownTable.props.dataSource[0]),'未返回或不可用')
assert.ok(nodes(view).every(n=>!n.props?.dangerouslySetInnerHTML))
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/key: 'placement_targets'.*<RgwPlacementClasses/)
assert.match(source,/key: 'placement_pools'.*<RgwPlacementClasses value=\{value\} zone/)
console.log('Placement class tables preserve native class, tier, ACL and pool semantics without credentials')
