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
  const collect = new Function('active', 'running', 'setCollecting', 'setCollectionError', 'refreshResource', 'clusterId', 'message', 'refresh', `${js}; return collect`)(active, running, value => calls.push(['busy', value]), value => calls.push(['error', value]), async body => {
    calls.push(['collect', body])
    if (scenario === 'switched') active.current = false
    if (scenario === 'failed') throw new Error('offline')
  }, 7, { success: () => calls.push(['success']) }, async () => calls.push(['read']))
  await collect()
  if (scenario === 'inactive' || scenario === 'busy') assert.deepEqual(calls, [])
  else {
    assert.deepEqual(calls.slice(0, 3), [['busy', true], ['error', ''], ['collect', { clusterId: 7, kind: 'mgr' }]])
    assert.deepEqual(calls.slice(3), scenario === 'switched' ? [] : scenario === 'failed' ? [['error', 'offline'], ['busy', false]] : [['success'], ['read'], ['busy', false]])
    assert.equal(running.current, false)
  }
}
assert.ok(source.includes('return () => { active.current = false }'))
