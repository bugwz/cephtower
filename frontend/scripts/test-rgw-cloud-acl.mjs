import assert from 'node:assert/strict'
import {readFileSync} from 'node:fs'
import ts from 'typescript'
const compile=name=>ts.transpileModule(readFileSync(new URL(`../src/pages/object/${name}`,import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS,target:ts.ScriptTarget.ES2022}}).outputText
const group={},restore={},acl={}
new Function('exports',compile('rgwZonegroupStorageClass.ts'))(group)
new Function('exports','require',compile('rgwCloudRestore.ts'))(restore,()=>group)
new Function('exports','require',compile('rgwCloudACL.ts'))(acl,()=>restore)
const old=[{source_id:'s',dest_id:'old',type:'id'}],desired=[{source_id:'s',dest_id:'new',type:'email'},{source_id:'true',dest_id:'',type:'uri'}]
const tier={storage_class:'COLD',tier_type:'cloud-s3',s3:{secret:'private-secret',acl_mappings:old.map(val=>({key:val.source_id,val}))}}
const row={id:'g',name:'group',realm_id:'r',default_placement:'p',placement_targets:[{name:'p',storage_classes:['STANDARD','COLD'],tier_targets:[{key:'COLD',val:tier}]}]}
const values={placement_id:'p',storage_class:'COLD',acls_json:JSON.stringify(desired),confirm_acl:'acknowledged',confirm_clear:'no'}
const input=acl.cloudACLInput(values,row)
assert.deepEqual(input.acls,desired);assert.deepEqual(input.expected_acls,old);assert.equal(input.confirm_clear,false)
assert.doesNotMatch(JSON.stringify(input),/private-secret/)
assert.match(acl.cloudACLConfirmation(values,row),/移除 0 个旧来源/)
const clear={...values,acls_json:'[]',confirm_clear:'acknowledged'}
assert.match(acl.cloudACLConfirmation(clear,row),/清空全部/)
assert.equal(acl.cloudACLInput(clear,row).confirm_clear,true)
assert.throws(()=>acl.cloudACLInput({...clear,confirm_clear:'no'},row))
assert.throws(()=>acl.cloudACLInput({...values,acls_json:JSON.stringify(old)},row))
assert.throws(()=>acl.cloudACLInput({...values,confirm_acl:undefined},row))
assert.throws(()=>acl.cloudACLInput(values,{...row,stale:true}))
for(const value of [null,{},[...old,...old],[{source_id:'',dest_id:'x',type:'id'}],[{source_id:'s',dest_id:'x',type:'ID'}],[{source_id:'s',dest_id:'x',type:'id',secret:'extra'}],[{source_id:'a,b',dest_id:'x',type:'id'}],[{source_id:'s',dest_id:'{x}',type:'id'}]])assert.throws(()=>acl.cloudACLList(value))
assert.deepEqual(acl.cloudACLList([{source_id:'001',dest_id:'quote"slash\\=',type:'id'}]),[{source_id:'001',dest_id:'quote"slash\\=',type:'id'}])
const initial=acl.cloudACLChanged({storage_class:'COLD'},values,row)
assert.deepEqual(JSON.parse(initial.acls_json),old);assert.equal(initial.confirm_acl,undefined)
assert.equal(acl.cloudACLChanged({placement_id:'new'},values,row).storage_class,undefined)
assert.equal(acl.cloudACLChanged({acls_json:'[]'},values,row).confirm_clear,undefined)
const bad={...row,placement_targets:[{...row.placement_targets[0],tier_targets:[{key:'COLD',val:{...tier,s3:{acl_mappings:[{key:'wrong',val:old[0]}]}}}]}]}
assert.throws(()=>acl.cloudACLInput(values,bad))
assert.equal(acl.cloudACLChanged({storage_class:'COLD'},values,bad).acls_json,undefined)
const source=readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8')
assert.match(source,/path:'\/rgw\/zonegroup\/placement\/acl',method:'PATCH'/)
assert.match(source,/changedValues:cloudACLChanged,confirmation:cloudACLConfirmation/)
assert.match(source,/cloudACLInput\(values,row\)/)
console.log('Cloud ACL form verifies snapshots, source identity, explicit clearing and secret-free payloads')
