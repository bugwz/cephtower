import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwUserCount.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Descriptions','Space'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const data={user_count:0,scope:'current_rgw_configuration',source:'radosgw-admin',started_at:'2026-01-01T00:00:00Z',observed_at:'2026-01-01T00:00:01Z'}
assert.deepEqual(ui.userCountData({...data,users:['private']}),data)
for(const bad of [null,{}, {...data,source:'other'},{...data,scope:'all_realms'},{...data,scope:undefined},{...data,observed_at:'bad'},{...data,started_at:'2027-01-01T00:00:00Z'}])assert.throws(()=>ui.userCountData(bad))
for(const count of [-1,1.5,'0',null,undefined,NaN,Infinity,2**53])assert.throws(()=>ui.userCountData({...data,user_count:count}))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwUserCountView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();assert.equal(calls.length,0)
const old=render().find(n=>n.type==='Button');old.props.onClick()
assert.equal(calls[0].path,'/rgw/users/count');assert.deepEqual(calls[0].body,{cluster_id:7});assert.equal(calls[0].method,'GET');assert.equal(calls[0].cache,'no-store')
assert.equal(render().find(n=>n.type==='Button').props.disabled,true)
calls[0].resolve(data);await tick()
let items=render().find(n=>n.type==='Descriptions').props.items
assert.equal(items[0].children,'0');assert.match(items[1].children,/非跨 Realm/)
old.props.onClick();assert.equal(render().some(n=>n.type==='Descriptions'),false)
calls[1].resolve({...data,scope:'all_realms'});await tick();assert.ok(render().some(n=>n.type==='Alert'&&n.props.type==='error'))
old.props.onClick();render(8);assert.equal(calls[2].signal.aborted,true);old.props.onClick();assert.equal(calls.length,3)
calls[2].resolve({...data,user_count:100});await tick();assert.equal(render(8).some(n=>n.type==='Descriptions'),false)
render(8).find(n=>n.type==='Button').props.onClick();calls[3].reject(new Error('private'));await tick();assert.ok(render(8).some(n=>n.type==='Alert'&&n.props.type==='error'))
render(8).find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[4].signal.aborted,true);calls[4].resolve(data);await tick();assert.equal(states[0].data,undefined)
render(null);assert.equal(render(null).find(n=>n.type==='Button').props.disabled,true)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwUserCount \/>/)
console.log('RGW user counts validate scope, preserve zero and isolate cluster reads')
