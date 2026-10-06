import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/monitoring/AlertGroups.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('AlertGroups.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['groupedAlertItems', 'groupAlertFacets'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
const exports = {}
new Function('exports', ts.transpileModule(functions, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 } }).outputText)(exports)
const rows = [{ status: { state: 'active' } }, { status: { state: 'suppressed' } }]
assert.equal(exports.groupedAlertItems(rows), rows)
const facets = exports.groupAlertFacets
assert.deepEqual(facets([...rows, rows[0], {}], 'status', 'state'), { values: ['active', 'suppressed'], summary: 'active: 2；suppressed: 1；未提供: 1' })
assert.deepEqual(facets([], 'status', 'state'), { values: [], summary: '无实例' })
assert.deepEqual(facets(null, 'status', 'state'), { values: [], summary: '未知' })
assert.deepEqual(facets([{ labels: { severity: 'future' } }, { labels: [] }, { labels: { severity: 0 } }], 'labels', 'severity'), { values: ['future'], summary: 'future: 1；未提供: 2' })
assert.deepEqual(exports.groupedAlertItems([]), [])
for (const bad of [null, undefined, {}, [null], [[]], [1]]) assert.equal(exports.groupedAlertItems(bad), null)
assert.ok(source.includes("path: '/alert/groups'"))
assert.ok(source.includes('expandedRowRender'))
const pages = readFileSync(new URL('../src/pages/monitoring/pages.tsx', import.meta.url), 'utf8')
assert.ok(pages.includes("key={grouped ? 'groups' : 'instances'}"))
console.log('Native alert group view preserves individual states and rejects invalid lists')
