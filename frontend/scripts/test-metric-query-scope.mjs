import assert from 'node:assert/strict'
import './test-metric-trend.mjs'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/MetricPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('MetricPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const helpers = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['metricSamples', 'normalizeSeries', 'readRecord'].includes(node.name.text))
const sampleExports = {}
new Function('exports', ts.transpileModule(helpers.map(node => node.getText(tree)).join('\n') + '\nexports.normalizeSeries = normalizeSeries', { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(sampleExports)
const samples = sampleExports.metricSamples({ values: [[0, '9007199254740993'], [1.125, '0'], [2, 'NaN'], [3, '+Inf'], [4, '-Inf'], [5, '1.2e-3']] })
assert.equal(samples.length, 6)
assert.equal(samples[0].utc, '1970-01-01T00:00:00.000Z')
assert.equal(samples[0].value, '9007199254740993')
assert.equal(samples[1].utc, '1970-01-01T00:00:01.125Z')
assert.equal(samples[1].status, '有效值')
for (const sample of samples.slice(2, 5)) assert.equal(sample.status, '非有限值（不代表零）')
assert.equal(samples[5].status, '有效值')
for (const value of [null, [], [1], ['1', '0'], [1, 0], [1, ''], [1, 'garbage'], [1e30, '1'], [1, '1', 'extra']]) {
  assert.equal(sampleExports.metricSamples({ values: [value] })[0].status, '格式异常')
}
assert.equal(sampleExports.metricSamples({ value: [0, '-2'] })[0].value, '-2')
assert.deepEqual(sampleExports.metricSamples({ values: [], value: [0, '1'] }), [])
assert.deepEqual(sampleExports.metricSamples({}), [])
const normalized = sampleExports.normalizeSeries([{ metric: { job: 'ceph' }, values: [[1, '1'], [2]] }])[0]
assert.equal(normalized.latest_value, '未提供或格式异常', 'missing value must not display its timestamp as a measurement')
assert.equal(normalized.points, 2)
assert.equal(normalized.samples.length, 2)
assert.ok(source.includes('dataSource={row.samples}'))
const wrapper = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MetricPage')
const ui = {}
let clusterId
new Function('exports', 'require', 'useClusterContext', 'MetricContent', ts.transpileModule(wrapper.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(ui, () => ({ jsx: (type, props, key) => ({ type, props, key }) }), () => ({ selectedClusterId: clusterId }), 'MetricContent')
const scopes = [undefined, 1, 2].map(id => { clusterId = id; return ui.MetricPage() })
assert.equal(new Set(scopes.map(item => item.key)).size, 3)
assert.equal(scopes[2].props.selectedClusterId, 2)
const content = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MetricContent')
const submit = content.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'submit')
const code = ts.transpileModule(submit.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['late-success', 'late-failure', 'unmount', 'failure']) {
  const calls = [], updates = []
  const active = { current: true }, pending = { current: null }
  const query = (id, input, init) => new Promise((resolve, reject) => calls.push({ id, input, init, resolve, reject }))
  const env = { active, pending, selectedClusterId: 7, blocked: false, queryMetric: query, queryMetricRange: query,
    setLoading: value => updates.push(['loading', value]), setResult: value => updates.push(['result', value]), setError: value => updates.push(['error', value]) }
  const run = new Function(...Object.keys(env), `${code}; return submit`)(...Object.values(env))
  const first = run({ mode: 'instant', metric_id: 'cluster_health' })
  assert.equal(calls[0].id, 7)
  assert.equal(calls[0].init.suppressErrorNotification, true)
  if (scenario.startsWith('late')) {
    const second = run({ mode: 'range', metric_id: 'client_read_bytes', start: 'start', end: 'end', step: '30s' })
    assert.equal(calls[0].init.signal.aborted, true)
    assert.deepEqual(calls[1].input, { metricId: 'client_read_bytes', start: 'start', end: 'end', step: '30s' })
    calls[1].resolve({ series: ['new'] }); await second
    const before = [...updates]
    if (scenario === 'late-success') calls[0].resolve({ series: ['old'] })
    else calls[0].reject(new Error('old error'))
    await first
    assert.deepEqual(updates, before)
  } else if (scenario === 'unmount') {
    active.current = false; pending.current.abort()
    const before = [...updates]
    calls[0].resolve({ series: ['old'] }); await first
    await run({ mode: 'instant', metric_id: 'cluster_health' })
    assert.equal(calls.length, 1)
    assert.deepEqual(updates, before)
  } else {
    calls[0].reject(new Error('')); await first
    assert.deepEqual(updates, [['loading', true], ['error', ''], ['result', null], ['error', '指标查询失败'], ['loading', false]])
  }
}
assert.ok(source.includes('return () => { active.current = false; pending.current?.abort() }'))
console.log('Metric queries isolate clusters and suppress superseded completions')
