import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwDaemonStatus.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending,active=true
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Descriptions','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const data={service_map_id:'12',status_stamp:'stamp',last_beacon:'beacon',observed_at:'now',status:{json:'{"count":9007199254740993}',empty:'',html:'<script>bad()</script>'}}
assert.deepEqual(ui.daemonStatusData({...data,extra:'private'},'12'),data)
for(const bad of [null,{}, {...data,service_map_id:'13'},{...data,status:[]},{...data,status:{count:1}},{...data,status:null}])assert.throws(()=>ui.daemonStatusData(bad,'12'))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id='12'){si=ri=0;const tree=nodes(ui.RgwDaemonStatus({clusterId:7,serviceMapId:id,metadata:{os:'<b>Linux</b>'},isCurrent:()=>active}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
assert.equal(render().find(n=>n.type==='Descriptions').props.items[0].children.props.children,'<b>Linux</b>')
render();const old=render().find(n=>n.type==='Button');old.props.onClick();assert.equal(calls.length,1);assert.deepEqual(calls[0].body,{cluster_id:7,service_map_id:'12'});assert.equal(calls[0].path,'/rgw/daemon/status');assert.equal(calls[0].cache,'no-store')
calls[0].resolve(data);await tick();let tree=render();const table=tree.find(n=>n.type==='Table');assert.equal(table.props.dataSource[0].value,data.status.json);assert.equal(table.props.columns[1].render(null,{value:data.status.html}).props.children,data.status.html)
old.props.onClick();render('13');assert.equal(calls[1].signal.aborted,true);old.props.onClick();assert.equal(calls.length,2);calls[1].resolve(data);await tick();assert.equal(render('13').some(n=>n.type==='Table'),false)
const button=render('13').find(n=>n.type==='Button');active=false;button.props.onClick();assert.equal(calls.length,2);active=true;button.props.onClick();calls[2].reject(new Error('private'));await tick();assert.ok(render('13').some(n=>n.type==='Alert'&&n.props.type==='error'))
render('13').find(n=>n.type==='Button').props.onClick();cleanup();assert.equal(calls[3].signal.aborted,true);calls[3].resolve({...data,service_map_id:'13'});await tick();assert.equal(states[0].data,undefined)
console.log('RGW status details preserve text precision and isolate scoped asynchronous reads')
