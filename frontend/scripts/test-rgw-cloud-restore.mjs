import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
const group={},data={}
new Function('exports',compile('rgwZonegroupStorageClass.ts'))(group)
new Function('exports','require',compile('rgwCloudRestore.ts'))(data,()=>group)
const tier={storage_class:'COLD',tier_type:'cloud-s3',retain_head_object:false,allow_read_through:true,read_through_restore_days:0,restore_storage_class:'STANDARD',s3:{secret:'private-secret'}}
const row={id:'g',name:'group',realm_id:'r',default_placement:'p',placement_targets:[{name:'p',storage_classes:['STANDARD','COLD','LOCAL'],tier_targets:[{key:'COLD',val:tier}]}]}
const values={placement_id:'p',storage_class:'COLD',retain_head_object:'false',allow_read_through:'true',read_through_restore_days:0,restore_storage_class:'STANDARD',confirm_restore:'acknowledged'}
assert.equal(data.cloudRestoreBlocked(row),undefined)
assert.deepEqual(data.cloudRestoreTargets(row),[{value:'p',label:'p'}])
assert.deepEqual(data.cloudRestoreClasses(row,'p'),[{value:'COLD',label:'COLD (cloud-s3)'}])
assert.deepEqual(data.cloudRestoreLocalClasses(row,'p').map(x=>x.value),['STANDARD','LOCAL'])
const input=data.cloudRestoreInput(values,row)
assert.equal(input.retain_head_object,false);assert.equal(input.allow_read_through,true);assert.equal(input.read_through_restore_days,0)
assert.equal(input.tier_type,'cloud-s3');assert.equal(input.confirm_restore,true)
assert.doesNotMatch(JSON.stringify(input),/private|secret/)
assert.match(data.cloudRestoreConfirmation(values,row),/Realm r.*Period/)
assert.match(data.cloudRestoreConfirmation(values,row),/不直接恢复/)
const initial=data.cloudRestoreChanged({storage_class:'COLD'},values,row)
assert.equal(initial.retain_head_object,'false');assert.equal(initial.read_through_restore_days,0);assert.equal(initial.confirm_restore,undefined)
assert.equal(data.cloudRestoreChanged({placement_id:'other'},values,row).storage_class,undefined)
assert.deepEqual(data.cloudRestoreChanged({allow_read_through:'false'},values,row),{confirm_restore:undefined})
assert.deepEqual(data.cloudRestoreChanged({confirm_restore:'acknowledged'},values,row),{})
for(const changed of [{realm_id:''},{stale:true},{placement_targets:[{...row.placement_targets[0],tier_targets:null}]},{placement_targets:[{...row.placement_targets[0],tier_targets:[]}]}])assert.ok(data.cloudRestoreBlocked({...row,...changed}))
for(const changed of [{storage_class:'STANDARD'},{restore_storage_class:'COLD'},{restore_storage_class:'missing'},{retain_head_object:false},{allow_read_through:undefined},{read_through_restore_days:-1},{read_through_restore_days:1.5},{read_through_restore_days:Number.MAX_SAFE_INTEGER+1},{read_through_restore_days:'1'},{confirm_restore:undefined}])assert.throws(()=>data.cloudRestoreInput({...values,...changed},row))
const glacier={...row,placement_targets:[{...row.placement_targets[0],tier_targets:[{key:'COLD',val:{...tier,tier_type:'cloud-s3-glacier'}}]}]}
assert.equal(data.cloudRestoreInput(values,glacier).tier_type,'cloud-s3-glacier')
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/path:'\/rgw\/zonegroup\/placement\/restore',method:'PATCH'/)
assert.match(source,/changedValues:cloudRestoreChanged/)
assert.match(source,/cloudRestoreInput\(values,row\)/)
console.log('Cloud restore form preserves native false and zero, validates local restore classes and confirms Period scope')
