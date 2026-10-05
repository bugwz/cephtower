import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwDaemonHistory.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending,active=true
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const series={metric:{__name__:'ceph_rgw_req',instance_id:'12',job:'<b>mgr</b>'},values:[[3600,'9007199254740993'],[7200,'NaN']]}
const data={service_map_id:'12',source:'prometheus',result_type:'matrix',available:true,observed_at:'now',start:'1970-01-01T01:00:00Z',end:'1970-01-01T02:00:00Z',step_seconds:60,series:[series,{...series,metric:{...series.metric,job:'other'}}]}
assert.deepEqual(ui.daemonHistoryData({...data,private:'extra'},'12'),data)
for(const bad of [null,{}, {...data,available:false},{...data,step_seconds:1},{...data,start:'bad'},{...data,end:data.start},{...data,service_map_id:'other'},{...data,series:[{...series,metric:{...series.metric,instance_id:'other'}}]}])assert.throws(()=>ui.daemonHistoryData(bad,'12'))
for(const values of [[],[[3599,'1']],[[7201,'1']],[[3600,'1'],[3600,'2']],[[7200,'1'],[3600,'2']],[[3600,1]],Array.from({length:62},(_,i)=>[3600+i,'1'])])assert.throws(()=>ui.daemonHistoryData({...data,series:[{...series,values}]},'12'))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id='12'){si=ri=0;const tree=nodes(ui.RgwDaemonHistory({clusterId:7,serviceMapId:id,isCurrent:()=>active}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();const old=render().find(n=>n.type==='Button');old.props.onClick();assert.equal(calls[0].path,'/rgw/daemon/perf/history');assert.deepEqual(calls[0].body,{cluster_id:7,service_map_id:'12'});assert.equal(calls[0].cache,'no-store')
calls[0].resolve(data);await tick();let table=render().find(n=>n.type==='Table');assert.equal(table.props.dataSource.length,2);assert.equal(table.props.expandable.expandedRowRender(series).props.dataSource[0].value,'9007199254740993');assert.equal(table.props.columns[1].render(null,series).props.children,JSON.stringify(series.metric,null,2))
old.props.onClick();calls[1].resolve({...data,available:false,series:[]});await tick();assert.equal(render().some(n=>n.type==='Table'),false);assert.ok(render().some(n=>n.type==='Alert'&&n.props.type==='warning'))
old.props.onClick();render('13');assert.equal(calls[2].signal.aborted,true);old.props.onClick();assert.equal(calls.length,3);calls[2].resolve(data);await tick();assert.equal(render('13').some(n=>n.type==='Table'),false)
const button=render('13').find(n=>n.type==='Button');active=false;button.props.onClick();assert.equal(calls.length,3);active=true;button.props.onClick();calls[3].reject(new Error('private'));await tick();assert.ok(render('13').some(n=>n.type==='Alert'&&n.props.type==='error'))
render('13').find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[4].signal.aborted,true);calls[4].resolve({...data,service_map_id:'13'});await tick();assert.equal(states[0].data,undefined)
assert.match(readFileSync(new URL('../src/pages/object/RgwDaemonPerf.tsx',import.meta.url),'utf8'),/<RgwDaemonHistory clusterId=\{clusterId\} serviceMapId=\{serviceMapId\} isCurrent=\{isCurrent\}/)
console.log('RGW history validates bounded series and preserves exact values with scoped reads')
