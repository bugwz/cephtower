import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/monitoring/AlertGroups.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('AlertGroups.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'groupedAlertItems')
const exports = {}
new Function('exports', ts.transpileModule(fn.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const rows = [{ status: { state: 'active' } }, { status: { state: 'suppressed' } }]
assert.equal(exports.groupedAlertItems(rows), rows)
assert.deepEqual(exports.groupedAlertItems([]), [])
for (const bad of [null, undefined, {}, [null], [[]], [1]]) assert.equal(exports.groupedAlertItems(bad), null)
assert.ok(source.includes("path: '/alert/groups'"))
assert.ok(source.includes('expandedRowRender'))
const pages = readFileSync(new URL('../src/pages/monitoring/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes("key={grouped ? 'groups' : 'instances'}"))
console.log('Native alert group view preserves individual states and rejects invalid lists')
