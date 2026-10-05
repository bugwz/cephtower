import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api={}
new Function('exports',ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRealmDelete.ts',import.meta.url),'utf8'),{compilerOptions:{module:ts.ModuleKind.CommonJS}}).outputText)(api)
const row={id:'realm-id',name:'realm',current_period:'period',is_default:false}
assert.deepEqual(api.realmDeleteInput(row),{realm_id:'realm-id',name:'realm',expected_current_period:'period',confirm_delete:true})
for(const change of [{stale:true},{is_default:true},{is_default:undefined},{id:''},{name:'-bad'},{current_period:null},{name:'bad\nname'}])assert.throws(()=>api.realmDeleteInput({...row,...change}))
assert.match(api.realmDeleteConfirmation(row),/不删除 Zonegroup、Zone、Period/)
assert.match(api.realmDeleteConfirmation(row),/不自动重试或回滚/)
assert.match(api.realmDeleteConfirmation(row),/realm-id/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx',import.meta.url),'utf8'),/multisite: \{\s*deleteAction: \{[\s\S]*?disabledWhen: realmDeleteBlocked,[\s\S]*?realmDeleteInput\(row\)/)
console.log('Realm deletion requires known non-default identity and explicit impact confirmation')
