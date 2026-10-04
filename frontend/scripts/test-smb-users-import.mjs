import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import * as yaml from 'yaml'

const compile = file => ts.transpileModule(readFileSync(new URL(file, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.React } }).outputText
const api = {}
new Function('exports', 'require', compile('../src/pages/file/smbUsersImport.ts'))(api, name => { assert.equal(name, 'yaml'); return yaml })
const resource = { resource_type:'ceph.smb.usersgroups', users_groups_id:'staff', values:{users:[{name:'alice',password:' secret\n'}],groups:[{name:'ops'}]},linked_to_cluster:'cluster-a' }
const expected = { name:'staff',users:'[{"name":"alice","password":" secret\\n"}]',groups:'ops',clear_groups:false,linked_to_cluster:'cluster-a' }
for (const text of [JSON.stringify(resource), yaml.stringify(resource)]) {
 assert.deepEqual(api.smbUsersImport(text),expected)
 assert.deepEqual(api.smbUsersImport(text,{name:'staff'}),expected)
 assert.deepEqual(api.smbUsersImport(text,{natural_key:'staff'}),expected)
 assert.throws(()=>api.smbUsersImport(text,{name:'other'}))
 assert.throws(()=>api.smbUsersImport(text,{name:'staff',stale:true}))
}
assert.deepEqual(api.smbUsersImport(JSON.stringify({...resource,values:{users:[],groups:[]},linked_to_cluster:undefined})),{name:'staff',users:'[]',groups:'',clear_groups:true,linked_to_cluster:undefined})
const invalid = [null, [], {}, {...resource,resource_type:'ceph.smb.cluster'}, {...resource,intent:'removed'}, {...resource,extra:'secret'}, {...resource,linked_to_cluster:null}, {...resource,users_groups_id:'bad.id'}, {...resource,values:{users:[{name:'a',password:123}],groups:[]}}, {...resource,values:{users:resource.values.users,groups:[{name:' padded '}]}}, {...resource,values:{users:[...resource.values.users,...resource.values.users],groups:[]}}, {...resource,values:{users:[],groups:[{name:'ops'},{name:'ops'}]}}, {...resource,values:{users:[],groups:[{name:'a\nb'}]}}]
for (const raw of invalid) assert.throws(()=>api.smbUsersImport(JSON.stringify(raw)),error=>error.message==='用户组资源文件无效或与当前编辑身份不一致')
for (const text of ['password: [secret', yaml.stringify(resource)+'---\n'+yaml.stringify(resource), yaml.stringify(resource)+'resource_type: ceph.smb.usersgroups\n', 'x: &a [*a]', 'x: !unknown secret', ' '.repeat(1048577)]) assert.throws(()=>api.smbUsersImport(text),error=>!error.message.includes('secret'))

const component = {}
let refs=[], states=[], effects=[], cursor=0, pending=[]
const hooks={
 useRef(initial){const i=cursor++;return refs[i]??=( {current:initial} )},
 useState(initial){const i=cursor++;if(!(i in states))states[i]=initial;return [states[i],v=>{states[i]=v}]},
 useEffect(fn,deps){const i=cursor++;if(!effects[i]||deps.some((v,n)=>v!==effects[i].deps[n]))pending.push(()=>{effects[i]?.cleanup?.();effects[i]={deps,cleanup:fn()}})}
}
const react={createElement:(type,props,...children)=>({type,props:{...props,children}})}
new Function('exports','require','React',compile('../src/pages/ResourceFormImport.tsx'))(component,name=>name==='react'?hooks:{Alert:'Alert',Space:'Space'},react)
let values={}, applied=[], props={parse:api.smbUsersImport,active:true,disabled:false,scope:1,values:()=>values,apply:next=>{applied.push(next);values={...values,...next}}}
function render(){cursor=0;const tree=component.ResourceFormImport(props);pending.splice(0).forEach(fn=>fn());return tree}
function nodes(node){return !node||typeof node!=='object'?[]:[node,...(node.props?.children??[]).flat(Infinity).flatMap(nodes)]}
function select(file){return nodes(render()).find(n=>n.type==='input').props.onChange({target:{files:[file],value:'file'}})}
const file={name:'resource.yaml',size:100,text:async()=>yaml.stringify(resource)}
await select(file);assert.deepEqual(applied,[expected])
await select({...file,text:async()=>{throw new Error('secret')}});assert.equal(applied.length,1)
assert.ok(!JSON.stringify(render()).includes('secret'))
await select({...file,size:1048577,text:()=>{throw new Error('must not read')}});assert.equal(applied.length,1)
let resolve
const waiting=()=>({...file,text:()=>new Promise(done=>{resolve=done})})
let promise=select(waiting());values={...values,groups:'changed'};resolve(yaml.stringify(resource));await promise;assert.equal(applied.length,1)
promise=select(waiting());props={...props,scope:2};render();resolve(yaml.stringify(resource));await promise;assert.equal(applied.length,1)
promise=select(waiting());props={...props,active:false};render();resolve(yaml.stringify(resource));await promise;assert.equal(applied.length,1)
props={...props,active:true};render()
promise=select(waiting());props={...props,disabled:true};render();resolve(yaml.stringify(resource));await promise;assert.equal(applied.length,1)
props={...props,disabled:false};render()
promise=select(waiting());props={...props,disabled:true};render();props={...props,disabled:false};render();resolve(yaml.stringify(resource));await promise;assert.equal(applied.length,1)
promise=select(waiting());const oldResolve=resolve;await select(file);oldResolve(yaml.stringify({...resource,users_groups_id:'old'}));await promise;assert.equal(applied.length,2)
promise=select(waiting());effects.forEach(effect=>effect?.cleanup?.());resolve(yaml.stringify(resource));await promise;assert.equal(applied.length,2)
console.log('SMB JSON/YAML import preserves secrets, identity and asynchronous form scope')
