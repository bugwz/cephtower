import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/RuleAlerts.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('RuleAlerts.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['ruleAlertInstances', 'ruleAlertText'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
const exports = {}
new Function('exports', ts.transpileModule(functions, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const rows = [{ labels: { instance: 'host1' }, state: 'pending', value: '9007199254740993', activeAt: '2026-10-06T01:00:00Z' }, { labels: { instance: 'host1' }, state: 'firing', value: 'NaN' }]
assert.equal(exports.ruleAlertInstances(rows), rows)
assert.deepEqual(exports.ruleAlertInstances([]), [])
for (const bad of [undefined, null, {}, '[]', [null], [0], [[]], [rows[0], null]]) assert.equal(exports.ruleAlertInstances(bad), null)
for (const value of ['9007199254740993', 'NaN', '+Inf', '0', '', '<script>']) assert.equal(exports.ruleAlertText(value), value)
for (const value of [undefined, null, 0, false, {}]) assert.equal(exports.ruleAlertText(value), '未提供')
const page = readFileSync(new URL('../src/pages/monitoring/pages.tsx', import.meta.url), 'utf8')
assert.ok(page.includes('detailContent: row => <RuleAlerts row={row} />'))
for (const field of ['labels', 'annotations', 'state', 'activeAt', 'value']) assert.ok(source.includes(`item.${field}`))
assert.ok(!source.includes('dangerouslySetInnerHTML'))
console.log('Rule alert instances preserve native values and distinguish missing from empty snapshots')
