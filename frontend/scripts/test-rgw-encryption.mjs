import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwEncryptionPage.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
const ui={},states=[],refs=[],requests=[],jsx=(type,props)=>({type,props})
let stateIndex=0,refIndex=0,effectDeps,cleanup,pendingEffect
const react={useRef:v=>refs[refIndex++]??(refs[refIndex-1]={current:v}),useState:v=>{const i=stateIndex++;if(!(i in states))states[i]=v;return [states[i],next=>{states[i]=next}]},useEffect:(fn,deps)=>{if(JSON.stringify(deps)!==JSON.stringify(effectDeps)){effectDeps=deps;pendingEffect=()=>{cleanup?.();cleanup=fn()}}}}
const client={jsonInit:(method,body,opts)=>({method,body,...opts}),request:(path,init)=>new Promise((resolve,reject)=>requests.push({path,...init,resolve,reject}))}
new Function('exports','require',code)(ui,name=>name==='react'?react:name==='antd'?Object.fromEntries(['Alert','Button','Card','Descriptions','Input','Select','Space','Table'].map(n=>[n,n])):name.includes('api/client')?client:name.includes('ClusterContext')?{useClusterContext:()=>({selectedClusterId:7})}:{jsx,jsxs:jsx})
const names=['addr','auth','prefix','secret_engine','namespace','token_file','ssl_cacert','ssl_clientcert','ssl_clientkey','verify_ssl']
const data={entity:'client.rgw.a',encryption_type:'kms',provider:'vault',backend:'barbican',observed_at:'now',fields:names.map(name=>({name,option:`rgw_crypt_vault_${name}`,value:name==='verify_ssl'?'false':'',redacted:false}))}
assert.deepEqual(ui.encryptionConfiguration(data,data.entity,'kms/vault'),data)
assert.throws(()=>ui.encryptionConfiguration(data,data.entity,'s3/kmip'))
for(const bad of [null,{}, {...data,entity:'client.rgw.b'},{...data,fields:data.fields.slice(1)},{...data,fields:[...data.fields.slice(1),data.fields[1]]},{...data,fields:[{...data.fields[0],option:'other'},...data.fields.slice(1)]}])assert.throws(()=>ui.encryptionConfiguration(bad,data.entity,'kms/vault'))
const kmip={...data,provider:'kmip',fields:['addr','username','password','client_cert','client_key','ca_path','kms_key_template','s3_key_template'].map(name=>({name,option:`rgw_crypt_kmip_${name}`,value:name==='password'?'[REDACTED]':'',redacted:name==='password'}))}
assert.equal(ui.encryptionConfiguration(kmip,data.entity,'kms/kmip').fields.length,8)
assert.throws(()=>ui.encryptionConfiguration({...kmip,fields:kmip.fields.map(f=>f.name==='password'?{...f,value:'secret'}:f)},data.entity,'kms/kmip'))
const s3={...data,encryption_type:'s3',fields:[...data.fields.map(f=>({...f,option:f.option.replace('rgw_crypt_','rgw_crypt_sse_s3_')})),{name:'key_template',option:'rgw_crypt_sse_s3_key_template',value:'%bucket_id',redacted:false}]}
assert.equal(ui.encryptionConfiguration(s3,data.entity,'s3/vault').fields.length,11)
function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
function render(id=7){stateIndex=refIndex=0;const tree=ui.RgwEncryptionView({clusterId:id});const effect=pendingEffect;pendingEffect=undefined;effect?.();return nodes(tree)}
const tick=()=>new Promise(resolve=>setTimeout(resolve,0))
let tree=render();assert.equal(requests.length,0);assert.equal(tree.find(n=>n.type==='Button').props.disabled,true)
tree.find(n=>n.type==='Input').props.onChange({target:{value:data.entity}});tree=render();tree=render()
const oldButton=tree.find(n=>n.type==='Button');oldButton.props.onClick()
assert.equal(requests[0].path,'/rgw/encryption/configuration');assert.equal(requests[0].cache,'no-store');assert.equal(requests[0].method,'GET')
assert.deepEqual(requests[0].body,{cluster_id:7,entity:data.entity,encryption_type:'kms',provider:'vault'})
requests[0].resolve(data);await tick();tree=render()
const table=tree.find(n=>n.type==='Table');assert.equal(table.props.dataSource.length,10)
assert.equal(table.props.columns[2].render(null,data.fields[9]).props.children,'"false"')
assert.ok(tree.some(n=>n.type==='Alert'&&n.props.type==='warning'))
tree.find(n=>n.type==='Button').props.onClick();tree.find(n=>n.type==='Select').props.onChange('kms/kmip');render()
assert.equal(requests[1].signal.aborted,true);oldButton.props.onClick();assert.equal(requests.length,2)
requests[1].resolve(data);await tick();assert.equal(render().some(n=>n.type==='Table'),false)
render().find(n=>n.type==='Button').props.onClick();render(8)
assert.equal(requests[2].signal.aborted,true);requests[2].resolve(kmip);await tick();assert.equal(render(8).some(n=>n.type==='Table'),false)
render(8).find(n=>n.type==='Button').props.onClick();requests[3].reject(new Error('offline'));await tick()
assert.ok(render(8).some(n=>n.type==='Alert'&&n.props.type==='error'))
const unmountedButton=render(8).find(n=>n.type==='Button');unmountedButton.props.onClick();cleanup();assert.equal(requests[4].signal.aborted,true);unmountedButton.props.onClick();assert.equal(requests.length,5);requests[4].resolve(kmip);await tick();assert.equal(states[2].data,undefined)
assert.match(readFileSync(new URL('../src/navigation.ts',import.meta.url),'utf8'),/key: 'rgwEncryption'.*permission: 'storage'/)
console.log('RGW encryption configuration verifies profile fields, protected secrets and isolated reads')
