import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwRealmUserCounts.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Descriptions','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const times={started_at:'2026-01-01T00:00:00Z',observed_at:'2026-01-01T00:00:01Z'}
const row={...times,realm_id:'r1',zone_id:'z1',service_map_id:'1',user_count:'9007199254740993',scope:'zone',source:'radosgw-admin'}
const data={...times,scope:'registered_realms',source:'service_map+radosgw-admin',selection:'lowest_zone_id_then_service_map_id',realm_count:2,user_count:'18014398509481986',items:[row,{...row,realm_id:'r2',zone_id:'z2',service_map_id:'2'}]}
assert.deepEqual(ui.realmUserCountData({...data,secret:'omit',items:data.items.map(r=>({...r,secret:'omit'}))}),data)
for(const bad of [null,{}, {...data,scope:'all_realms'},{...data,source:'other'},{...data,selection:'random'},{...data,realm_count:1},{...data,user_count:'18014398509481985'},{...data,user_count:2},{...data,items:null},{...data,observed_at:'bad'}])assert.throws(()=>ui.realmUserCountData(bad))
for(const change of [{realm_id:'r2'},{zone_id:'z2'},{service_map_id:'2'},{realm_id:''},{zone_id:'z\n'},{scope:'other'},{source:'other'},{user_count:'01'},{user_count:'-1'},{user_count:null},{started_at:'2025-01-01T00:00:00Z'},{observed_at:'2027-01-01T00:00:00Z'}])assert.throws(()=>ui.realmUserCountData({...data,items:[{...row,...change},data.items[1]]}))
const zero={...data,realm_count:0,user_count:'0',items:[]}
assert.deepEqual(ui.realmUserCountData(zero),zero)
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwRealmUserCountsView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();assert.equal(calls.length,0)
const old=render().find(n=>n.type==='Button');old.props.onClick()
assert.equal(calls[0].path,'/rgw/realms/users/counts');assert.deepEqual(calls[0].body,{cluster_id:7});assert.equal(calls[0].cache,'no-store')
calls[0].resolve(data);await tick()
assert.equal(render().find(n=>n.type==='Descriptions').props.items[1].children,'18014398509481986')
assert.deepEqual(render().find(n=>n.type==='Table').props.dataSource,data.items)
old.props.onClick();assert.equal(render().some(n=>n.type==='Table'),false);calls[1].resolve({...data,user_count:'1'});await tick();assert.ok(render().some(n=>n.type==='Alert'&&n.props.type==='error'))
old.props.onClick();render(8);assert.equal(calls[2].signal.aborted,true);old.props.onClick();assert.equal(calls.length,3);calls[2].resolve(data);await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
render(8).find(n=>n.type==='Button').props.onClick();calls[3].resolve(zero);await tick();assert.ok(render(8).some(n=>n.type==='Alert'&&n.props.type==='warning'))
render(8).find(n=>n.type==='Button').props.onClick();calls[4].reject(new Error('private'));await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
render(8).find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[5].signal.aborted,true);calls[5].resolve(data);await tick();assert.equal(states[0].data,undefined)
render(null);assert.equal(render(null).find(n=>n.type==='Button').props.disabled,true)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwRealmUserCounts \/>/)
console.log('Registered realm user aggregates preserve exact totals, validate representatives and isolate reads')
