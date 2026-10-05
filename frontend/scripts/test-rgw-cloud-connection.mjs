import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
const group={},restore={},data={}
new Function('exports',compile('rgwZonegroupStorageClass.ts'))(group)
new Function('exports','require',compile('rgwCloudRestore.ts'))(restore,()=>group)
new Function('exports','require',compile('rgwCloudConnection.ts'))(data,()=>restore)
const tier={storage_class:'COLD',tier_type:'cloud-s3',s3:{endpoint:'https://old.example',endpoint_redacted:true,access_key:'[REDACTED]',secret:'[REDACTED]'}}
const row={id:'g',name:'group',realm_id:'r',default_placement:'p',placement_targets:[{name:'p',storage_classes:['STANDARD','COLD'],tier_targets:[{key:'COLD',val:tier}]}]}
const values={placement_id:'p',storage_class:'COLD',endpoint:'https://new.example:9443/path',access_key:'new-private-access',secret:'new-private-secret',credentials_saved:'acknowledged',confirm_connection:'acknowledged'}
const input=data.cloudConnectionInput(values,row)
assert.equal(input.endpoint,values.endpoint);assert.equal(input.secret,values.secret)
assert.doesNotMatch(JSON.stringify(input),/old.example|REDACTED|endpoint_redacted/)
const confirmation=data.cloudConnectionConfirmation(values,row)
assert.doesNotMatch(confirmation,/new-private-access|new-private-secret/)
assert.match(confirmation,/进程参数/);assert.match(confirmation,/不测试远端连通性/);assert.match(confirmation,/Realm r/)
assert.match(data.cloudConnectionConfirmation({...values,endpoint:'http://host'},row),/不提供 TLS/)
for(const endpoint of ['', 'ftp://host','https://@host','https://user:pass@host','https://host?','https://host#','https://host?token=x','https://host\\path','https://host\n','https://host:0','https://host:65536','https://host:','https:///path','https://host:bad'])assert.throws(()=>data.cloudConnectionInput({...values,endpoint},row),endpoint)
for(const endpoint of ['http://host/path','https://host','https://[::1]:443/path','https://host/path%20name'])assert.equal(data.cloudConnectionEndpoint(endpoint),endpoint)
for(const key of ['access_key','secret'])for(const value of [undefined,'',' ','[REDACTED]','x\n','x'.repeat(4097)])assert.throws(()=>data.cloudConnectionInput({...values,[key]:value},row))
for(const changed of [{credentials_saved:undefined},{confirm_connection:undefined},{storage_class:'STANDARD'}])assert.throws(()=>data.cloudConnectionInput({...values,...changed},row))
assert.throws(()=>data.cloudConnectionInput(values,{...row,stale:true}))
assert.throws(()=>data.cloudConnectionInput(values,{...row,realm_id:''}))
for(const changed of [{placement_id:'p'},{storage_class:'COLD'}]) {
 const reset=data.cloudConnectionChanged(changed,values,row)
 for(const key of ['endpoint','access_key','secret','credentials_saved','confirm_connection'])assert.equal(reset[key],undefined)
 assert.doesNotMatch(JSON.stringify(reset),/old.example|REDACTED/)
}
for(const key of ['endpoint','access_key','secret'])assert.deepEqual(data.cloudConnectionChanged({[key]:'changed'}),{credentials_saved:undefined,confirm_connection:undefined})
assert.deepEqual(data.cloudConnectionChanged({credentials_saved:'acknowledged'}),{})
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/path:'\/rgw\/zonegroup\/placement\/connection',method:'PATCH'/)
assert.match(source,/changedValues:cloudConnectionChanged,confirmation:cloudConnectionConfirmation/)
assert.match(source,/name:'access_key',label:'远端已配置的 Access Key（重新输入）',type:'password'/)
assert.match(source,/name:'secret',label:'远端已配置的 Secret（重新输入）',type:'password'/)
console.log('Cloud connection form uses fresh credentials, safe endpoints and secret-free confirmations')
