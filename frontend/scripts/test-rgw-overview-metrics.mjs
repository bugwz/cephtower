import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwOverviewMetrics.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],calls=[];let si=0,ri=0,deps,cleanup,pending
const jsx=(type,props)=>({type,props}),react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pending=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Space','Table'].map(n=>[n,n])):name.includes('api/client')?{jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>calls.push({path,...init,resolve,reject}))}:{jsx,jsxs:jsx})
const sample=value=>({result_type:'vector',series:[{metric:{},value:[123,value]}]})
assert.equal(ui.overviewMetricValue(sample('9007199254740993')).value,'9007199254740993')
assert.equal(ui.overviewMetricValue(sample('0')).status,'已读取')
assert.equal(ui.overviewMetricValue(sample('NaN')).status,'不可计算')
assert.equal(ui.overviewMetricValue({result_type:'vector',series:[]}).status,'无样本')
for(const bad of [null,{},sample('-1'),sample(''),sample('bad'),sample(1),{...sample('1'),series:[...sample('1').series,...sample('1').series]}])assert.throws(()=>ui.overviewMetricValue(bad))
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){si=ri=0;const tree=nodes(ui.RgwOverviewMetricsView({clusterId:id}));const effect=pending;pending=undefined;effect?.();return tree}
const tick=()=>new Promise(r=>setTimeout(r,0))
render();const old=render().find(n=>n.type==='Button');old.props.onClick();assert.equal(calls.length,5)
const times=calls.map(call=>{assert.deepEqual(call.body,{cluster_id:7});assert.equal(call.cache,'no-store');return new URLSearchParams(call.path.split('?')[1]).get('time')});assert.equal(new Set(times).size,1)
calls[0].resolve(sample('0'));calls[1].reject(new Error('private'));calls[2].resolve({result_type:'vector',series:[]});calls[3].resolve(sample('NaN'));calls[4].resolve(sample('1.5'));await tick()
const table=render().find(n=>n.type==='Table');assert.deepEqual(table.props.dataSource.map(row=>row.status),['已读取','读取失败或响应无效','无样本','不可计算','已读取'])
assert.equal(table.props.columns[1].render('0'),'0')
old.props.onClick();render(8);assert.ok(calls.slice(5).every(call=>call.signal.aborted));old.props.onClick();assert.equal(calls.length,10);calls.slice(5).forEach(call=>call.resolve(sample('1')));await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
render(8).find(n=>n.type==='Button').props.onClick();cleanup();assert.ok(calls.slice(10).every(call=>call.signal.aborted));calls.slice(10).forEach(call=>call.resolve(sample('2')));await tick();assert.equal(states[0].rows,undefined)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/<RgwOverviewMetrics \/>/)
console.log('RGW overview metrics preserve independent states and isolate cluster reads')
