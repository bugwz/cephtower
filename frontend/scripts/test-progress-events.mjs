import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/overview/ProgressEvents.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('ProgressEvents.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const code = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['progressEventRows', 'eventPercent'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
const exports = {}
new Function('exports', ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.deepEqual(exports.progressEventRows({}), [])
assert.deepEqual(exports.progressEventRows({ actual: { event_id: 'spoof', message: '<script>', progress: 0 } }), [{ event_id: 'actual', message: '<script>', progress: 0 }])
for (const bad of [undefined, null, [], '', { bad: null }, { bad: [] }]) assert.equal(exports.progressEventRows(bad), null)
for (const [input, output] of [[0, 0], [1, 100], [0.125, 12.5]]) assert.equal(exports.eventPercent(input), output)
for (const bad of [undefined, null, '0.5', -1, 1.01, NaN, Infinity]) assert.equal(exports.eventPercent(bad), undefined)
assert.ok(!source.includes('dangerouslySetInnerHTML'))
const overview = readFileSync(new URL('../src/pages/overview/OverviewPage.tsx', import.meta.url), 'utf8')
assert.ok(overview.includes('<ProgressEvents value={data?.overview.progress_events} />'))
console.log('Native Ceph progress snapshots preserve IDs and distinguish unknown progress')
