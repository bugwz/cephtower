import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
function load(file, require = () => ({})) {
  const exports = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(`../src/pages/object/${file}`, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports, require)
  return exports
}
const setup = load('rgwRealmSetup.ts')
const action = load('rgwRealmMigration.ts', () => setup).rgwRealmMigrationAction
const values = { name: 'realm', zonegroup: 'group', zone: 'primary', username: 'sys', zone_mode: 'normal', zonegroup_endpoints: 'https://group.example', zone_endpoints: 'https://zone.example', expected_services: 'rgw.gateway', confirm_setup: 'acknowledged', expected_zonegroup_id: 'g', expected_zone_id: 'z' }
const body = action.buildBody(values, 7)
assert.equal(body.confirm_migration, true)
assert.equal(body.expected_zone_id, 'z')
assert.equal(body.expected_zonegroup_id, 'g')
assert.equal(body.confirm_setup, true)
assert.equal(action.path, '/rgw/realm/migrate')
assert.equal(action.buildBody({ ...values, zone_mode: 'archive' }, 7).tier_type, 'archive')
for (const bad of [{ expected_zone_id: '' }, { expected_zonegroup_id: 'g z' }, { zone: 'default' }, { zonegroup: 'default' }, { confirm_setup: undefined }]) assert.throws(() => action.buildBody({ ...values, ...bad }, 7))
assert.deepEqual(action.changedValues({ expected_zone_id: 'other' }, values), { confirm_setup: undefined })
assert.match(action.confirmation(values), /不是对象搬迁/)
assert.match(action.confirmation(values), /不会自动改写/)
console.log('Default topology migration requires source identities and explicit maintenance confirmation')
