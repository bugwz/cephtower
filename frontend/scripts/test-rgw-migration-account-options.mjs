import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const helpers = {}
const calls = []
const id = 'RGW12345678901234567'
const another = 'RGW00000000000000000'
let inventory = { items: [], stale: false }
let fail = false
new Function('exports', 'require', ts.transpileModule(readFileSync(new URL('../src/pages/object/rgwMigrationAccountOptions.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(helpers, () => ({ listAllResources: async (...args) => {
  calls.push(args)
  if (fail) throw new Error('offline')
  return inventory
} }))
const options = helpers.rgwMigrationAccountOptions
assert.deepEqual(options(inventory, ''), [])
const account = { account_id: id, tenant: '', account_name: 'team', stale: false }
inventory.items = [account, { account_id: another, tenant: 'other', stale: false }]
assert.deepEqual(options(inventory, ''), [{ value: id, label: `team · ${id}` }])
assert.equal(options(inventory, 'other')[0].value, another)
assert.deepEqual(options(inventory, 'missing'), [])
inventory.stale = true
assert.match(options(inventory, '')[0].label, /库存过期/)
inventory.stale = false
delete account.stale
assert.match(options(inventory, '')[0].label, /新鲜度未知/)
inventory.items.push({ ...account, tenant: 'other' })
assert.deepEqual(options(inventory, ''), [])
inventory.items = [{ account_id: id, tenant: null }, { account_id: 'bad', tenant: '' }, { tenant: '' }]
assert.deepEqual(options(inventory, ''), [])
inventory.items = [account]
assert.deepEqual(await helpers.loadRgwMigrationAccountOptions(7, { tenant: '' }), options(inventory, ''))
assert.deepEqual(calls.pop(), ['/rgw/accounts', 7])
await helpers.loadRgwMigrationAccountOptions(8, { tenant: 'other' })
assert.deepEqual(calls.pop(), ['/rgw/accounts', 8])
const count = calls.length
for (const cluster of [undefined, 0, -1, 1.5]) await assert.rejects(helpers.loadRgwMigrationAccountOptions(cluster, { tenant: '' }))
for (const row of [undefined, {}, { tenant: null }]) await assert.rejects(helpers.loadRgwMigrationAccountOptions(7, row))
assert.equal(calls.length, count)
fail = true
await assert.rejects(helpers.loadRgwMigrationAccountOptions(7, { tenant: '' }), /offline/)
console.log('RGW migration account options preserve tenant, cluster and inventory freshness')
