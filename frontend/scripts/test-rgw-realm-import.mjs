import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRealmImport.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(api)
const action = api.rgwRealmImportAction
const v = { name: 'secondary', realm_token: 'c2VjcmV0', port: 80, placement_mode: 'default', zone_mode: 'normal', confirm_import: 'acknowledged' }
assert.deepEqual(action.buildBody(v, 7), { cluster_id: 7, name: 'secondary', realm_token: v.realm_token, port: 80, placement: {}, confirm_import: true })
assert.deepEqual(action.buildBody({ ...v, placement_mode: 'hosts', hosts: 'host-a\nhost-b', label: 'ignored', count: 2 }, 7).placement, { hosts: ['host-a', 'host-b'], count: 2 })
assert.deepEqual(action.buildBody({ ...v, placement_mode: 'label', label: 'rgw', hosts: 'ignored' }, 7).placement, { label: 'rgw' })
for (const bad of [{ confirm_import: undefined }, { port: 0 }, { port: 1.5 }, { port: 65536 }, { port: '80' }, { count: 0 }, { count: 1.5 }, { placement_mode: 'hosts', hosts: 'a\na' }, { placement_mode: 'label', label: '' }, { realm_token: 'secret\n' }, { name: '-invalid' }]) assert.throws(() => action.buildBody({ ...v, ...bad }, 7))
assert.deepEqual(action.changedValues({ name: 'next' }, v), { confirm_import: undefined })
assert.deepEqual(action.changedValues({ confirm_import: 'acknowledged' }, v), {})
assert.equal(action.fields.find(f => f.name === 'realm_token').type, 'password')
assert.match(action.confirmation(v), /非事务/)
assert.equal(action.buildBody({ ...v, zone_mode: 'archive' }, 7).tier_type, 'archive')
assert.throws(() => action.buildBody({ ...v, zone_mode: 'unknown' }, 7))
assert.match(action.confirmation({ ...v, zone_mode: 'archive' }), /只从当前主 Zone 同步/)
assert.match(action.confirmation(v), /特权进程/)
assert.equal(action.path, '/rgw/realm/import')
assert.match(action.successMessage, /目标系统用户及 Token 密钥一致/)
assert.match(action.successMessage, /不代表 HTTP\/TLS 可用或全部同步完成/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), /toolbarActions: \[rgwRealmImportAction, rgwRealmSetupAction, rgwRealmMigrationAction\]/)
console.log('Realm token import form validates placement, credentials and explicit side effects')
