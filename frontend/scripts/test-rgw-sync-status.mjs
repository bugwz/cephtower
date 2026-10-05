import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const refs=[],states=[],effects=[],pending=[]
let cursor=0,resolve,reject
const hooks={useRef(v){const i=cursor++;return refs[i]??={current:v}},useState(v){const i=cursor++;if(!(i in states))states[i]=v;return [states[i],v=>{states[i]=v}]},useEffect(fn,deps){const i=cursor++;if(!effects[i]||deps.some((v,n)=>v!==effects[i].deps[n]))pending.push(()=>{effects[i]?.cleanup?.();effects[i]={deps,cleanup:fn()}})}}
const calls=[]
const client={jsonInit:(method,body,options)=>({method,body,...options}),request:(path,init)=>{calls.push({path,...init});return new Promise((yes,no)=>{resolve=yes;reject=no})}}
const api={}
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwSyncStatus.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.React}}).outputText
new Function('exports','require','React',code)(api,name=>name==='react'?hooks:name==='antd'?{Alert:'Alert',Button:'Button',Card:'Card',Space:'Space'}:client,{createElement:(type,props,...children)=>({type,props:{...props,children}})})
let props={row:{id:'zone-id',name:'zone-a'},clusterId:7}
function render(){cursor=0;const tree=api.RgwSyncStatus(props);pending.splice(0).forEach(fn=>fn());return tree}
function nodes(node){return !node||typeof node!=='object'?[]:[node,...(node.props?.children??[]).flat(Infinity).flatMap(nodes)]}
const find=type=>nodes(render()).find(n=>n.type===type)
const settle=async()=>{await Promise.resolve();await Promise.resolve()}
render();assert.equal(calls.length,0)
const click=find('Button').props.onClick
click();click();assert.equal(calls.length,1)
assert.deepEqual(calls[0],{path:'/rgw/zone/sync/status',method:'POST',body:{cluster_id:7,zone_id:'zone-id',name:'zone-a'},cache:'no-store',suppressErrorNotification:true})
const report='metadata sync no sync (zone is master)\ndata sync source: a\nbehind on 2 shards\n<script>not html</script>'
resolve({report});await settle();assert.deepEqual(find('pre').props.children,[report])
find('Button').props.onClick();assert.equal(find('pre'),undefined)
const oldResolve=resolve
props={...props,clusterId:8};render();oldResolve({report});await settle();assert.equal(find('pre'),undefined)
find('Button').props.onClick();reject(new Error('private diagnostics'));await settle();assert.ok(!JSON.stringify(render()).includes('private diagnostics'))
for(const bad of [null,{}, {report:''},{report:3},{report:'x'.repeat(1048577)}]){
  find('Button').props.onClick();resolve(bad);await settle();assert.equal(find('pre'),undefined)
}
find('Button').props.onClick();const oldZoneResolve=resolve
props={...props,row:{id:'other',name:'other'}};render();oldZoneResolve({report});await settle();assert.equal(find('pre'),undefined)
find('Button').props.onClick();effects.forEach(effect=>effect?.cleanup?.());resolve({report});await settle();assert.equal(find('pre'),undefined)
props={...props,row:{...props.row,stale:true}};render();assert.equal(find('Button').props.disabled,true)
const before=calls.length;find('Button').props.onClick();assert.equal(calls.length,before)
const pages=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(pages,/rgwZones: \{[\s\S]*?detailContent:.*<RgwSyncStatus/)
console.log('Zone sync reports are on demand, verbatim and invalidated on scope changes')
