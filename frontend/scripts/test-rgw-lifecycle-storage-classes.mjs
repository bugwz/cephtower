import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const data={},form={},editor={},calls=[],jsx=(type,props)=>({type,props})
const bucket={kind:'rgw_bucket',natural_key:'AGJ1Y2tldA',stale:false,observed_at:'now',data:{zonegroup:'g',placement_rule:'p/CURRENT'}}
const row={key:'g-p-COLD',groupId:'g',groupName:'group',placement:'p',storageClass:'COLD',type:'云 S3',declared:'已声明',stale:false,fields:[{name:'分层内部类名',value:'COLD'}]}
const inventory={groups:2,stale:false,issues:[],rows:[row,{...row,key:'local',storageClass:'LOCAL',type:'本地',fields:[]},{...row,key:'std',storageClass:'STANDARD',type:'本地'},{...row,key:'other-group',groupId:'g2'},{...row,key:'other-target',placement:'q'},{...row,key:'unknown',storageClass:'UNKNOWN',type:'未知'},{...row,key:'orphan',storageClass:'ORPHAN',declared:'仅分层配置存在（声明缺失）'},{...row,key:'mismatch',storageClass:'MISMATCH'}]}
let bucketReply=bucket,inventoryReply=inventory
const client={jsonInit:(method,body,opts)=>({method,body,...opts}),request:async(path,init)=>{calls.push({path,...init});if(bucketReply instanceof Error)throw bucketReply;return typeof bucketReply==='function'?bucketReply():bucketReply}}
const inventoryApi={readStorageClassInventory:async(id,signal)=>{calls.push({inventory:true,id,signal});if(inventoryReply instanceof Error)throw inventoryReply;return inventoryReply}}
new Function('exports','require',compile('rgwLifecycleStorageClasses.ts'))(data,name=>name==='./rgwStorageClassInventory'?inventoryApi:client)
const result=data.lifecycleStorageClasses(bucket,bucket.natural_key,inventory)
assert.deepEqual(result.options.map(o=>o.value),['COLD','LOCAL']);assert.equal(result.placement,'p');assert.equal(result.groupId,'g')
assert.match(result.options[0].label,/group \(g\) \/ p/)
assert.equal(data.lifecycleStorageClasses({...bucket,data:{...bucket.data,placement_rule:'p'}},bucket.natural_key,inventory).options.length,2)
for(const bad of [{...bucket,natural_key:'different'},{...bucket,kind:'other'},{...bucket,stale:true},{...bucket,stale:undefined},{...bucket,data:{zonegroup:'',placement_rule:'p'}},{...bucket,data:{zonegroup:'g',placement_rule:''}},{...bucket,data:{zonegroup:'g',placement_rule:'/COLD'}},{...bucket,data:{zonegroup:'missing',placement_rule:'p'}},null])assert.throws(()=>data.lifecycleStorageClasses(bad,bucket.natural_key,inventory))
assert.throws(()=>data.lifecycleStorageClasses(bucket,bucket.natural_key,{...inventory,stale:true}))
assert.throws(()=>data.lifecycleStorageClasses(bucket,bucket.natural_key,{...inventory,rows:[row,row]}))
assert.equal(data.lifecycleStorageClasses(bucket,bucket.natural_key,{...inventory,rows:[{...row,type:'本地',storageClass:'STANDARD'}]}).options.length,0)
const signal=new AbortController().signal
await data.readLifecycleStorageClasses(7,bucket.natural_key,signal)
assert.deepEqual(calls[0].body,{cluster_id:7,bucket_id:bucket.natural_key});assert.equal(calls[0].path,'/rgw/bucket');assert.equal(calls[0].method,'GET');assert.equal(calls[0].signal,signal);assert.equal(calls[0].cache,'no-store');assert.equal(calls[1].id,7);assert.equal(calls[1].signal,signal)
bucketReply=new Error('offline');await assert.rejects(data.readLifecycleStorageClasses(7,bucket.natural_key));bucketReply=bucket
inventoryReply=new Error('partial page');await assert.rejects(data.readLifecycleStorageClasses(7,bucket.natural_key));inventoryReply=inventory
for(const [cluster,id] of [[0,bucket.natural_key],[7,'bad/value'],[7,'']])await assert.rejects(data.readLifecycleStorageClasses(cluster,id))
const cancelled=new AbortController();cancelled.abort();let before=calls.length
await assert.rejects(data.readLifecycleStorageClasses(7,bucket.natural_key,cancelled.signal));assert.equal(calls.length,before)
const during=new AbortController();bucketReply=()=>{during.abort();return bucket};before=calls.length
await assert.rejects(data.readLifecycleStorageClasses(7,bucket.natural_key,during.signal));assert.equal(calls.length,before+1);bucketReply=bucket

