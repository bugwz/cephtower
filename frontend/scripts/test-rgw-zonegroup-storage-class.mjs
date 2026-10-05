import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwZonegroupStorageClass.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(api)
const row={id:'g',name:'group',realm_id:'r',default_placement:'p',placement_targets:[{name:'p',storage_classes:['STANDARD'],tags:['restricted']}]}
const values={placement_id:'p',storage_class:'COLD',confirm_create:'acknowledged'}
assert.equal(api.groupStorageClassBlocked(row),undefined)
assert.deepEqual(api.groupStorageClassInput(values,row),{zonegroup_id:'g',name:'group',realm_id:'r',placement_id:'p',storage_class:'COLD',expected_default_placement:'p',confirm_create:true})
for(const change of [{storage_class:'STANDARD'},{placement_id:'missing'},{storage_class:''},{confirm_create:undefined}])assert.throws(()=>api.groupStorageClassInput({...values,...change},row))
for(const change of [{stale:true},{default_placement:undefined},{placement_targets:null},{placement_targets:[{name:'p',storage_classes:['STANDARD','STANDARD']}]}])assert.throws(()=>api.groupStorageClassInput(values,{...row,...change}))
assert.match(api.groupStorageClassConfirmation(values,row),/不配置各 Zone 的池/)
assert.match(api.groupStorageClassConfirmation(values,row),/保留默认放置目标 p/)
assert.match(api.groupStorageClassConfirmation(values,{...row,default_placement:''}),/STANDARD 设为默认/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/path: '\/rgw\/zonegroup\/storage\/class'/)
console.log('Zonegroup class declaration guards scope, duplicates and default initialization')
const create={placement_id:'new',tags_json:'["z","a"]',confirm_create:'acknowledged'}
assert.deepEqual(api.groupPlacementCreateInput(create,row).tags,['z','a'])
assert.equal('storage_class' in api.groupPlacementCreateInput(create,row),false)
assert.equal(api.groupPlacementCreateBlocked({...row,placement_targets:[]}),undefined)
assert.deepEqual(api.groupPlacementCreateInput({...create,tags_json:'[]'},{...row,placement_targets:[]}).tags,[])
for(const change of [{placement_id:'p'},{placement_id:'p/CLASS'},{tags_json:'["a,b"]'},{tags_json:'["a","a"]'},{tags_json:'null'},{tags_json:'[""]'},{confirm_create:undefined}])assert.throws(()=>api.groupPlacementCreateInput({...create,...change},row))
assert.match(api.groupPlacementCreateConfirmation(create,row),/保留当前默认目标 p/)
assert.match(api.groupPlacementCreateConfirmation(create,{...row,default_placement:''}),/新目标设为默认/)
assert.match(api.groupPlacementCreateConfirmation(create,row),/不配置任何 Zone 池/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/path: '\/rgw\/zonegroup\/placement'/)
console.log('New group placements preserve explicit tags, STANDARD initialization and default scope')
const defaultValues={placement_id:'p',storage_class:'STANDARD',confirm_default:'acknowledged'}
assert.deepEqual(api.groupPlacementDefaultClasses(row,'p'),[{value:'STANDARD',label:'STANDARD'}])
assert.equal(api.groupPlacementDefaultInput(defaultValues,row).expected_default_placement,'p')
for(const change of [{placement_id:'missing'},{storage_class:'COLD'},{confirm_default:undefined}])assert.throws(()=>api.groupPlacementDefaultInput({...defaultValues,...change},row))
assert.throws(()=>api.groupPlacementDefaultInput(defaultValues,{...row,stale:true}))
assert.match(api.groupPlacementDefaultConfirmation(defaultValues,row),/不迁移已有桶或对象/)
assert.match(api.groupPlacementDefaultConfirmation(defaultValues,row),/不自动发布 Period/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/optionsDependencies:\['placement_id'\],optionsLoader:async .*groupPlacementDefaultClasses/)
const tagValues={placement_id:'p',storage_class:'STANDARD',tags_json:'[]',confirm_tags:'acknowledged'}
assert.deepEqual(api.groupPlacementTagsInput(tagValues,row).tags,[])
assert.deepEqual(api.groupPlacementTagsInput(tagValues,row).expected_tags,['restricted'])
assert.deepEqual(api.groupPlacementTagsChanged({placement_id:'p'},tagValues,row),{storage_class:undefined,tags_json:'["restricted"]',confirm_tags:undefined})
assert.deepEqual(api.groupPlacementTagsChanged({tags_json:'[]'},tagValues,row),{confirm_tags:undefined})
for(const change of [{placement_id:'missing'},{storage_class:'missing'},{tags_json:'null'},{tags_json:'["a,b"]'},{tags_json:'["a","a"]'},{confirm_tags:undefined}])assert.throws(()=>api.groupPlacementTagsInput({...tagValues,...change},row))
assert.match(api.groupPlacementTagsConfirmation(tagValues,row),/清空标签可能放宽/)
assert.match(api.groupPlacementTagsConfirmation(tagValues,{...row,default_placement:''}),/STANDARD 设为默认/)
assert.match(api.groupPlacementTagsConfirmation(tagValues,row),/不自动发布 Period/)
const deleteRow={...row,default_placement:'p/COLD',placement_targets:[{name:'p',storage_classes:['COLD','STANDARD'],tags:[]}]}
const deleteValues={placement_id:'p',storage_class:'COLD',confirm_delete:'acknowledged'}
assert.equal(api.groupStorageClassDeleteInput(deleteValues,deleteRow).expected_default_placement,'p/COLD')
assert.throws(()=>api.groupStorageClassDeleteInput({...deleteValues,storage_class:'STANDARD'},row))
for(const change of [{placement_id:'missing'},{storage_class:'missing'},{confirm_delete:undefined}])assert.throws(()=>api.groupStorageClassDeleteInput({...deleteValues,...change},deleteRow))
assert.match(api.groupStorageClassDeleteConfirmation(deleteValues,deleteRow),/云分层配置/)
assert.match(api.groupStorageClassDeleteConfirmation(deleteValues,deleteRow),/回退为此目标的 STANDARD/)
assert.match(api.groupStorageClassDeleteConfirmation(deleteValues,deleteRow),/可能同时发布其他待提交/)
assert.match(api.groupStorageClassDeleteConfirmation(deleteValues,{...deleteRow,realm_id:''}),/无 Realm，不发布 Period/)
