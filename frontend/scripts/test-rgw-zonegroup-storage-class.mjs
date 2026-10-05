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
