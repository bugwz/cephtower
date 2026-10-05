import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const placement={},data={},calls=[]
let replies=[]
new Function('exports',compile('rgwPlacementClassData.ts'))(placement)
const client={jsonInit:(method,body,opts)=>({method,body,...opts}),request:async(path,init)=>{
 calls.push({path,...init});const value=replies.shift();if(value instanceof Error)throw value;return typeof value==='function'?value():value
}}
new Function('exports','require',compile('rgwStorageClassInventory.ts'))(data,name=>name==='./rgwPlacementClassData'?placement:client)
const tier={key:'COLD',val:{storage_class:'COLD',tier_type:'cloud-s3',s3:{region:'north',endpoint:'https://user:private-password@cloud.example/path?token=private-token#private-fragment',access_key:'private-access',secret:'private-secret',acl_mappings:[]},future:{secret:'private-nested'}}}
const group={id:'g',name:'group',realm_id:'r',stale:false,observed_at:'now',placement_targets:[{name:'p',tags:[],storage_classes:['STANDARD','COLD'],tier_targets:[tier]}],unknown:'private-top'}
const other={...group,id:'g2',name:'group',stale:true,placement_targets:[{...group.placement_targets[0],name:'q'},{...group.placement_targets[0],name:'p'}]}
const inventory=data.storageClassInventory([group,other])
assert.equal(inventory.rows.length,6);assert.equal(new Set(inventory.rows.map(r=>r.key)).size,6);assert.equal(inventory.groups,2);assert.equal(inventory.stale,true)
assert.equal(inventory.rows[0].realm,'r');assert.equal(inventory.rows[0].observed,'now')
assert.doesNotMatch(JSON.stringify(inventory),/private-/)
assert.equal(data.filterStorageClasses(inventory.rows,'','g').length,2)
assert.equal(data.filterStorageClasses(inventory.rows,'','g2','云 S3').length,2)
assert.equal(data.filterStorageClasses(inventory.rows,'  NORTH  ').length,3)
assert.equal(data.filterStorageClasses(inventory.rows,'cloud.example').length,3)
assert.equal(data.filterStorageClasses(inventory.rows,'private-secret').length,0)
assert.equal(data.filterStorageClasses(inventory.rows,'missing').length,0)
for(const rows of [[group,group],[{...group,id:''}],[{...group,name:undefined}],[{...group,stale:undefined}],[null]])assert.throws(()=>data.storageClassInventory(rows))
assert.equal(data.storageClassInventory([{...group,realm_id:''}]).rows[0].realm,'未关联 Realm')
assert.equal(data.storageClassInventory([{...group,realm_id:undefined}]).rows[0].realm,'Realm 未返回')
assert.deepEqual(data.storageClassInventory([]),{rows:[],issues:[],stale:false,groups:0})
const malformed=data.storageClassInventory([{...group,placement_targets:[{...group.placement_targets[0],tier_targets:null}]}])
assert.equal(malformed.rows[0].type,'未知');assert.ok(malformed.issues.length)
const orphan=data.storageClassInventory([{...group,placement_targets:[{...group.placement_targets[0],storage_classes:['STANDARD']}]}])
assert.match(orphan.rows.find(r=>r.storageClass==='COLD').declared,/声明缺失/)
const page=(groups,next=null,stale=false)=>({items:groups.map(g=>({data:g,stale:g.stale,observed_at:g.observed_at})),meta:{stale},pagination:{next_cursor:next}})
replies=[page([group],'next'),page([other])]
const signal=new AbortController().signal
assert.equal((await data.readStorageClassInventory(7,signal)).rows.length,6)
assert.equal(calls.length,2);assert.match(calls[1].path,/cursor=next/)
assert.ok(calls.every(c=>c.method==='GET'&&c.body.cluster_id===7&&c.cache==='no-store'&&c.signal===signal&&c.suppressErrorNotification))
for(const bad of [[new Error('offline')],[{}],[page([],'same'),page([],'same')],[page([group],'next'),new Error('failed second page')],[page([group],'next'),page([group])],[{...page([group]),meta:{}}],[{...page([group]),items:[{data:group}]}]]) {
 replies=bad;await assert.rejects(data.readStorageClassInventory(7))
}
replies=[page([],null,true)];assert.equal((await data.readStorageClassInventory(7)).stale,true)
replies=Array.from({length:100},(_,i)=>page([],String(i)));await assert.rejects(data.readStorageClassInventory(7),/limit/)
for(const id of [0,-1,1.5,NaN,Number.MAX_SAFE_INTEGER+1])await assert.rejects(data.readStorageClassInventory(id))
const cancelled=new AbortController();cancelled.abort();const before=calls.length
await assert.rejects(data.readStorageClassInventory(7,cancelled.signal));assert.equal(calls.length,before)
const during=new AbortController();replies=[()=>{during.abort();return page([group],'next')}]
await assert.rejects(data.readStorageClassInventory(7,during.signal));assert.equal(calls.length,before+1)

