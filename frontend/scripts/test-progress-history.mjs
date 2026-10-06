import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/overview/ProgressHistory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ProgressHistory.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const code = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['progressHistoryLists', 'progressTime', 'progressSeconds', 'completedDuration', 'nativeDuration'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
const exports = {}
new Function('exports', ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const value = { events: [{ id: 'one', refs: { origin: 'rbd_support' } }], completed: [{ id: 'two', failed: true, failure_message: 'error' }] }
assert.deepEqual(exports.progressHistoryLists(value), value)
assert.deepEqual(exports.progressHistoryLists({ events: [], completed: [] }), { events: [], completed: [] })
for (const bad of [null, undefined, {}, [], { events: [], completed: null }, { events: [null], completed: [] }]) assert.equal(exports.progressHistoryLists(bad), null)
assert.equal(exports.progressTime(0), '1970-01-01T00:00:00.000Z')
for (const bad of [null, '0', Infinity, NaN, 1e99]) assert.equal(exports.progressTime(bad), '未知')
assert.ok(source.includes('expandedRowRender'))
assert.equal(exports.progressSeconds(0), '0 秒')
assert.equal(exports.progressSeconds(12.3456), '12.346 秒')
for (const bad of [null, undefined, '', '12', -1, Infinity, NaN]) assert.equal(exports.progressSeconds(bad), '未知')
assert.equal(exports.completedDuration({ started_at: 10, finished_at: 12.5 }), '2.5 秒')
assert.equal(exports.completedDuration({ started_at: 0, finished_at: 0 }), '0 秒')
for (const bad of [{}, { started_at: null, finished_at: 10 }, { started_at: '1', finished_at: 10 }, { started_at: 11, finished_at: 10 }, { started_at: Infinity, finished_at: Infinity }]) assert.equal(exports.completedDuration(bad), '未知')
assert.equal(exports.nativeDuration('00h 00m 12s'), '00h 00m 12s')
for (const bad of [null, 0, {}, '', '   ']) assert.equal(exports.nativeDuration(bad), '未知')
assert.ok(source.includes("dataIndex: 'time_remaining', render: progressSeconds"))
assert.ok(source.includes("dataIndex: 'duration', render: nativeDuration"))
console.log('Progress history keeps native metadata and validates both task lists')
