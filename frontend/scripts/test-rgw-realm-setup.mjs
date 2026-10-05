import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const api = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwRealmSetup.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(api)
const action = api.rgwRealmSetupAction
const v = { name: 'realm', zonegroup: 'group', zone: 'primary', username: 'sys', zone_mode: 'normal', zonegroup_endpoints: 'https://group.example', zone_endpoints: 'https://zone.example', expected_services: 'rgw.gateway', confirm_setup: 'acknowledged' }
assert.deepEqual(action.buildBody(v, 7), { cluster_id: 7, name: 'realm', zonegroup: 'group', zone: 'primary', username: 'sys', zonegroup_endpoints: ['https://group.example'], zone_endpoints: ['https://zone.example'], expected_services: ['rgw.gateway'], confirm_setup: true })
assert.equal(action.buildBody({ ...v, zone_mode: 'archive' }, 7).tier_type, 'archive')
for (const bad of [{ confirm_setup: undefined }, { name: '-bad' }, { zone_mode: '' }, { zone_endpoints: '' }, { zone_endpoints: 'https://u:password@host' }, { zone_endpoints: 'file:///secret' }, { zone_endpoints: 'https://host/?a=b' }, { expected_services: 'rgw.a\nrgw.a' }]) assert.throws(() => action.buildBody({ ...v, ...bad }, 7))
assert.deepEqual(action.changedValues({ zone_mode: 'archive' }, v), { confirm_setup: undefined })
assert.match(action.confirmation(v), /不会部署新网关/)
assert.match(action.confirmation(v), /端口和路径保留/)
assert.match(action.confirmation(v), /转换不代表连通性验证/)
assert.match(action.fields.find(f => f.name === 'zone_endpoints').label, /主机名自动转换/)
assert.match(action.successMessage, /重启已提交/)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), /toolbarActions: \[rgwRealmImportAction, rgwRealmSetupAction\]/)
console.log('Primary realm setup validates endpoints, explicit restart scope and side effects')
