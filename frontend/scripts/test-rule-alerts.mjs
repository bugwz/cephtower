import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/RuleAlerts.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('RuleAlerts.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['ruleAlertInstances', 'ruleAlertText', 'ruleEvaluation', 'evaluationSeconds'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
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
for (const row of [{ health: 'err', alerts: [] }, { health: 'ok', lastError: '<script>failed</script>' }, { lastError: 'query failed' }]) assert.equal(exports.ruleEvaluation(row).type, 'error')
for (const health of [undefined, null, '', 'unknown', 'future', false, 0]) assert.equal(exports.ruleEvaluation({ health }).type, 'warning')
assert.equal(exports.ruleEvaluation({ health: 'ok' }).type, 'info')
assert.equal(exports.ruleEvaluation({ health: 'ok', lastError: '' }).type, 'info')
assert.equal(exports.evaluationSeconds(0), '0 秒')
assert.equal(exports.evaluationSeconds(0.00001), '0.00001 秒')
for (const value of [undefined, null, '0', -1, NaN, Infinity]) assert.equal(exports.evaluationSeconds(value), '未提供')
for (const field of ['health', 'lastEvaluation', 'evaluationTime', 'lastError']) assert.ok(source.includes(`row.${field}`))
console.log('Rule alert instances preserve native values and distinguish missing from empty snapshots')
