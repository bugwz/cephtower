import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const modules={}
for(const name of ['rgwZonegroupStorageClass','rgwCloudRestore','rgwCloudConnection','rgwCloudTarget','rgwCloudACL','rgwCloudCreate']) {
 const js=ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}.ts`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
 modules[name]={};new Function('exports','require',js)(modules[name],path=>modules[path.slice(2)])
}
const data=modules.rgwCloudCreate
const row={id:'g',name:'group',realm_id:'r',default_placement:'',placement_targets:[{name:'p',storage_classes:['STANDARD'],tier_targets:[]}]}
const values={placement_id:'p',storage_class:'COLD',tier_type:'cloud-s3',endpoint:'https://cloud.example',access_key:'create-private-access',secret:'create-private-secret',region_mode:'clear',target_path_mode:'set',target_path:'new,{path}',target_storage_class_mode:'clear',host_style:'path',multipart_sync_threshold:0,multipart_min_part_size:16,acls_json:'[]',retain_head_object:'false',allow_read_through:'true',read_through_restore_days:0,restore_storage_class:'STANDARD',credentials_saved:'acknowledged',confirm_create:'acknowledged'}
assert.equal(data.cloudCreateBlocked(row),undefined)
assert.deepEqual(data.cloudCreateTargets(row),[{value:'p',label:'p'}])
const input=data.cloudCreateInput(values,row)
assert.equal(input.secret,values.secret);assert.equal(input.target.region,'');assert.equal(input.target.target_path,'new,{path}');assert.equal(input.target.multipart_sync_threshold,0)
assert.equal(input.retain_head_object,false);assert.equal(input.allow_read_through,true);assert.equal(input.read_through_restore_days,0);assert.deepEqual(input.acls,[])
assert.equal(input.confirm_create,true);assert.equal(input.credentials_saved,true)
const message=data.cloudCreateConfirmation(values,row)
assert.doesNotMatch(message,/create-private-access|create-private-secret/)
assert.match(message,/进程参数/);assert.match(message,/不迁移对象/);assert.match(message,/初始化为默认/);assert.match(message,/Realm r/)
const glacier={...values,tier_type:'cloud-s3-glacier',glacier_restore_days:0,glacier_restore_tier_type:'Expedited'}
assert.equal(data.cloudCreateInput(glacier,row).glacier_restore_days,0)
assert.match(data.cloudCreateConfirmation(glacier,row),/等级=Expedited/)
assert.match(data.cloudCreateConfirmation({...values,endpoint:'http://host'},row),/不提供 TLS/)
for(const changed of [{storage_class:'STANDARD'},{storage_class:'-bad'},{tier_type:'unknown'},{endpoint:'https://u:p@host'},{secret:'[REDACTED]'},{confirm_create:undefined},{credentials_saved:undefined},{region_mode:undefined},{target_path_mode:'set',target_path:''},{host_style:'future'},{multipart_min_part_size:-1},{read_through_restore_days:1.5},{restore_storage_class:'MISSING'},{acls_json:'{}'},{acls_json:'[{"source_id":"s","dest_id":"","type":"bad"}]'},{glacier_restore_days:0}])assert.throws(()=>data.cloudCreateInput({...values,...changed},row))
for(const changed of [{glacier_restore_days:undefined},{glacier_restore_tier_type:'Bulk'}])assert.throws(()=>data.cloudCreateInput({...glacier,...changed},row))
for(const bad of [{...row,stale:true},{...row,realm_id:''},{...row,placement_targets:[{...row.placement_targets[0],storage_classes:['STANDARD','COLD']}]},{...row,placement_targets:[{...row.placement_targets[0],tier_targets:[{key:'COLD',val:{storage_class:'COLD'}}]}]},{...row,placement_targets:[{...row.placement_targets[0],tier_targets:null}]}])assert.throws(()=>data.cloudCreateInput(values,bad))
const reset=data.cloudCreateChanged({placement_id:'q'})
for(const key of ['storage_class','restore_storage_class','endpoint','access_key','secret','credentials_saved','confirm_create'])assert.equal(reset[key],undefined)
const typeReset=data.cloudCreateChanged({tier_type:'cloud-s3'})
assert.equal(typeReset.glacier_restore_days,undefined);assert.equal(typeReset.glacier_restore_tier_type,undefined)
assert.equal(data.cloudCreateChanged({target_path:'new'}).confirm_create,undefined)
assert.deepEqual(data.cloudCreateChanged({credentials_saved:'acknowledged'}),{})
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/path:'\/rgw\/zonegroup\/storage\/class\/cloud',method:'POST'/)
assert.match(source,/disabledWhen:cloudCreateBlocked,changedValues:cloudCreateChanged,confirmation:cloudCreateConfirmation/)
assert.match(source,/name:'access_key',label:'远端已配置的 Access Key',type:'password'/)
assert.match(source,/name:'secret',label:'远端已配置的 Secret',type:'password'/)
assert.match(source,/cloudCreateInput\(values,row\)/)
console.log('Cloud class creation requires complete explicit settings, fresh credentials and an unused class')
