import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
const group={},restore={},data={}
new Function('exports',compile('rgwZonegroupStorageClass.ts'))(group)
new Function('exports','require',compile('rgwCloudRestore.ts'))(restore,()=>group)
new Function('exports','require',compile('rgwCloudTarget.ts'))(data,()=>restore)
const old={region:'region',host_style:'path',target_path:'old/path',target_storage_class:'',multipart_sync_threshold:32,multipart_min_part_size:8}
const tier={storage_class:'COLD',tier_type:'cloud-s3',s3:{...old,endpoint:'https://cloud.example',secret:'private-secret'}}
const row={id:'g',name:'group',realm_id:'r',default_placement:'p',placement_targets:[{name:'p',storage_classes:['STANDARD','COLD'],tier_targets:[{key:'COLD',val:tier}]}]}
const selection={placement_id:'p',storage_class:'COLD'}
const initial=data.cloudTargetChanged({storage_class:'COLD'},selection,row)
assert.equal(initial.region_mode,'set');assert.equal(initial.target_storage_class_mode,'clear');assert.equal(initial.multipart_min_part_size,8)
const values={...selection,...initial,region_mode:'clear',target_path:'new,{path}',host_style:'virtual',multipart_sync_threshold:0,confirm_target:'acknowledged'}
const input=data.cloudTargetInput(values,row)
assert.deepEqual(input.expected_target,old)
assert.equal(input.target.region,'');assert.equal(input.target.target_storage_class,'');assert.equal(input.target.target_path,'new,{path}');assert.equal(input.target.multipart_sync_threshold,0)
assert.doesNotMatch(JSON.stringify(input),/private-secret|endpoint/)
assert.match(data.cloudTargetConfirmation(values,row),/不迁移对象/)
assert.match(data.cloudTargetConfirmation(values,row),/Realm r/)
assert.equal(data.cloudTargetChanged({placement_id:'new'},values,row).storage_class,undefined)
assert.equal(data.cloudTargetChanged({region_mode:'clear'},values,row).confirm_target,undefined)
assert.throws(()=>data.cloudTargetInput({...selection,...initial,confirm_target:'acknowledged'},row))
for(const changed of [{region_mode:undefined},{region_mode:'set',region:''},{confirm_target:undefined},{host_style:'other'},{multipart_sync_threshold:-1},{multipart_min_part_size:1.5},{multipart_min_part_size:Number.MAX_SAFE_INTEGER+1}])assert.throws(()=>data.cloudTargetInput({...values,...changed},row))
for(const change of [{region:null},{multipart_min_part_size:undefined},{host_style:'future'}]) {
 const bad={...row,placement_targets:[{...row.placement_targets[0],tier_targets:[{key:'COLD',val:{...tier,s3:{...tier.s3,...change}}}]}]}
 assert.throws(()=>data.cloudTargetInput(values,bad))
 assert.equal(data.cloudTargetChanged({storage_class:'COLD'},selection,bad).region_mode,undefined)
}
assert.throws(()=>data.cloudTargetInput(values,{...row,stale:true}))
assert.throws(()=>data.cloudTargetValues({...old,endpoint:'other'}))
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/path:'\/rgw\/zonegroup\/placement\/target',method:'PATCH'/)
assert.match(source,/changedValues:cloudTargetChanged,confirmation:cloudTargetConfirmation/)
assert.match(source,/cloudTargetTextFields.flatMap/)
console.log('Cloud target form preserves explicit clears and exact sizes without resubmitting endpoints or credentials')
