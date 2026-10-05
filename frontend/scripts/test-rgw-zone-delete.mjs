import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwZoneDelete.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(api)
const row={id:'z',name:'zone',realm_id:'r',is_default:false}
assert.deepEqual(api.zoneDeleteInput(row),{zone_id:'z',name:'zone',realm_id:'r',confirm_delete:true})
for(const change of [{stale:true},{is_default:true},{is_default:undefined},{realm_id:undefined},{id:'-bad'},{name:''}])assert.throws(()=>api.zoneDeleteInput({...row,...change}))
assert.equal(api.zoneDeleteInput({...row,realm_id:''}).realm_id,'')
assert.match(api.zoneDeleteConfirmation(row),/保留所有池/)
assert.match(api.zoneDeleteConfirmation(row),/其他默认引用可能残留/)
assert.match(api.zoneDeleteConfirmation(row),/可能同时发布其他待提交变更/)
assert.match(api.zoneDeleteConfirmation({...row,realm_id:''}),/无 Realm，不发布 Period/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/rgwZones: \{\s*deleteAction: \{[\s\S]*?zoneDeleteInput\(row\)/)
console.log('Zone deletion confirms identity, retained pools and publication scope')
