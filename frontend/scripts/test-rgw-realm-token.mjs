import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
let refs=[],states=[],effects=[],cursor=0,pending=[]
const hooks={useRef(v){const i=cursor++;return refs[i]??={current:v}},useState(v){const i=cursor++;if(!(i in states))states[i]=v;return [states[i],v=>{states[i]=v}]},useEffect(fn,deps){const i=cursor++;if(!effects[i]||deps.some((v,n)=>v!==effects[i].deps[n]))pending.push(()=>{effects[i]?.cleanup?.();effects[i]={deps,cleanup:fn()}})}}
const calls=[]
let resolve,reject
const client={jsonInit:(method,body,options)=>({method,body,...options}),request:(path,init)=>{calls.push({path,...init});return new Promise((yes,no)=>{resolve=yes;reject=no})}}
const api={}
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwRealmToken.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.React}}).outputText
const ui={Alert:'Alert',Button:'Button',Card:'Card',Checkbox:'Checkbox',Input:{Password:'Password'},Space:'Space'}
new Function('exports','require','React',code)(api,name=>name==='react'?hooks:name==='antd'?ui:client,{createElement:(type,props,...children)=>({type,props:{...props,children}})})
let props={row:{id:'id',name:'realm'},clusterId:7}
function render(){cursor=0;const tree=api.RgwRealmToken(props);pending.splice(0).forEach(fn=>fn());return tree}
function nodes(node){return !node||typeof node!=='object'?[]:[node,...(node.props?.children??[]).flat(Infinity).flatMap(nodes)]}
const find=(type,label)=>nodes(render()).find(n=>n.type===type&&(!label||n.props.children.includes(label)))
const settle=async()=>{await Promise.resolve();await Promise.resolve()}
render()
find('Button','读取敏感 Token').props.onClick();assert.equal(calls.length,0)
find('Checkbox').props.onChange({target:{checked:true}})
find('Button','读取敏感 Token').props.onClick()
assert.deepEqual(calls[0],{path:'/rgw/realm/token',method:'POST',body:{cluster_id:7,realm_id:'id',name:'realm'},cache:'no-store',suppressErrorNotification:true})
resolve({token:'c2VjcmV0'});await settle()
assert.equal(find('Password').props.value,'c2VjcmV0')
find('Button','清空').props.onClick();assert.equal(find('Password'),undefined)
find('Checkbox').props.onChange({target:{checked:true}})
find('Button','读取敏感 Token').props.onClick();const oldResolve=resolve
props={...props,clusterId:8};render();oldResolve({token:'c2VjcmV0'});await settle();assert.equal(find('Password'),undefined)
find('Checkbox').props.onChange({target:{checked:true}})
find('Button','读取敏感 Token').props.onClick();reject(new Error('secret-server-error'));await settle();assert.ok(!JSON.stringify(render()).includes('secret-server-error'))
find('Button','读取敏感 Token').props.onClick();find('Button','清空').props.onClick();resolve({token:'c2VjcmV0'});await settle();assert.equal(find('Password'),undefined)
find('Checkbox').props.onChange({target:{checked:true}})
find('Button','读取敏感 Token').props.onClick();effects.forEach(effect=>effect?.cleanup?.());resolve({token:'c2VjcmV0'});await settle();assert.equal(find('Password'),undefined)
props={...props,row:{...props.row,stale:true}};render();assert.equal(find('Button','读取敏感 Token').props.disabled,true)
const pages=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(pages,/multisite: \{[\s\S]*?detailContent:.*<RgwRealmToken/)
console.log('Realm token disclosure is explicit, masked and invalidated on scope changes')