new Function('exports',compile('rgwBucketLifecycleForm.ts'))(form)
const antd={Alert:'Alert',Button:'Button',Checkbox:'Checkbox',Input:{TextArea:'TextArea'},Select:'Select',Space:'Space',Form:{useWatch:()=>bucket.natural_key}}
new Function('exports','require',compile('RgwBucketLifecycleEditor.tsx'))(editor,name=>name==='./rgwBucketLifecycleForm'?form:name==='antd'?antd:{jsx,jsxs:jsx})
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
const draft={rules:[{id:'kept',status:'Disabled',selector:{kind:'Filter',and:false,prefix:'keep/',tags:[],object_size_greater_than:null,object_size_less_than:null,archive_zone:false},actions:[{type:'Transition',fields:{Days:'0',StorageClass:'OLD'}},{type:'NoncurrentVersionTransition',fields:{NoncurrentDays:'3',StorageClass:'HISTORY'}}]}]}
let changed
const show=(disabled=false)=>nodes(editor.RgwBucketLifecycleEditor({value:draft,onChange:value=>{changed=value},disabled,storageClassOptions:result.options}))
const options=show().filter(n=>n.type==='Select'&&n.props['aria-label']?.includes('库存转换类候选'))
assert.equal(options.length,2);assert.equal(changed,undefined);assert.equal(options[0].props.value,undefined)
options[0].props.onChange('COLD')
assert.equal(changed.rules[0].actions[0].fields.StorageClass,'COLD');assert.equal(changed.rules[0].actions[0].fields.Days,'0');assert.deepEqual(changed.rules[0].actions[1],draft.rules[0].actions[1]);assert.deepEqual(changed.rules[0].selector,draft.rules[0].selector)
assert.match(form.lifecycleDocument(changed),/<StorageClass>COLD<\/StorageClass>/);assert.equal(draft.rules[0].actions[0].fields.StorageClass,'OLD')
changed=undefined;options[0].props.onChange('OTHER-GROUP');assert.equal(changed,undefined)
options[1].props.onChange('LOCAL');assert.equal(changed.rules[0].actions[1].fields.StorageClass,'LOCAL')
changed=undefined;show(true).find(n=>n.props?.['aria-label']?.includes('库存转换类候选')).props.onChange('COLD');assert.equal(changed,undefined)

const ui={},refs=[],requests=[];let refIndex=0,state,cleanup,deps,pendingEffect
const react={useRef:v=>refs[refIndex++]??(refs[refIndex-1]={current:v}),useState:v=>[state??v,x=>{state=x}],useEffect:(fn,next)=>{if(JSON.stringify(next)!==JSON.stringify(deps)){deps=next;pendingEffect=()=>{cleanup?.();cleanup=fn()}}}}
new Function('exports','require',compile('RgwBucketLifecycleStorageEditor.tsx'))(ui,name=>name==='react'?react:name==='antd'?antd:name.includes('ClusterContext')?{useClusterContext:()=>({selectedClusterId:7})}:name==='./RgwBucketLifecycleEditor'?{RgwBucketLifecycleEditor:'Editor'}:name==='./rgwLifecycleStorageClasses'?{readLifecycleStorageClasses:(cluster,id,signal)=>new Promise((resolve,reject)=>requests.push({cluster,id,signal,resolve,reject}))}:{jsx,jsxs:jsx})
function render(cluster,id=bucket.natural_key,disabled=false){refIndex=0;const view=ui.RgwLifecycleStorageEditorScope({clusterId:cluster,bucketId:id,value:draft,disabled,onChange:v=>{changed=v}});const effect=pendingEffect;pendingEffect=undefined;effect?.();return view}
const tick=()=>new Promise(resolve=>setTimeout(resolve,0))
let view=render(7);assert.equal(requests.length,0)
nodes(view).find(n=>n.type==='Button').props.onClick();assert.equal(requests[0].cluster,7)
requests[0].resolve(result);await tick();view=render(7)
const oldEditor=nodes(view).find(n=>n.type==='Editor');assert.equal(oldEditor.props.storageClassOptions.length,2)
nodes(view).find(n=>n.type==='Button').props.onClick();render(7,'bmV3')
assert.equal(requests[1].signal.aborted,true);requests[1].resolve(result);await tick();assert.equal(state.data,undefined)
changed=undefined;oldEditor.props.onChange(draft);assert.equal(changed,undefined)
view=render(7,'bmV3');nodes(view).find(n=>n.type==='Button').props.onClick();render(8,'bmV3')
requests[2].resolve(result);await tick();assert.equal(state.data,undefined)
view=render(8,'bmV3');nodes(view).find(n=>n.type==='Button').props.onClick();requests[3].reject(new Error('offline'));await tick()
assert.ok(nodes(render(8,'bmV3')).some(n=>n.type==='Alert'&&n.props.message.includes('读取失败')))
assert.equal(nodes(render(8,'bmV3')).find(n=>n.type==='Editor').props.storageClassOptions,undefined)
nodes(render(8,'bmV3')).find(n=>n.type==='Button').props.onClick();cleanup();requests[4].resolve(result);await tick();assert.equal(state.data,undefined)
assert.equal(nodes(render(undefined,'bmV3')).find(n=>n.type==='Button').props.disabled,true)
assert.equal(nodes(render(8,'bmV3',true)).find(n=>n.type==='Button').props.disabled,true)
const wrapper=ui.RgwBucketLifecycleStorageEditor({value:draft});assert.equal(wrapper.props.bucketId,bucket.natural_key);assert.equal(wrapper.props.clusterId,7)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/renderControl: \(disabled\) => <RgwBucketLifecycleStorageEditor disabled=\{disabled\}/)
console.log('Lifecycle transition candidates match bucket placement, preserve rules and isolate request scopes')
