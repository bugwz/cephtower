import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/overview/ProgressHistory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ProgressHistory.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const code = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['progressHistoryLists', 'progressTime'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
const exports = {}
new Function('exports', ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const value = { events: [{ id: 'one', refs: { origin: 'rbd_support' } }], completed: [{ id: 'two', failed: true, failure_message: 'error' }] }
assert.deepEqual(exports.progressHistoryLists(value), value)
assert.deepEqual(exports.progressHistoryLists({ events: [], completed: [] }), { events: [], completed: [] })
for (const bad of [null, undefined, {}, [], { events: [], completed: null }, { events: [null], completed: [] }]) assert.equal(exports.progressHistoryLists(bad), null)
assert.equal(exports.progressTime(0), '1970-01-01T00:00:00.000Z')
for (const bad of [null, '0', Infinity, NaN, 1e99]) assert.equal(exports.progressTime(bad), '未知')
assert.ok(source.includes('expandedRowRender'))
console.log('Progress history keeps native metadata and validates both task lists')
