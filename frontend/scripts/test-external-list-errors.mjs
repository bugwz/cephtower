import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={},client={}
const compile=file=>ts.transpileModule(readFileSync(new URL(file,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
// Use the actual normalization helper so missing configuration rows are not
// mistaken for arrays or dropped by the API adapter.
new Function('exports','require','localStorage',compile('../src/api/client.ts'))(client,()=>({}),{getItem:()=>null})
let response,failure
const calls=[]
new Function('exports','require',compile('../src/api/external.ts'))(api,()=>({...client,jsonInit:(method,body)=>({method,body}),request:async(path,init)=>{calls.push({path,...init});if(failure)throw failure;return response}}))
const query=new URLSearchParams({kind:'lifecycle'})
for(const error of [Object.assign(new Error('unavailable'),{status:501,code:'capability_unavailable'}),Object.assign(new Error('unavailable'),{status:501}),Object.assign(new Error('forbidden'),{status:403,code:'forbidden'}),Object.assign(new Error('upstream'),{status:502,code:'s3_failed'}),new Error('network')]){
  failure=error
  await assert.rejects(api.readExternalList('/rgw/bucket/policy',7,{bucket_id:'encoded'},query),value=>value===error)
}
failure=undefined
response={items:[],meta:{source:'native'}}
assert.deepEqual(await api.readExternalList('/rgw/bucket/policy',7,{bucket_id:'encoded'},query),response)
const row={bucket_id:'encoded',kind:'lifecycle',configured:false,document:null,lifecycle_rules:[]}
response={items:[row]}
assert.deepEqual((await api.readExternalList('/rgw/bucket/policy',7,{bucket_id:'encoded'},query)).items,[row])
response=row
assert.deepEqual((await api.readExternalList('/rgw/bucket/policy',7,{bucket_id:'encoded'},query)).items,[row])
for(const call of calls)assert.deepEqual(call,{path:'/rgw/bucket/policy?kind=lifecycle',method:'GET',body:{cluster_id:7,bucket_id:'encoded'}})
console.log('External list failures remain failures while empty lists and absent configurations are preserved')
