import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/pages.tsx', import.meta.url), 'utf8')
const file = ts.createSourceFile('pages.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const node = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'deviceUsage')
const code = ts.transpileModule(node.getText(file), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const usage = new Function('deviceReasonValues', 'isUsedDeviceReason', 'readableDeviceReason', `${code}; return deviceUsage`)(v => v ?? [], v => v === 'LVM', v => v)
assert.equal(usage({ available: true }).state, 'available')
assert.equal(usage({ available: false }).state, 'unavailable')
assert.equal(usage({ available: false, rejected_reasons: ['LVM'] }).state, 'used')
for (const available of [undefined, null, 'false', 0, {}]) {
  const result = usage({ available, rejected_reasons: ['LVM'] })
  assert.equal(result.state, 'unknown')
  assert.deepEqual(result.notes, ['LVM'])
}
console.log('Device availability retains unknown native states and rejection diagnostics')
const page = file.statements.find(n => ts.isFunctionDeclaration(n) && n.name.text === 'DeviceManagementPage')
const declarations = page.body.statements.filter(ts.isVariableStatement).flatMap(n => [...n.declarationList.declarations])
const loader = declarations.find(n => n.name.getText(file) === 'loader').initializer.arguments[0]
const loaderCode = ts.transpileModule(`const load = ${loader.getText(file)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const calls = []
const load = new Function('selectedClusterId', 'deviceTableFilters', 'listAllResources', `${loaderCode}; return load`)(42, { filters: { hostname: ['node1'] } }, async (...args) => { calls.push(args); return { items: [] } })
await load()
assert.deepEqual(calls, [['/devices', 42, { filters: { hostname: ['node1'] } }]])
const rows = declarations.find(n => n.name.getText(file) === 'deviceRows').initializer.arguments[0]
const rowsCode = ts.transpileModule(`const select = ${rows.getText(file)}`, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scope of ['all', 'available', 'used', 'unavailable', 'unknown']) {
  const items = ['available', 'used', 'unavailable', 'unknown'].map(usage_state => ({ usage_state }))
  const select = new Function('data', 'normalizeDeviceRow', 'availabilityScope', `${rowsCode}; return select`)({ items }, row => row, scope)
  assert.deepEqual(select(), scope === 'all' ? items : items.filter(row => row.usage_state === scope))
}
assert.ok(!page.getText(file).includes("filterKey: 'usage_state'"))
assert.ok(!page.getText(file).includes("filterKey: 'size_display'"))
