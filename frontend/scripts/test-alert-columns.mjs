import assert from 'node:assert/strict'
import './test-silenced-alerts.mjs'
import './test-rule-alerts.mjs'
import './test-alert-routing.mjs'
import './test-alert-auto-refresh.mjs'
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
const listSource = readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8')
const listTree = ts.createSourceFile('ExternalListPage.tsx', listSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const buildNode = listTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'buildColumns')
const build = new Function(`${ts.transpileModule(buildNode.getText(listTree), { compilerOptions: { target: ts.ScriptTarget.ES2020, jsx: ts.JsxEmit.React } }).outputText}; return buildColumns`)()
const tableColumns = build({ columns: exports.alertColumns }, () => {}, () => {}, () => {}, false)
const stateFilter = tableColumns.find(column => column.key === 'status')
assert.deepEqual(stateFilter.defaultFilteredValue, ['active'])
assert.deepEqual(stateFilter.filters.map(option => option.value), ['active', 'suppressed', 'unprocessed'])
assert.equal(stateFilter.onFilter('active', { status: { state: 'active' } }), true)
assert.equal(stateFilter.onFilter('active', { status: { state: 'suppressed' } }), false)
for (const status of [null, [], {}, { state: 0 }, { state: 'ACTIVE' }]) assert.equal(stateFilter.onFilter('active', { status }), false)
const severityFilter = tableColumns.find(column => column.key === 'severity')
assert.equal(severityFilter.onFilter('warning', row), true)
assert.equal(severityFilter.onFilter('critical', row), false)
assert.equal(severityFilter.onFilter('warning', { annotations: { severity: 'warning' } }), false)
assert.equal(tableColumns.find(column => column.key === 'fingerprint').onFilter, undefined)
const silenceFilters = build({ columns: exports.silenceColumns }, () => {}, () => {}, () => {}, false).find(column => column.key === 'status')
assert.equal(silenceFilters.defaultFilteredValue, undefined)
assert.deepEqual(silenceFilters.filters.map(option => option.value), ['active', 'pending', 'expired'])
assert.equal(silenceFilters.onFilter('expired', { status: { state: 'expired' } }), true)
assert.ok(listSource.includes('key={JSON.stringify([selectedClusterId, definition.path])}'))
const silenceState = exports.silenceColumns.find(column => column.key === 'status')
for (const state of ['active', 'pending', 'expired', 'future-state']) assert.equal(silenceState.render({ state }, {}), state)
assert.equal(silenceState.render(undefined, {}), '未提供')
for (const key of ['updatedAt', 'comment', 'startsAt', 'endsAt', 'matchers']) assert.ok(exports.silenceColumns.some(column => column.key === key))
assert.ok(page.includes('columns: silenceColumns'))
for (const key of ['name', 'group', 'file', 'state', 'health', 'duration', 'query']) assert.ok(exports.alertRuleColumns.some(column => column.key === key))
assert.equal(exports.alertRuleColumns.find(column => column.key === 'severity').render(undefined, row), 'warning')
assert.equal(exports.alertRuleColumns.find(column => column.key === 'summary').render(undefined, row), 'warning summary')
assert.ok(page.includes("rowKeyCandidates: ['rule_key']"))
assert.ok(page.includes('columns: alertRuleColumns'))

