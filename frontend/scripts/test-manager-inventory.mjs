import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/ManagerInventory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('inventory.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'managerServices')
const exports = {}
new Function('exports', ts.transpileModule(fn.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.deepEqual(exports.managerServices({}), [])
assert.deepEqual(exports.managerServices({ dashboard: 'https://host:8443/', empty: '', unsafe: 'javascript:alert(1)' }), [{ name: 'dashboard', uri: 'https://host:8443/' }, { name: 'empty', uri: '' }, { name: 'unsafe', uri: 'javascript:alert(1)' }])
for (const value of [undefined, null, [], '', { a: null }, { a: 7 }]) assert.equal(exports.managerServices(value), null)
assert.ok(!source.includes('href='))
assert.ok(source.includes("listAllResources('/managers', clusterId)"))
assert.ok(source.includes('row.active !== true'))
assert.ok(readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8').includes('<ManagerInventory key={selectedClusterId} clusterId={selectedClusterId} />'))
console.log('Manager inventory preserves service URIs as inert text and scopes active services')
const component = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'ManagerInventory')
const collectNode = component.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'collect')
const js = ts.transpileModule(collectNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const scenario of ['inactive', 'busy', 'current', 'switched', 'failed']) {
  const active = { current: scenario !== 'inactive' }, running = { current: scenario === 'busy' }, calls = []
  const collect = new Function('active', 'running', 'setCollecting', 'setCollectionError', 'refreshResource', 'clusterId', 'message', 'refresh', 'setFailResult', `${js}; return collect`)(active, running, value => calls.push(['busy', value]), value => calls.push(['error', value]), async body => {
    calls.push(['collect', body])
    if (scenario === 'switched') active.current = false
    if (scenario === 'failed') throw new Error('offline')
  }, 7, { success: () => calls.push(['success']) }, async () => calls.push(['read']), () => {})
  await collect()
  if (scenario === 'inactive' || scenario === 'busy') assert.deepEqual(calls, [])
  else {
    assert.deepEqual(calls.slice(0, 3), [['busy', true], ['error', ''], ['collect', { clusterId: 7, kind: 'mgr' }]])
    assert.deepEqual(calls.slice(3), scenario === 'switched' ? [] : scenario === 'failed' ? [['error', 'offline'], ['busy', false]] : [['success'], ['read'], ['busy', false]])
    assert.equal(running.current, false)
  }
}
assert.ok(source.includes('return () => { active.current = false }'))
const failNode = component.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'fail')
const failJS = ts.transpileModule(failNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const scenario of ['cancel', 'switch-before', 'switch-after', 'success', 'failure', 'busy', 'invalid']) {
  const active = { current: true }, running = { current: scenario === 'busy' }, calls = []
  const fail = new Function('active', 'running', 'canFail', 'setFailing', 'setFailResult', 'Modal', 'clusterId', 'mutateResource', `${failJS}; return fail`)(active, running, () => scenario !== 'invalid', () => {}, value => calls.push(['result', value]), { confirm: options => {
    if (scenario === 'switch-before') active.current = false
    if (scenario === 'cancel') options.onCancel(); else options.onOk()
  } }, 7, async (...args) => {
    calls.push(['write', ...args])
    if (scenario === 'switch-after') active.current = false
    if (scenario === 'failure') throw new Error('uncertain')
  })
  await fail({ name: 'mgr.a', active: true, resource_version: 12 })
  const writes = calls.filter(c => c[0] === 'write')
  assert.equal(writes.length, ['success', 'failure', 'switch-after'].includes(scenario) ? 1 : 0)
  if (writes.length) assert.deepEqual(writes[0], ['write', '/manager/fail', 'POST', { cluster_id: 7, name: 'mgr.a' }, { ifMatch: '12' }])
  if (scenario.startsWith('switch')) assert.deepEqual(calls.filter(c => c[0] === 'result'), [['result', '']])
  if (scenario === 'failure') assert.match(calls.at(-1)[1], /不要盲目重试/)
  if (scenario === 'success') assert.match(calls.at(-1)[1], /不代表接管成功/)
}
