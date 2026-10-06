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
