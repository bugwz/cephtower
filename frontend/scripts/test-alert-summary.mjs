import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/monitoring/AlertSummary.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('AlertSummary.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const code = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'alertSnapshotCounts').getText(tree)
const exports = {}
new Function('exports', ts.transpileModule(code, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
assert.equal(exports.alertSnapshotCounts(undefined), null)
const empty = { active: 0, critical: 0, warning: 0, info: 0, otherSeverity: 0, suppressed: 0, unprocessed: 0, unknownState: 0 }
assert.deepEqual(exports.alertSnapshotCounts([]), empty)
const rows = ['critical', 'warning', 'info', 'future', undefined].map(severity => ({ status: { state: 'active' }, labels: { severity } }))
rows.push({ status: { state: 'suppressed' }, labels: { severity: 'critical' } }, { status: { state: 'unprocessed' } }, {}, { status: { state: 'future' } }, { status: [] })
const before = JSON.stringify(rows)
assert.deepEqual(exports.alertSnapshotCounts(rows), { active: 5, critical: 1, warning: 1, info: 1, otherSeverity: 2, suppressed: 1, unprocessed: 1, unknownState: 3 })
assert.equal(JSON.stringify(rows), before)
const page = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('loading || error || !selectedClusterId ? undefined : data'))
assert.ok(readFileSync(new URL('../src/pages/monitoring/pages.tsx', import.meta.url), 'utf8').includes('summaryContent: rows => <AlertSummary rows={rows} />'))
console.log('Alert snapshot counts separate native states and active severity without mutating rows')
