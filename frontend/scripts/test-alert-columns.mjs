import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/alertColumns.ts', import.meta.url), 'utf8')
const exports = {}
new Function('exports', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const row = { labels: { alertname: 'CephHealth', severity: 'warning' }, annotations: { summary: 'warning summary' }, status: { state: 'suppressed' } }
for (const [key, expected] of [['alertname', 'CephHealth'], ['summary', 'warning summary'], ['severity', 'warning'], ['status', 'suppressed']]) {
  const column = exports.alertColumns.find(column => column.key === key)
  assert.equal(column.render(row[key], row), expected)
  assert.equal(column.render(undefined, {}), '未提供')
}
for (const value of [null, undefined, [], { state: 0 }, { state: false }]) assert.equal(exports.alertField(value, 'state'), '未提供')
assert.equal(exports.alertField({ state: 'new-state' }, 'state'), 'new-state')
assert.equal(exports.alertField({ summary: '' }, 'summary'), '')
for (const key of ['fingerprint', 'startsAt', 'endsAt', 'generatorURL']) assert.ok(exports.alertColumns.some(column => column.key === key))
const page = readFileSync(new URL('../src/pages/monitoring/pages.tsx', import.meta.url), 'utf8')
assert.ok(page.includes("rowKeyCandidates: ['fingerprint']"))
assert.ok(page.includes('columns: alertColumns'))
console.log('Alert list columns read native nested state and labels')
for (const key of ['name', 'group', 'file', 'state', 'health', 'duration', 'query']) assert.ok(exports.alertRuleColumns.some(column => column.key === key))
assert.equal(exports.alertRuleColumns.find(column => column.key === 'severity').render(undefined, row), 'warning')
assert.equal(exports.alertRuleColumns.find(column => column.key === 'summary').render(undefined, row), 'warning summary')
assert.ok(page.includes("rowKeyCandidates: ['rule_key']"))
assert.ok(page.includes('columns: alertRuleColumns'))

const actions = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/monitoring/silenceActions.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(actions)
const action = actions.silenceFromAlertAction
const name = 'Ceph[Warning].* "中文"'
const input = { labels: { alertname: name, instance: 'host1' } }
const initial = action.initialValues(input)
assert.deepEqual(JSON.parse(initial.matchers_json), [{ name: 'alertname', value: name, isRegex: false, isEqual: true }])
assert.equal(Date.parse(initial.endsAt) - Date.parse(initial.startsAt), 7200000)
assert.equal(action.disabledWhen(input), undefined)
for (const labels of [undefined, null, [], {}, { alertname: '' }, { alertname: ' ' }, { alertname: 5 }]) assert.ok(action.disabledWhen({ labels }))
assert.equal(action.path, '/alert/silence'); assert.equal(action.method, 'POST')
assert.equal(action.buildBody(initial, 9, input).cluster_id, 9)
assert.equal(action.buildBody(initial, 9, input).matchers[0].value, name)
const changed = { ...initial, matchers_json: JSON.stringify([{ name: 'instance', value: 'host1', isRegex: false, isEqual: true }]) }
assert.equal(action.buildBody(changed, 9, input).matchers[0].name, 'instance')
for (const bad of ['null', '{}', '[]', '[null]', '[{"name":"a","value":"b"}]', '[{"name":"a","value":"b","isRegex":"false","isEqual":true}]']) assert.throws(() => actions.silenceMatchers(bad))
assert.ok(action.confirmation().includes('所有同名告警'))
assert.ok(page.includes('extraActions: [silenceFromAlertAction]'))
assert.ok(page.includes('createAction: silenceCreateAction'))
console.log('Create silence from alert prefills literal matchers and requires review')
