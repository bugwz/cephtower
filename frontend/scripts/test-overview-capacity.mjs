import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/overview/OverviewPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('OverviewPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const code = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['capacityPercent', 'nonNegativeQuantity', 'formatBytes', 'objectHealthSummary', 'formatCount'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
const api = readFileSync(new URL('../src/api/client.ts', import.meta.url), 'utf8')
const apiTree = ts.createSourceFile('client.ts', api, ts.ScriptTarget.Latest, true)
const numberNode = apiTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'numberValue')
const exports = {}
const numberValue = new Function('exports', `${ts.transpileModule(numberNode.getText(apiTree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText}; return numberValue`)(exports)
const { capacityPercent, formatBytes, objectHealthSummary, formatCount } = new Function('numberValue', `${ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText}; return { capacityPercent, formatBytes, objectHealthSummary, formatCount }`)(numberValue)
const healthy = { copies: 100, degraded: 0, misplaced: 0, unfound: 0 }
assert.deepEqual(objectHealthSummary(healthy), { known: true, affected: 0, healthy: 100, percent: 100 })
assert.deepEqual(objectHealthSummary({ ...healthy, degraded: 10, misplaced: 20, unfound: 5 }), { known: true, affected: 35, healthy: 65, percent: 65 })
assert.deepEqual(objectHealthSummary({ ...healthy, copies: 0 }), { known: true, affected: 0, healthy: 0, percent: undefined })
for (const field of Object.keys(healthy)) {
  for (const invalid of [undefined, null, '', -1, 1.5, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) assert.equal(objectHealthSummary({ ...healthy, [field]: invalid }).known, false)
}
assert.equal(objectHealthSummary({ ...healthy, degraded: 101 }).known, false)
assert.equal(objectHealthSummary({ ...healthy, degraded: 60, misplaced: 50 }).known, false)
for (const invalid of [undefined, '', ' ', -1, 0.5, Number.MAX_SAFE_INTEGER + 1]) assert.equal(formatCount(invalid), '—')
assert.equal(formatCount(0), '0')
assert.equal(capacityPercent({ used_bytes: 0, total_bytes: 100 }), 0)
assert.equal(capacityPercent({ used_bytes: 100, total_bytes: 100 }), 100)
assert.equal(capacityPercent({ used_bytes: '25', total_bytes: '100' }), 25)
assert.equal(capacityPercent({ used_bytes: 1e300, total_bytes: 2e300 }), 50)
for (const bad of [undefined, null, '', ' ', false, [], {}, -1, Infinity, NaN, '-1', '0x10', '1e999']) {
  assert.equal(capacityPercent({ used_bytes: bad, total_bytes: 100 }), undefined)
  assert.equal(capacityPercent({ used_bytes: 0, total_bytes: bad }), undefined)
  assert.equal(formatBytes(bad), '-')
}
for (const row of [{}, { used_bytes: 0, total_bytes: 0 }, { used_bytes: 101, total_bytes: 100 }]) assert.equal(capacityPercent(row), undefined)
assert.equal(formatBytes(0), '0 B')
assert.equal(formatBytes(1024), '1.0 KiB')
assert.ok(source.includes("usedPercent === undefined ? '未知'"))
assert.ok(source.includes('usedPercent === undefined ? <Alert'))
console.log('Overview capacity distinguishes unknown readings from valid zero usage')
