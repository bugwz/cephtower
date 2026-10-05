import {math} from './test-bucket-usage-math.mjs'
import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketUsage.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='./bucketUsageMath'?math:name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Descriptions','Space'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const data={bucket_count:2,object_count:'18014398509481986',size_actual_bytes:'36893488147419103230',usage_category:'rgw.main',source:'radosgw-admin',scope:'current_rgw_configuration',started_at:'2026-01-01T00:00:00Z',observed_at:'2026-01-01T00:00:01Z'}
for(const scope of [undefined,null,'all_realms',''])assert.throws(()=>ui.bucketUsageData({...data,scope}))
assert.deepEqual(ui.bucketUsageData({...data,secret:'ignored'}),data)
for(const bad of [null,{}, {...data,source:'other'},{...data,usage_category:'rgw.multimeta'},{...data,bucket_count:-1},{...data,bucket_count:2**53},{...data,bucket_count:1.1},{...data,bucket_count:'2'},{...data,bucket_count:0},{...data,observed_at:'bad'},{...data,started_at:'2027-01-01T00:00:00Z'}])assert.throws(()=>ui.bucketUsageData(bad))
for(const key of ['object_count','size_actual_bytes'])for(const value of [null,undefined,0,9007199254740992,'','-1','1.2','1e2','01',' 1','NaN','<script>'])assert.throws(()=>ui.bucketUsageData({...data,[key]:value}))
assert.equal(math.averageBucketObjectBytes(data),'2047.99 B/对象')
assert.equal(math.averageBucketObjectBytes({...data,object_count:'3',size_actual_bytes:'10'}),'3.33 B/对象')
assert.equal(math.averageBucketObjectBytes({...data,object_count:'1'}),'36893488147419103230.00 B/对象')
const zero={...data,bucket_count:0,object_count:'0',size_actual_bytes:'0'}
assert.deepEqual(ui.bucketUsageData(zero),zero)
assert.equal(math.averageBucketObjectBytes(zero),'不适用（对象数为 0）')
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwBucketUsageView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();assert.equal(calls.length,0)
assert.ok(render().some(n=>n.type==='Alert'&&n.props.message.includes('不是跨 Realm 总量')))
const old=render().find(n=>n.type==='Button');old.props.onClick()
assert.equal(calls[0].path,'/rgw/buckets/usage');assert.deepEqual(calls[0].body,{cluster_id:7});assert.equal(calls[0].method,'GET');assert.equal(calls[0].cache,'no-store')
assert.equal(render().find(n=>n.type==='Button').props.disabled,true)
calls[0].resolve(data);await tick()
let items=render().find(n=>n.type==='Descriptions').props.items
assert.equal(items[1].children,data.object_count);assert.equal(items[2].children,data.size_actual_bytes)
old.props.onClick();assert.equal(render().some(n=>n.type==='Descriptions'),false)
calls[1].resolve({...data,object_count:100});await tick();assert.ok(render().some(n=>n.type==='Alert'&&n.props.type==='error'))
old.props.onClick();render(8);assert.equal(calls[2].signal.aborted,true);old.props.onClick();assert.equal(calls.length,3)
calls[2].resolve(data);await tick();assert.equal(render(8).some(n=>n.type==='Descriptions'),false)
render(8).find(n=>n.type==='Button').props.onClick();calls[3].resolve(zero);await tick()
items=render(8).find(n=>n.type==='Descriptions').props.items
assert.deepEqual(items.slice(0,3).map(i=>i.children),['0','0','0'])
render(8).find(n=>n.type==='Button').props.onClick();calls[4].reject(new Error('private'));await tick();assert.equal(render(8).some(n=>n.type==='Descriptions'),false)
render(8).find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[5].signal.aborted,true);calls[5].resolve(data);await tick();assert.equal(states[0].data,undefined)
render(null);assert.equal(render(null).find(n=>n.type==='Button').props.disabled,true)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwBucketUsage \/>/)
console.log('RGW bucket usage validates complete aggregates, preserves precision and isolates reads')
