import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const code=ts.transpileModule(readFileSync(new URL('../src/pages/object/RgwEncryptionEditor.tsx',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022,jsx:ts.JsxEmit.ReactJSX}}).outputText
function fixture(provider='vault',encryption_type='kms') {
 const states=[],refs=[],calls=[],exports={};let si=0,ri=0,cleanup,current=true
 const jsx=(type,props)=>({type,props})
 const react={useState:v=>{const i=si++;if(!(i in states))states[i]=v;return[states[i],v=>states[i]=v]},useRef:v=>refs[ri++]??(refs[ri-1]={current:v}),useEffect:fn=>{if(!cleanup)cleanup=fn()}}
 new Function('exports','require',code)(exports,name=>name==='react'?react:name==='antd'?{...Object.fromEntries(['Alert','Button','Checkbox','Select','Space'].map(n=>[n,n])),Input:Object.assign(()=>{}, {Password:'Password'})}:name.includes('api/resource')?{mutateResource:(...args)=>new Promise((resolve,reject)=>calls.push({args,resolve,reject}))}:{jsx,jsxs:jsx})
 const names=provider==='kmip'?['addr','username','password','client_cert','client_key','ca_path','kms_key_template','s3_key_template']:['addr','auth','prefix','secret_engine','namespace','token_file','ssl_cacert','ssl_clientcert','ssl_clientkey','verify_ssl',...(encryption_type==='s3'?['key_template']:[])]
 const configuration={entity:'client.rgw.a',provider,encryption_type,backend:'barbican',fields:names.map(name=>({name,value:'[REDACTED]',redacted:true}))}
 function nodes(n){return Array.isArray(n)?n.flatMap(nodes):n&&typeof n==='object'?[n,...nodes(n.props?.children)]:[]}
 function render(){si=ri=0;return nodes(exports.RgwEncryptionEditor({configuration,clusterId:7,labels:{},isCurrent:()=>current}))}
 const field=name=>render().find(n=>n.props?.['aria-label']===name)
 const checks=()=>render().filter(n=>n.type==='Checkbox'&&!n.props['aria-label'])
 const button=()=>render().find(n=>n.type==='Button')
 return {render,field,checks,button,calls,states,leave:()=>{current=false},unmount:()=>cleanup()}
}
const tick=()=>new Promise(r=>setTimeout(r,0))
for(const [provider,type] of [['vault','kms'],['vault','s3'],['kmip','kms']]) {
 const f=fixture(provider,type)
 assert.equal(f.button().props.disabled,true)
 const name=provider==='kmip'?'password':'verify_ssl'
 f.field('修改 '+name).props.onChange({target:{checked:true}})
 if(name==='password')assert.equal(f.field(name).props.value,'')
 const obsolete=f.button()
 f.checks()[0].props.onChange({target:{checked:true}})
 if(name==='password')f.checks()[1].props.onChange({target:{checked:true}})
 assert.equal(f.button().props.disabled,false)
 f.field(name).props.onChange(name==='password'?{target:{value:'new-private-password'}}:'false')
 obsolete.props.onClick();assert.equal(f.calls.length,0)
 assert.equal(f.button().props.disabled,true)
 f.checks()[0].props.onChange({target:{checked:true}})
 if(name==='password')f.checks()[1].props.onChange({target:{checked:true}})
 const submit=f.button();submit.props.onClick();submit.props.onClick()
 assert.equal(f.calls.length,1)
 const [path,method,body]=f.calls[0].args
 assert.equal(path,'/rgw/encryption/configuration');assert.equal(method,'PATCH')
 assert.deepEqual(body,{cluster_id:7,entity:'client.rgw.a',encryption_type:type,provider,expected_backend:'barbican',values:{[name]:name==='password'?'new-private-password':false},confirm_disruption:true,confirm_credentials_saved:name==='password'})
 assert.deepEqual(f.states[0],{})
 f.calls[0].resolve({});await tick();assert.equal(f.states[3],'done')
}
const stale=fixture();stale.field('修改 namespace').props.onChange({target:{checked:true}});stale.checks()[0].props.onChange({target:{checked:true}})
const staleButton=stale.button();stale.leave();staleButton.props.onClick();assert.equal(stale.calls.length,0)
const revoked=fixture();revoked.field('修改 namespace').props.onChange({target:{checked:true}});revoked.checks()[0].props.onChange({target:{checked:true}});const revokedButton=revoked.button();revoked.checks()[0].props.onChange({target:{checked:false}});revokedButton.props.onClick();assert.equal(revoked.calls.length,0)
const failed=fixture();failed.field('修改 namespace').props.onChange({target:{checked:true}});failed.checks()[0].props.onChange({target:{checked:true}});failed.button().props.onClick();assert.deepEqual(failed.calls[0].args[2].values,{namespace:''});failed.calls[0].reject(new Error('private response'));await tick();assert.equal(failed.states[3],'error');assert.ok(!JSON.stringify(failed.render()).includes('private response'))
const unmounted=fixture();unmounted.field('修改 namespace').props.onChange({target:{checked:true}});unmounted.checks()[0].props.onChange({target:{checked:true}});unmounted.button().props.onClick();unmounted.unmount();unmounted.calls[0].resolve({});await tick();assert.equal(unmounted.states[3],'running')
console.log('RGW encryption editor preserves explicit patches, confirmations and stale-submit isolation')
