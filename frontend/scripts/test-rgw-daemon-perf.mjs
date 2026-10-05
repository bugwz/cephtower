import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwDaemonPerf.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending,active=true
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const sample={metric:{__name__:'ceph_rgw_req',instance_id:'12',job:'<script>bad()</script>'},value:[123,'9007199254740993']}
const data={service_map_id:'12',source:'prometheus',result_type:'vector',available:true,observed_at:'now',series:[sample,{...sample,metric:{...sample.metric,job:'other'}}]}
assert.deepEqual(ui.daemonPerfData({...data,extra:'private'},'12'),data)
for(const bad of [null,{}, {...data,service_map_id:'13'},{...data,available:false},{...data,source:'other'},{...data,series:[{...sample,value:[1,2]}]},{...data,series:[{...sample,metric:{...sample.metric,instance_id:'other'}}]},{...data,series:[{...sample,metric:{...sample.metric,__name__:'ceph_osd_op'}}]}])assert.throws(()=>ui.daemonPerfData(bad,'12'))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id='12'){si=ri=0;const tree=nodes(ui.RgwDaemonPerf({clusterId:7,serviceMapId:id,isCurrent:()=>active}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();const old=render().find(n=>n.type==='Button');old.props.onClick();assert.equal(calls.length,1);assert.deepEqual(calls[0].body,{cluster_id:7,service_map_id:'12'});assert.equal(calls[0].path,'/rgw/daemon/perf');assert.equal(calls[0].cache,'no-store')
calls[0].resolve(data);await tick();let table=render().find(n=>n.type==='Table');assert.equal(table.props.dataSource.length,2);assert.equal(table.props.columns[3].render(null,sample),'9007199254740993');assert.equal(table.props.columns[1].render(null,sample).props.children,JSON.stringify(sample.metric,null,2))
old.props.onClick();calls[1].resolve({...data,available:false,series:[]});await tick();assert.equal(render().some(n=>n.type==='Table'),false);assert.ok(render().some(n=>n.type==='Alert'&&n.props.type==='warning'))
old.props.onClick();render('13');assert.equal(calls[2].signal.aborted,true);old.props.onClick();assert.equal(calls.length,3);calls[2].resolve(data);await tick();assert.equal(render('13').some(n=>n.type==='Table'),false)
const button=render('13').find(n=>n.type==='Button');active=false;button.props.onClick();assert.equal(calls.length,3);active=true;button.props.onClick();calls[3].reject(new Error('private'));await tick();assert.ok(render('13').some(n=>n.type==='Alert'&&n.props.type==='error'))
render('13').find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[4].signal.aborted,true);calls[4].resolve({...data,service_map_id:'13'});await tick();assert.equal(states[0].data,undefined)
assert.match(readFileSync(new URL('../src/pages/object/RgwDaemonStatus.tsx',import.meta.url),'utf8'),/<RgwDaemonPerf clusterId=\{clusterId\} serviceMapId=\{serviceMapId\} isCurrent=\{isCurrent\}/)
console.log('RGW performance snapshots preserve series and precision and isolate asynchronous reads')
