import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,jsx:ts.JsxEmit.ReactJSX,target:ts.ScriptTarget.ES2022}}).outputText
const placement={},data={},calls=[]
let replies=[]
new Function('exports',compile('rgwPlacementClassData.ts'))(placement)
new Function('exports','require',compile('rgwLocalClassMappings.ts'))(data,name=>name==='./rgwPlacementClassData'?placement:{jsonInit:(...args)=>args,request:async(...args)=>{calls.push(args);const reply=replies.shift();if(reply instanceof Error)throw reply;return reply}})
const group={id:'g',realm_id:'r',zones:[{id:'z',name:'zone'},{id:'z2',name:'zone2'}],placement_targets:[{name:'p',storage_classes:['STANDARD','COLD'],tier_targets:[{key:'COLD',val:{storage_class:'COLD',tier_type:'cloud-s3'}}]}]}
const zone={id:'z',name:'zone',realm_id:'r',observed_at:'2026-10-05',placement_pools:[{key:'p',val:{storage_classes:{STANDARD:{data_pool:'data:ns',compression_type:'zstd'}}}}],system_key:{secret_key:'private-secret'}}
const mapped=data.localClassMappings(group,[zone,{...zone,id:'outside'}])
assert.equal(mapped.rows.length,2)
assert.equal(mapped.rows[0].pool,'data:ns')
assert.equal(mapped.rows[0].compression,'zstd')
assert.match(mapped.rows[1].status,/未找到成员库存/)
assert.doesNotMatch(JSON.stringify(mapped),/private-secret|outside|COLD/)
const both=data.localClassMappings(group,[zone,{...zone,id:'z2',name:'zone2'}])
assert.equal(both.rows.filter(r=>r.pool==='data:ns').length,2)
for(const changed of [{realm_id:'other'},{name:'renamed'},{realm_id:undefined}]) {
 const row=data.localClassMappings(group,[{...zone,...changed}]).rows[0]
 assert.equal(row.pool,'未知');assert.match(row.status,/不一致/)
}
assert.match(data.localClassMappings(group,[{...zone,stale:true}]).rows[0].status,/过期/)
assert.match(data.localClassMappings(group,[{...zone,placement_pools:[]}]).rows[0].status,/未列出/)
for(const pools of [undefined,null,[...zone.placement_pools,...zone.placement_pools]])assert.match(data.localClassMappings(group,[{...zone,placement_pools:pools}]).rows[0].status,/歧义/)
assert.throws(()=>data.localClassMappings(group,[zone,zone]))
assert.throws(()=>data.localClassMappings({...group,zones:[group.zones[0],group.zones[0]]},[zone]))
assert.equal(data.localClassMappings({...group,placement_targets:[{...group.placement_targets[0],tier_targets:null}]},[zone]).rows.length,0)
assert.ok(data.localClassMappings({...group,zones:[]},[]).issues.length)
const page=(items,next=null,stale=false)=>({items:items.map(data=>({data,stale:false,observed_at:'now'})),meta:{stale},pagination:{next_cursor:next}})
replies=[page([zone],'next'),page([{...zone,id:'z2'}],null,true)]
const inventory=await data.readLocalClassZones(42)
assert.equal(inventory.rows.length,2);assert.equal(inventory.stale,true)
assert.equal(calls[0][1][1].cluster_id,42);assert.equal(calls[0][1][2].cache,'no-store')
assert.match(calls[1][0],/cursor=next/)
for(const input of [[new Error('offline')],[{}],[page([],'same'),page([],'same')],[page([],'next'),new Error('page failure')],[{...page([]),items:[{data:zone}]}]]) {
 replies=input;await assert.rejects(data.readLocalClassZones(42))
}
const jsx=(type,props)=>({type,props}),ui={};let state
new Function('exports','require',compile('RgwLocalClassDetails.tsx'))(ui,name=>name==='react'?{useEffect:()=>{},useRef:v=>({current:v}),useState:v=>[v,x=>{state=x}]}:name==='antd'?{Alert:'Alert',Button:'Button',Space:'Space',Table:'Table'}:name==='./rgwLocalClassMappings'?data:{jsx,jsxs:jsx})
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
const view=ui.RgwLocalClassDetails({row:group,clusterId:42}),button=nodes(view).find(n=>n.type==='Button')
replies=[page([zone])];button.props.onClick();await new Promise(resolve=>setTimeout(resolve,0))
assert.equal(state.data.rows[0].pool,'data:ns')
assert.equal(nodes(ui.RgwLocalClassDetails({row:group})).find(n=>n.type==='Button').props.disabled,true)
// Keep refs across renders to verify that a previous scope cannot publish a late result.
const raceUI={},refs=[];let refIndex=0,raceState,cleanup,effectDeps,resolvePending
const react={useRef:v=>refs[refIndex++]??(refs[refIndex-1]={current:v}),useState:v=>[raceState??v,x=>{raceState=x}],useEffect:(fn,deps)=>{
 if(JSON.stringify(deps)!==JSON.stringify(effectDeps)){cleanup?.();cleanup=fn();effectDeps=deps}
}}
new Function('exports','require',compile('RgwLocalClassDetails.tsx'))(raceUI,name=>name==='react'?react:name==='antd'?{Alert:'Alert',Button:'Button',Space:'Space',Table:'Table'}:name==='./rgwLocalClassMappings'?{...data,readLocalClassZones:()=>new Promise(resolve=>{resolvePending=resolve})}:{jsx,jsxs:jsx})
function render(clusterId){refIndex=0;return raceUI.RgwLocalClassDetails({row:group,clusterId})}
nodes(render(1)).find(n=>n.type==='Button').props.onClick()
render(2);resolvePending({rows:[zone],stale:false});await new Promise(resolve=>setTimeout(resolve,0))
assert.equal(raceState.data,undefined)
assert.equal(JSON.parse(raceState.scope)[0],2)
nodes(render(2)).find(n=>n.type==='Button').props.onClick()
cleanup();resolvePending({rows:[zone],stale:false});await new Promise(resolve=>setTimeout(resolve,0))
assert.equal(raceState.data,undefined)
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/rgwZonegroups:\s*\{[\s\S]*?detailContent:.*RgwLocalClassDetails/)
console.log('Local class mappings verify member identity, all matching zones, stale data and complete pagination')
