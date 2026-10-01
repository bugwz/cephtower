import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const page = tree.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'MgrManagementPage')
const toggle = page.body.statements.find((node) => ts.isFunctionDeclaration(node) && node.name.text === 'toggleModule')
const code = ts.transpileModule(toggle.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const row = { name: 'prometheus', stale: false, enabled: false, always_on: false, can_run: true }
function setup(overrides = {}) {
  const events = [], scope = { clusterId: 1 }, moduleScope = { current: scope }, moduleRunning = { current: false }
  const env = {
    selectedClusterId: 1, scope, moduleScope, moduleRunning, loading: false, error: '',
    textValue: (value, fallback) => typeof value === 'string' ? value : fallback,
    setPendingModule: (name) => events.push(['pending', name]),
    setMgrModuleEnabled: async (...args) => events.push(['write', ...args]),
    refreshResource: async (args) => events.push(['collect', args]),
    refresh: async () => events.push(['read']),
    message: { success: () => events.push(['success']), warning: () => events.push(['warning']) },
    ...overrides
  }
  return { env, events, run: new Function(...Object.keys(env), `${code}; return toggleModule`)(...Object.values(env)) }
}
for (const bad of [{ ...row, stale: true }, { ...row, stale: undefined }, { ...row, always_on: true }, { ...row, enabled: undefined }, { ...row, can_run: undefined }, { ...row, can_run: false }, { ...row, name: '' }, { ...row, enabled: true }]) {
  const test = setup(); await test.run(bad, true); assert.deepEqual(test.events, [])
}
for (const overrides of [{ loading: true }, { error: 'failed' }, { selectedClusterId: null }]) {
  const test = setup(overrides); await test.run(row, true); assert.deepEqual(test.events, [])
}
{
  const test = setup(); await test.run(row, true)
  assert.deepEqual(test.events.map((event) => event[0]), ['pending', 'write', 'success', 'collect', 'read', 'pending'])
  assert.deepEqual(test.events[1], ['write', 1, 'prometheus', true])
}
{
  const test = setup({ refreshResource: async () => { throw new Error('collection failed') } })
  await test.run(row, true)
  assert.deepEqual(test.events.map((event) => event[0]), ['pending', 'write', 'success', 'warning', 'read', 'pending'])
}
{
  const test = setup({ setMgrModuleEnabled: async () => { throw new Error('write failed') } })
  await assert.rejects(test.run(row, true), /write failed/)
  assert.deepEqual(test.events.map((event) => event[0]), ['pending', 'pending'])
  assert.equal(test.env.moduleRunning.current, false)
}
for (const timing of ['before', 'write', 'collect']) {
  let release
  const test = setup(timing === 'before' ? {} : { [timing === 'write' ? 'setMgrModuleEnabled' : 'refreshResource']: () => new Promise((resolve) => { release = resolve }) })
  if (timing === 'before') test.env.moduleScope.current = { clusterId: 2 }
  const pending = test.run(row, true)
  await Promise.resolve()
  if (timing !== 'before') {
    await test.run(row, true)
    assert.equal(test.events.filter((event) => event[0] === 'pending').length, 1, 'synchronous lock must block overlap')
    test.env.moduleScope.current = { clusterId: 2 }; release()
  }
  await pending
  assert.ok(!test.events.some((event) => event[0] === 'read'))
  if (timing !== 'collect') assert.ok(!test.events.some((event) => event[0] === 'success'))
}
console.log('Manager module action scope and outcome checks passed')