const actions = {}
new Function('exports', ts.transpileModule(readFileSync(new URL('../src/pages/monitoring/silenceActions.ts', import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(actions)
const action = actions.silenceFromAlertAction
const range = actions.silenceTimeRange
const start = '2026-10-06T00:00:00Z'
const end = '2026-10-06T01:00:00Z'
assert.deepEqual(range(start, end), { startsAt: start, endsAt: end })
assert.doesNotThrow(() => range('2026-10-06T00:00:00.000000001Z', '2026-10-06T00:00:00.000000002Z'))
assert.doesNotThrow(() => range('2024-02-29T00:00:00Z', end))
for (const invalid of [undefined, null, 0, '', '2026-10-06', '2026-10-06T00:00:00', '2026-02-29T00:00:00Z', '2026-04-31T00:00:00Z', '2026-10-06T24:00:00Z', '2026-10-06T00:00:60Z', '2026-10-06T00:00:00+24:00']) assert.throws(() => range(invalid, end))
for (const invalidEnd of [start, '2026-10-06T08:00:00+08:00', '2026-10-05T23:59:59Z']) assert.throws(() => range(start, invalidEnd), /结束时间/)
for (const candidate of [actions.silenceCreateAction, actions.silenceRecreateAction]) assert.throws(() => candidate.buildBody({ matchers_json: '[{"name":"alertname","value":"x","isRegex":false,"isEqual":true}]', startsAt: end, endsAt: start }, 1), /结束时间/)
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

const recreate = actions.silenceRecreateAction
const old = { id: 'old-id', status: { state: 'expired' }, startsAt: '2020-01-01T00:00:00Z', endsAt: '2020-01-02T00:00:00Z', updatedAt: '2020-01-01T00:00:00Z', createdBy: 'operator', comment: '维护', matchers: [{ name: 'instance', value: 'host.*', isRegex: true, isEqual: false }] }
assert.equal(recreate.disabledWhen(old), undefined)
for (const state of ['active', 'pending', 'unknown', undefined]) assert.ok(recreate.disabledWhen({ ...old, status: { state } }))
for (const matchers of [undefined, [], [null], [{ name: 'instance', value: 'host' }]]) assert.ok(recreate.disabledWhen({ ...old, matchers }))
const copy = recreate.initialValues(old)
assert.deepEqual(JSON.parse(copy.matchers_json), old.matchers)
assert.equal(copy.createdBy, old.createdBy)
assert.equal(copy.comment, old.comment)
assert.ok(Date.parse(copy.startsAt) > Date.parse(old.endsAt))
assert.equal(Date.parse(copy.endsAt) - Date.parse(copy.startsAt), 7200000)
const recreated = recreate.buildBody({ ...copy, id: old.id, status: old.status, updatedAt: old.updatedAt }, 4, old)
assert.equal(recreated.cluster_id, 4)
assert.deepEqual(recreated.matchers, old.matchers)
for (const field of ['id', 'status', 'updatedAt']) assert.equal(field in recreated, false)
assert.equal(old.startsAt, '2020-01-01T00:00:00Z')
assert.equal(recreate.path, '/alert/silence'); assert.equal(recreate.method, 'POST')
assert.ok(recreate.confirmation().includes('不修改原已过期记录'))
assert.ok(page.includes('extraActions: [silenceRecreateAction]'))
console.log('Expired silence recreation preserves matchers and excludes old identity')

const expire = actions.silenceExpireAction
for (const state of ['active', 'pending']) assert.equal(expire.disabledWhen({ id: 'silence-a', status: { state } }), undefined)
for (const state of ['expired', 'unknown', undefined]) assert.ok(expire.disabledWhen({ id: 'silence-a', status: { state } }))
for (const id of [undefined, 0, '', 'a/b', ' a', 'a ', '.', '..', 'a\nb']) {
  assert.throws(() => actions.silenceTarget({ id }))
  assert.ok(expire.disabledWhen({ id, status: { state: 'active' } }))
}
assert.deepEqual(expire.buildBody({ id: 'silence-a' }, 7), { cluster_id: 7, silence_id: 'silence-a' })
assert.equal(expire.resourceKey({ id: 'silence-a' }), 'silence/silence-a')
assert.equal(expire.buttonLabel, '结束静默')
assert.ok(expire.confirmation().includes('不删除历史记录'))
assert.ok(page.includes('deleteAction: silenceExpireAction'))
assert.ok(readFileSync(new URL('../src/pages/ExternalListPage.tsx', import.meta.url), 'utf8').includes("definition.deleteAction.buttonLabel ?? '删除'"))
console.log('Silence expiry preserves full identity and describes native expiration')

const edit = actions.silenceUpdateAction
const editable = { ...old, status: { state: 'active' } }
assert.equal(edit.disabledWhen(editable), undefined)
assert.ok(edit.disabledWhen(old))
for (const field of ['startsAt', 'endsAt', 'updatedAt']) assert.ok(edit.disabledWhen({ ...editable, [field]: undefined }))
const editValues = edit.initialValues(editable)
assert.equal(editValues.startsAt, old.startsAt)
assert.equal(editValues.endsAt, old.endsAt)
assert.deepEqual(JSON.parse(editValues.matchers_json), old.matchers)
const edited = edit.buildBody({ ...editValues, comment: 'changed', silence_id: 'other', expected_updated_at: 'other' }, 8, editable)
assert.throws(() => edit.buildBody({ ...editValues, endsAt: editValues.startsAt }, 8, editable), /结束时间/)
assert.equal(edited.silence_id, old.id)
assert.equal(edited.expected_updated_at, old.updatedAt)
assert.equal(edited.comment, 'changed')
assert.equal(edited.cluster_id, 8)
assert.equal('id' in edited, false)
assert.equal(edit.method, 'PATCH')
assert.ok(edit.confirmation().includes('不提供原子条件更新'))
assert.ok(page.includes('updateAction: silenceUpdateAction'))
