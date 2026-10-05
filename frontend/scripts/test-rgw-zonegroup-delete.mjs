import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwZonegroupDelete.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(api)
const row={id:'g',name:'group',realm_id:'r',is_default:false,is_master:false,zones:[{id:'z'}]}
assert.deepEqual(api.zonegroupDeleteInput(row),{zonegroup_id:'g',name:'group',realm_id:'r',expected_zones:['z'],confirm_delete:true})
for(const change of [{stale:true},{is_default:true},{is_master:true},{is_master:undefined},{realm_id:undefined},{zones:null},{zones:[{id:'z'},{id:'z'}]}])assert.throws(()=>api.zonegroupDeleteInput({...row,...change}))
assert.deepEqual(api.zonegroupDeleteInput({...row,realm_id:'',zones:[]}).expected_zones,[])
assert.match(api.zonegroupDeleteConfirmation(row),/可能同时发布其他待提交变更/)
assert.match(api.zonegroupDeleteConfirmation(row),/其他默认引用可能残留/)
assert.match(api.zonegroupDeleteConfirmation({...row,realm_id:''}),/无 Realm，不发布 Period/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/rgwZonegroups: \{\s*deleteAction: \{[\s\S]*?zonegroupDeleteInput\(row\)/)
console.log('Zonegroup deletion explicitly retains zones and pools and confirms publication scope')
