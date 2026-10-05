import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const refs=[],states=[],effects=[],pending=[]
let cursor=0,resolve,reject
const hooks={useRef(v){const i=cursor++;return refs[i]??={current:v}},useState(v){const i=cursor++;if(!(i in states))states[i]=v;return [states[i],v=>{states[i]=v}]},useEffect(fn,deps){const i=cursor++;if(!effects[i]||deps.some((v,n)=>v!==effects[i].deps[n]))pending.push(()=>{effects[i]?.cleanup?.();effects[i]={deps,cleanup:fn()}})}}
const calls=[]
const client={jsonInit:(method,body,options)=>({method,body,...options}),request:(path,init)=>{calls.push({path,...init});return new Promise((yes,no)=>{resolve=yes;reject=no})}}
const api={}
const parser={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwSyncReport.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText)(parser)
const native='          realm r (realm)\n      zonegroup g (group)\n           zone z (zone)\n   current time 2026-10-05T00:00:00Z\n  metadata sync syncing\n                full sync: 2/64 shards\n      data sync source: a (first)\n                       failed to fetch source: diagnostic\n                source: b (second)\n                       full sync: 9007199254740993 entries to sync\n'
const parts=parser.rgwSyncReportSections(native)
assert.equal(parts.sources.length,2)
assert.match(parts.sources[0],/failed to fetch source: diagnostic/)
assert.match(parts.sources[1],/9007199254740993/)
assert.equal(parts.metadata,'syncing\nfull sync: 2/64 shards')
assert.equal(parser.rgwSyncReportSections(native.split('      data sync')[0]).sources.length,0)
assert.equal(parser.rgwSyncReportSections('unknown report'),undefined)
assert.match(parser.rgwSyncReportSections(native.replace('syncing','failed to read sync status: Permission denied')).metadata,/Permission denied/)
const counterText='syncing\nfull sync: 2/64 shards\nfull sync: 9007199254740993 buckets to sync\nincremental sync: 62/64 shards\ndata is behind on 4 shards\n1 shards are recovering\nfailed to fetch source sync status: Permission denied'
assert.deepEqual(parser.rgwSyncCounters(counterText),{full:'2',total:'64',remaining:'9007199254740993',remainingUnit:'buckets',incremental:'62',behind:'4',recovering:'1'})
assert.deepEqual(parser.rgwSyncCounters('full sync: 0/0 shards\nincremental sync: 0/0 shards\nmetadata is behind on 0 shards'),{full:'0',total:'0',incremental:'0',behind:'0'})
assert.deepEqual(parser.rgwSyncCounters('full sync: 18446744073709551615 entries to sync'),{remaining:'18446744073709551615',remainingUnit:'entries'})
for(const text of ['no sync (zone is master)','failed to fetch source sync status: full sync: 1/2 shards','full sync: -1/2 shards','full sync: 3/2 shards','full sync: 1/2 shards\nincremental sync: 1/3 shards','full sync: 2/2 shards\nincremental sync: 1/2 shards','full sync: 1/2 shards\nfull sync: 1/2 shards','full sync: 1 entries to sync\nfull sync: 1 buckets to sync'])assert.equal(parser.rgwSyncCounters(text),undefined)
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwSyncStatus.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.React}}).outputText
new Function('exports','require','React',code)(api,name=>name==='react'?hooks:name==='antd'?{Alert:'Alert',Button:'Button',Card:'Card',Space:'Space',Descriptions:'Descriptions'}:name==='./rgwSyncReport'?parser:client,{createElement:(type,props,...children)=>({type,props:{...props,children}})})
const counterItems=api.RgwSyncCounterDetails({section:counterText}).props.items
assert.equal(counterItems.find(item=>item.key==='remaining').children,'9007199254740993')
assert.match(counterItems.find(item=>item.key==='full').label,/非已完成/)
assert.equal(api.RgwSyncCounterDetails({section:'no sync (zone is master)'}),null)
let props={row:{id:'zone-id',name:'zone-a'},clusterId:7}
function render(){cursor=0;const tree=api.RgwSyncStatus(props);pending.splice(0).forEach(fn=>fn());return tree}
function nodes(node){return !node||typeof node!=='object'?[]:[node,...(node.props?.children??[]).flat(Infinity).flatMap(nodes)]}
const find=type=>nodes(render()).find(n=>n.type===type)
const settle=async()=>{await Promise.resolve();await Promise.resolve()}
render();assert.equal(calls.length,0)
const click=find('Button').props.onClick
click();click();assert.equal(calls.length,1)
assert.deepEqual(calls[0],{path:'/rgw/zone/sync/status',method:'POST',body:{cluster_id:7,zone_id:'zone-id',name:'zone-a'},signal:calls[0].signal,cache:'no-store',suppressErrorNotification:true})
assert.equal(calls[0].signal.aborted,false)
const report='metadata sync no sync (zone is master)\ndata sync source: a\nbehind on 2 shards\n<script>not html</script>'
resolve({report});await settle();assert.deepEqual(find('pre').props.children,[report])
find('Button').props.onClick();resolve({report:native});await settle()
assert.equal(nodes(render()).filter(node=>node.type==='Card'&&String(node.props.title).startsWith('数据同步来源')).length,2)
assert.equal(nodes(render()).find(node=>node.type==='pre'&&node.props['aria-label']==='Zone 原生同步报告').props.children[0],native)
assert.equal(find('details').props.open,false)
find('Button').props.onClick();assert.equal(find('pre'),undefined)
const oldResolve=resolve
const clusterSignal=calls.at(-1).signal
props={...props,clusterId:8};render();assert.equal(clusterSignal.aborted,true)
let oldCount=calls.length;click();assert.equal(calls.length,oldCount)
oldResolve({report});await settle();assert.equal(find('pre'),undefined)
find('Button').props.onClick();reject(new Error('private diagnostics'));await settle();assert.ok(!JSON.stringify(render()).includes('private diagnostics'))
for(const bad of [null,{}, {report:''},{report:3},{report:'x'.repeat(1048577)}]){
  find('Button').props.onClick();resolve(bad);await settle();assert.equal(find('pre'),undefined)
}
const zoneClick=find('Button').props.onClick;zoneClick();const oldZoneResolve=resolve,zoneSignal=calls.at(-1).signal
props={...props,row:{id:'other',name:'other'}};render();assert.equal(zoneSignal.aborted,true)
oldCount=calls.length;zoneClick();assert.equal(calls.length,oldCount)
oldZoneResolve({report});await settle();assert.equal(find('pre'),undefined)
const unmountClick=find('Button').props.onClick;unmountClick();const unmountSignal=calls.at(-1).signal
effects.forEach(effect=>effect?.cleanup?.());assert.equal(unmountSignal.aborted,true)
oldCount=calls.length;unmountClick();assert.equal(calls.length,oldCount)
resolve({report});await settle();assert.equal(find('pre'),undefined)
props={...props,row:{...props.row,stale:true}};render();assert.equal(find('Button').props.disabled,true)
const before=calls.length;find('Button').props.onClick();assert.equal(calls.length,before)
const pages=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(pages,/rgwZones: \{[\s\S]*?detailContent:.*<RgwSyncStatus/)
console.log('Zone sync reports are on demand, verbatim and invalidated on scope changes')
