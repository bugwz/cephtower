import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const load = (file, require = () => ({})) => {
  const api = {}
  new Function('exports', 'require', ts.transpileModule(readFileSync(new URL(`../src/pages/object/${file}`, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(api, require)
  return api
}
const imports = load('rgwRealmImport.ts')
const { transferRealm, validTransferPair } = load('realmTransferWorkflow.ts', () => imports)
const source = { id: 1, fsid: 'source', generation: 2, enabled: true }
const target = { id: 2, fsid: 'target', generation: 3, enabled: true }
const values = { name: 'secondary', zone_mode: 'archive', port: 8080, placement_mode: 'hosts', hosts: 'node-a\nnode-b', count: 2, confirm_import: 'acknowledged' }
const input = { source, target, realmId: 'realm-id', realmName: 'realm', values }
function harness() {
  const calls = []
  const token = { token: 'c2VjcmV0' }
  const deps = {
    current: () => true,
    clusters: async () => { calls.push('clusters'); return [source, target] },
    token: async body => { calls.push(['token', body]); return token },
    importZone: async body => { calls.push(['import', structuredClone(body)]); return { details: {} } }
  }
  return { deps, calls, token }
}
assert.equal(validTransferPair(source, target), true)
for (const bad of [{ ...target, id: 1 }, { ...target, fsid: 'SOURCE' }, { ...target, fsid: '' }, { ...target, enabled: false }, undefined]) assert.equal(validTransferPair(source, bad), false)
let h = harness()
await transferRealm(input, h.deps)
assert.deepEqual(h.calls, ['clusters', ['token', { cluster_id: 1, realm_id: 'realm-id', name: 'realm' }], 'clusters', ['import', { cluster_id: 2, name: 'secondary', realm_token: 'c2VjcmV0', port: 8080, placement: { hosts: ['node-a', 'node-b'], count: 2 }, confirm_import: true, tier_type: 'archive' }]])
assert.equal(h.token.token, '')
assert.equal(values.realm_token, undefined)
for (const bad of [{ port: 0 }, { confirm_import: undefined }, { name: '-bad' }]) {
  h = harness()
  await assert.rejects(transferRealm({ ...input, values: { ...values, ...bad } }, h.deps))
  assert.equal(h.calls.length, 0)
}
for (const changed of [{ ...target, generation: 4 }, { ...target, enabled: false }, { ...target, fsid: 'different' }]) {
  h = harness()
  let reads = 0
  h.deps.clusters = async () => ++reads === 1 ? [source, target] : [source, changed]
  await assert.rejects(transferRealm(input, h.deps))
  assert.equal(h.calls.some(c => c[0] === 'import'), false)
  assert.equal(h.token.token, '')
}
h = harness()
let active = true
h.deps.current = () => active
h.deps.token = async () => { active = false; return h.token }
await assert.rejects(transferRealm(input, h.deps))
assert.equal(h.calls.some(c => c[0] === 'import'), false)
assert.equal(h.token.token, '')
h = harness()
let submitted
h.deps.importZone = async body => { submitted = body; throw new Error('failure') }
await assert.rejects(transferRealm(input, h.deps))
assert.equal(submitted.realm_token, undefined)
assert.equal(h.token.token, '')
const ui = readFileSync(new URL('../src/pages/object/RgwRealmTransfer.tsx', import.meta.url), 'utf8')
assert.match(ui, /locked\.current/)
assert.match(ui, /sequence\.current\+\+; setAcknowledged\(false\)/)
assert.match(ui, /cache: 'no-store'/)
assert.match(ui, /不会撤销/)
assert.doesNotMatch(ui, /localStorage|sessionStorage|console\./)
assert.match(readFileSync(new URL('../src/pages/object/pages.tsx', import.meta.url), 'utf8'), /<RgwRealmTransfer/)
console.log('Existing realm transfer validates cluster identities and guards credential handoff')
