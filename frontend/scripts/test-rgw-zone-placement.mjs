import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwZonePlacement.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(api)
const row={id:'z',name:'zone',realm_id:'r',zonegroup_memberships:[{zonegroup_id:'g',realm_id:'r'}],placement_pools:[{key:'p',val:{index_pool:'i',data_extra_pool:'',storage_classes:{STANDARD:{data_pool:'d',compression_type:'none'},COLD:{data_pool:'cold',compression_type:'zstd'}}}}]}
const values={zonegroup_id:'g',placement_id:'p',storage_class:'COLD',index_pool:'new',data_pool:'cold:new',data_extra_pool:'',compression:'zstd',confirm_placement:'acknowledged'}
assert.equal(api.zonePlacementBlocked(row),undefined)
assert.deepEqual(api.zonePlacementClasses(row,'p').map(c=>c.value),['STANDARD','COLD'])
assert.equal(api.zonePlacementInput(values,row).data_extra_pool,'')
assert.equal(api.zonePlacementInput(values,row).zone_id,'z')
assert.equal(api.zonePlacementChanged({placement_id:'p'},values,row).storage_class,undefined)
assert.equal(api.zonePlacementChanged({storage_class:'COLD'},values,row).data_pool,'cold')
for(const change of [{zonegroup_id:'other'},{placement_id:'other'},{storage_class:'other'},{compression:'random'},{confirm_placement:undefined},{index_pool:''}])assert.throws(()=>api.zonePlacementInput({...values,...change},row))
assert.throws(()=>api.zonePlacementInput(values,{...row,stale:true}))
assert.match(api.zonePlacementConfirmation(values,row),/不搬迁现有数据/)
assert.match(api.zonePlacementConfirmation(values,row),/影响整个放置目标/)
assert.match(api.zonePlacementConfirmation(values,{...row,realm_id:'',zonegroup_memberships:[{zonegroup_id:'g',realm_id:''}]}),/无 Realm，不发布 Period/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/path: '\/rgw\/zone\/placement'/)
console.log('Zone placement editing preserves class selection and confirms pool migration risks')
