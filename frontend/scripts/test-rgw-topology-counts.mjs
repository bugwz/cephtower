import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwTopologyCounts.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Descriptions','Space'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const data={realm_count:0,zonegroup_count:2,zone_count:3,started_at:'2026-01-01T00:00:00Z',observed_at:'2026-01-01T00:00:01Z',source:'radosgw-admin'}
assert.deepEqual(ui.topologyCountData({...data,extra:'private'}),data)
for(const bad of [null,{}, {...data,source:'other'},{...data,realm_count:-1},{...data,zone_count:1.5},{...data,zone_count:'1'},{...data,zone_count:2**53},{...data,observed_at:'bad'},{...data,started_at:'2027-01-01T00:00:00Z'}])assert.throws(()=>ui.topologyCountData(bad))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwTopologyCountsView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();const old=render().find(n=>n.type==='Button');old.props.onClick();assert.equal(calls[0].path,'/rgw/topology/counts');assert.deepEqual(calls[0].body,{cluster_id:7});assert.equal(calls[0].cache,'no-store');calls[0].resolve(data);await tick()
assert.equal(render().find(n=>n.type==='Descriptions').props.items[0].children,'0')
old.props.onClick();render(8);assert.equal(calls[1].signal.aborted,true);old.props.onClick();assert.equal(calls.length,2);calls[1].resolve(data);await tick();assert.equal(render(8).some(n=>n.type==='Descriptions'),false)
render(8).find(n=>n.type==='Button').props.onClick();calls[2].reject(new Error('private'));await tick();assert.ok(render(8).some(n=>n.type==='Alert'&&n.props.type==='error'))
render(8).find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[3].signal.aborted,true);calls[3].resolve(data);await tick();assert.equal(states[0].data,undefined)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwTopologyCounts \/>/)
console.log('RGW topology counts preserve zero, validate complete responses and isolate cluster reads')