const ui={},jsx=(type,props)=>({type,props}),states=[],refs=[],requests=[]
let stateIndex=0,refIndex=0,cleanup,effectDeps,pendingEffect
const react={useRef:value=>refs[refIndex++]??(refs[refIndex-1]={current:value}),useState:value=>{
 const i=stateIndex++;if(!(i in states))states[i]=value
 return [states[i],next=>{states[i]=typeof next==='function'?next(states[i]):next}]
},useEffect:(fn,deps)=>{if(JSON.stringify(deps)!==JSON.stringify(effectDeps)){pendingEffect=()=>{cleanup?.();cleanup=fn()};effectDeps=deps}}}
new Function('exports','require',compile('RgwStorageClassesPage.tsx'))(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Input','Select','Space','Table','Tag'].map(n=>[n,n])):name==='react-router-dom'?{Link:'Link'}:name==='./rgwStorageClassInventory'?{...data,readStorageClassInventory:(id,signal)=>new Promise((resolve,reject)=>requests.push({id,signal,resolve,reject}))}:name==='./RgwPlacementClasses'?{renderPlacementClassDetails:()=>null}:name.includes('ClusterContext')?{useClusterContext:()=>({selectedClusterId:7})}:{jsx,jsxs:jsx})
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children),...nodes(n.props?.extra)]:[]}
function render(id){stateIndex=refIndex=0;const view=ui.RgwStorageClassesView({clusterId:id});const effect=pendingEffect;pendingEffect=undefined;effect?.();return view}
const tick=()=>new Promise(resolve=>setTimeout(resolve,0))
render(7);assert.equal(requests.length,1)
requests[0].resolve(inventory);await tick()
let view=render(7),table=nodes(view).find(n=>n.type==='Table')
assert.equal(table.props.dataSource.length,6)
assert.ok(table.props.columns.some(c=>c.title==='组 ID'));assert.ok(table.props.columns.some(c=>c.title==='采集时间'))
assert.equal(nodes(view).filter(n=>n.type==='Link').length,2)
nodes(view).find(n=>n.type==='Input').props.onChange({target:{value:'north'}})
assert.equal(nodes(render(7)).find(n=>n.type==='Table').props.dataSource.length,3)
nodes(view).find(n=>n.type==='Button').props.onClick()
assert.equal(nodes(render(7)).some(n=>n.type==='Table'),false)
assert.equal(requests.length,2)
render(8);assert.equal(requests[1].signal.aborted,true)
requests[1].resolve(inventory);await tick();assert.equal(states[0].data,undefined);assert.equal(states[0].clusterId,8)
requests[2].reject(new Error('offline'));await tick()
assert.equal(nodes(render(8)).some(n=>n.type==='Table'),false)
assert.ok(nodes(render(8)).some(n=>n.type==='Alert'&&n.props.type==='error'))
nodes(render(8)).find(n=>n.type==='Button').props.onClick()
cleanup();requests[3].resolve(inventory);await tick();assert.equal(states[0].data,undefined);assert.equal(requests[3].signal.aborted,true)
render(undefined);assert.equal(nodes(render(undefined)).find(n=>n.type==='Button').props.disabled,true)
const navigation=readFileSync(new URL('../src/navigation.ts',import.meta.url),'utf8')
assert.match(navigation,/key: 'rgwStorageClasses'.*path: '\/object\/storage-class'.*permission: 'storage'/)
assert.match(readFileSync(new URL('../src/pages/index.ts',import.meta.url),'utf8'),/rgwStorageClasses: RgwStorageClassesPage/)
console.log('Storage class inventory preserves scoped identities, complete pagination, safe fields and request isolation')
