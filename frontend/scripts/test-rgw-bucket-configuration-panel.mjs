import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={},states=[],refs=[]
let si=0,ri=0,cleanup,selectedClusterId=7
const jsx=(type,props,key)=>({type,props,key})
const options=['policy','cors','lifecycle','encryption','tagging','versioning','object-lock','acl','replication','notification'].map(value=>({value,label:value}))
const react={useState:value=>{const i=si++;if(!(i in states))states[i]=value;return [states[i],value=>states[i]=value]},useRef:value=>refs[ri++]??(refs[ri-1]={current:value}),useEffect:fn=>{if(!cleanup)cleanup=fn()}}
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwBucketConfigurationPanel.tsx',import.meta.url),'utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS,jsx:ts.JsxEmit.ReactJSX}}).outputText
new Function('exports','require',code)(api,name=>name==='react'?react:name==='react/jsx-runtime'?{jsx,jsxs:jsx}:name==='antd'?Object.fromEntries(['Alert','Button','Select','Space'].map(name=>[name,name])):name.includes('ClusterContext')?{useClusterContext:()=>({selectedClusterId})}:name==='../ExternalListPage'?{ExternalListPage:'ExternalListPage'}:{rgwBucketConfigurationReadOptions:options})
const definition={path:'/rgw/bucket/policy',body:{kind:'policy'},filterFields:[{name:'bucket_id'}],createAction:{title:'unscoped'},updateAction:{title:'scoped edit'},extraActions:[{title:'structured edit'}],deleteAction:{title:'delete'},buildQuery:body=>new URLSearchParams({kind:body.kind})}
const row={natural_key:'encoded_tenant_bucket',id:'native-id'}
const wrap=(value=row,clusterId=7)=>api.RgwBucketConfigurationPanel({row:value,clusterId,definition})
assert.equal(wrap().props.bucketId,row.natural_key)
assert.equal(wrap().key,`7:${row.natural_key}`)
for(const invalid of [{...row,stale:true},{id:'native-id'},{natural_key:''},{natural_key:'tenant/bucket'}])assert.equal(wrap(invalid).type,'Alert')
selectedClusterId=8;assert.equal(wrap().type,'Alert');assert.equal(wrap(row,8).key,`8:${row.natural_key}`);selectedClusterId=7
assert.equal(api.RgwBucketConfigurationPanel({row,definition}).type,'Alert')
function nodes(node){return !node||typeof node!=='object'?[]:[node,...[node.props?.children].flat(Infinity).flatMap(nodes)]}
function render(){si=ri=0;return nodes(api.RgwBucketConfigurationView({bucketId:row.natural_key,definition}))}
const find=type=>render().find(node=>node.type===type)
assert.equal(find('ExternalListPage'),undefined)
assert.deepEqual(find('Select').props.options,options)
const firstRead=find('Button').props.onClick
firstRead()
let view=find('ExternalListPage')
assert.equal(view.props.embedded,true)
assert.deepEqual(view.props.definition.body,{bucket_id:row.natural_key,kind:'policy'})
assert.deepEqual(view.props.definition.filterFields,[])
assert.equal(view.props.definition.createAction,undefined)
for(const key of ['updateAction','extraActions','deleteAction'])assert.equal(view.props.definition[key],definition[key])
assert.equal(definition.createAction.title,'unscoped')
for(const {value:kind} of options){
  find('Select').props.onChange(kind)
  assert.equal(find('ExternalListPage'),undefined)
  find('Button').props.onClick();view=find('ExternalListPage')
  assert.equal(view.props.definition.body.kind,kind)
  assert.equal(view.props.definition.body.bucket_id,row.natural_key)
  assert.equal(view.props.definition.buildQuery(view.props.definition.body).get('kind'),kind)
}
find('Select').props.onChange('cors');assert.equal(find('ExternalListPage'),undefined)
firstRead();assert.equal(find('ExternalListPage'),undefined)
find('Select').props.onChange('invalid');assert.equal(find('Select').props.value,'cors')
const oldRead=find('Button').props.onClick,oldSelect=find('Select').props.onChange
cleanup();oldRead();oldSelect('policy');assert.equal(find('ExternalListPage'),undefined);assert.equal(find('Select').props.value,'cors')
console.log('Bucket detail configuration reads preserve identity, explicit selection and scoped actions')
