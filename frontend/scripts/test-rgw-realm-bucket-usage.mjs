import {math} from './test-bucket-usage-math.mjs'
import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwRealmBucketUsage.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='./bucketUsageMath'?math:name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Descriptions','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const times={started_at:'2026-01-01T00:00:00Z',observed_at:'2026-01-01T00:00:01Z'}
const row={...times,object_count:'9007199254740993',size_actual_bytes:'18446744073709551615',usage_category:'rgw.main',realm_id:'r1',zone_id:'z1',service_map_id:'1',bucket_count:'9007199254740993',scope:'zone',source:'radosgw-admin'}
const data={...times,object_count:'18014398509481986',size_actual_bytes:'36893488147419103230',usage_category:'rgw.main',scope:'registered_realms',source:'service_map+radosgw-admin',selection:'lowest_zone_id_then_service_map_id',realm_count:2,bucket_count:'18014398509481986',items:[row,{...row,realm_id:'r2',zone_id:'z2',service_map_id:'2'}]}
assert.deepEqual(ui.realmBucketUsageData({...data,secret:'omit',items:data.items.map(r=>({...r,secret:'omit'}))}),data)
for(const bad of [null,{}, {...data,scope:'all_realms'},{...data,source:'other'},{...data,selection:'random'},{...data,realm_count:1},{...data,bucket_count:'18014398509481985'},{...data,bucket_count:2},{...data,items:null},{...data,observed_at:'bad'}])assert.throws(()=>ui.realmBucketUsageData(bad))
for(const change of [{realm_id:'r2'},{zone_id:'z2'},{service_map_id:'2'},{realm_id:''},{zone_id:'z\n'},{scope:'other'},{source:'other'},{bucket_count:'01'},{bucket_count:'-1'},{bucket_count:null},{started_at:'2025-01-01T00:00:00Z'},{observed_at:'2027-01-01T00:00:00Z'}])assert.throws(()=>ui.realmBucketUsageData({...data,items:[{...row,...change},data.items[1]]}))
for(const field of ['object_count','size_actual_bytes']) {
 for(const v of [undefined,1,'-1','01','1']) {
  assert.throws(()=>ui.realmBucketUsageData({...data,[field]:v}))
  assert.throws(()=>ui.realmBucketUsageData({...data,items:[{...row,[field]:v},data.items[1]]}))
 }
}
assert.throws(()=>ui.realmBucketUsageData({...data,usage_category:'other'}))
assert.throws(()=>ui.realmBucketUsageData({...data,items:[{...row,usage_category:'other'},data.items[1]]}))
const zero={...data,realm_count:0,bucket_count:'0',object_count:'0',size_actual_bytes:'0',items:[]}
assert.deepEqual(ui.realmBucketUsageData(zero),zero)
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwRealmBucketUsageView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();assert.equal(calls.length,0)
const old=render().find(n=>n.type==='Button');old.props.onClick()
assert.equal(calls[0].path,'/rgw/realms/buckets/usage');assert.deepEqual(calls[0].body,{cluster_id:7});assert.equal(calls[0].cache,'no-store')
calls[0].resolve(data);await tick()
assert.equal(render().find(n=>n.type==='Descriptions').props.items[1].children,'18014398509481986')
assert.deepEqual(render().find(n=>n.type==='Table').props.dataSource,data.items)
assert.equal(render().find(n=>n.type==='Descriptions').props.items[3].children,'36893488147419103230')
assert.equal(render().find(n=>n.type==='Descriptions').props.items.find(i=>i.key==='average').children,'2047.99 B/对象')
const averageColumn=render().find(n=>n.type==='Table').props.columns.find(c=>c.key==='average')
assert.equal(averageColumn.render(undefined,row),'2047.99 B/对象')
assert.equal(averageColumn.render(undefined,{...row,object_count:'0'}),'不适用（对象数为 0）')
// Unequal populations require total bytes / total objects, not an average of averages.
const weighted={...data,bucket_count:'2',object_count:'4',size_actual_bytes:'16',items:[{...row,bucket_count:'1',object_count:'1',size_actual_bytes:'10'},{...data.items[1],bucket_count:'1',object_count:'3',size_actual_bytes:'6'}]}
assert.equal(math.averageBucketObjectBytes(ui.realmBucketUsageData(weighted)),'4.00 B/对象')
old.props.onClick();assert.equal(render().some(n=>n.type==='Table'),false);calls[1].resolve({...data,bucket_count:'1'});await tick();assert.ok(render().some(n=>n.type==='Alert'&&n.props.type==='error'))
old.props.onClick();render(8);assert.equal(calls[2].signal.aborted,true);old.props.onClick();assert.equal(calls.length,3);calls[2].resolve(data);await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
render(8).find(n=>n.type==='Button').props.onClick();calls[3].resolve(zero);await tick();assert.ok(render(8).some(n=>n.type==='Alert'&&n.props.type==='warning'))
render(8).find(n=>n.type==='Button').props.onClick();calls[4].reject(new Error('private'));await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
render(8).find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[5].signal.aborted,true);calls[5].resolve(data);await tick();assert.equal(states[0].data,undefined)
render(null);assert.equal(render(null).find(n=>n.type==='Button').props.disabled,true)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwRealmBucketUsage \/>/)
console.log('Registered realm bucket aggregates preserve exact totals, validate representatives and isolate reads')
