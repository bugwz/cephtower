import assert from 'node:assert/strict'
import './test-metric-response.mjs'
import './test-metric-notices.mjs'
import './test-alert-columns.mjs'
import './test-metric-trend.mjs'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/MetricPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('MetricPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const optionsNode = tree.statements.filter(ts.isVariableStatement).flatMap(node => [...node.declarationList.declarations]).find(node => node.name.getText(tree) === 'metricOptions')
const options = new Function(`return ${optionsNode.initializer.getText(tree)}`)()
for (const id of ['rgw_delete_latency_ms', 'rgw_copy_latency_ms', 'rgw_list_objects_latency_ms', 'rgw_list_buckets_latency_ms', 'rgw_delete_buckets_latency_ms']) {
  assert.equal(options.filter(option => option.value === id).length, 1)
  assert.ok(options.find(option => option.value === id).description.includes('操作数加权平均'))
  assert.ok(options.find(option => option.value === id).label.includes('ms'))
}
for (const id of ['rgw_get_ops_total', 'rgw_put_ops_total', 'rgw_delete_ops_total', 'rgw_copy_ops_total', 'rgw_list_objects_total', 'rgw_list_buckets_total', 'rgw_delete_buckets_total']) {
  assert.equal(options.filter(option => option.value === id).length, 1)
  assert.ok(options.find(option => option.value === id).label.includes('累计操作数'))
}
assert.ok(options.find(option => option.value === 'rgw_sync_delta_seconds').description.includes('原始 gauge'))
assert.ok(options.find(option => option.value === 'rgw_sync_poll_latency_ms').description.includes('sum rate / count rate'))
for (const id of ['rgw_sync_bytes_rate', 'rgw_sync_objects_rate', 'rgw_sync_errors_rate']) {
  assert.equal(options.filter(option => option.value === id).length, 1)
  assert.ok(options.find(option => option.value === id).description.includes('source_zone'))
}
for (const id of ['smb_cluster_nodes', 'smb_cluster_sessions_mean', 'smb_cluster_users_mean', 'smb_cluster_shares_mean']) assert.equal(options.filter(option => option.value === id).length, 1)
const filterNode = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'filterMetricRows')
const filterExports = {}
const labelOptionsNode = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'metricLabelOptions')
const labelOptionsExports = {}
new Function('exports', ts.transpileModule(labelOptionsNode.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(labelOptionsExports)
const optionRows = [{ labels: '{"instance":"a","empty":"","count":2}' }, { labels: '{"instance":"a"}' }, { labels: '{"instance":"A"}' }, { labels: 'null' }, { labels: '[]' }, { labels: 'bad' }]
assert.deepEqual(labelOptionsExports.metricLabelOptions(optionRows), [{ value: 'empty', label: 'empty' }, { value: 'instance', label: 'instance' }])
assert.deepEqual(labelOptionsExports.metricLabelOptions(optionRows, 'instance'), [{ value: 'A', label: 'A' }, { value: 'a', label: 'a' }])
assert.deepEqual(labelOptionsExports.metricLabelOptions(optionRows, 'empty'), [{ value: '', label: '空字符串' }])
assert.deepEqual(labelOptionsExports.metricLabelOptions(optionRows, 'missing'), [])
assert.deepEqual(labelOptionsExports.metricLabelOptions([]), [])
new Function('exports', ts.transpileModule(filterNode.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(filterExports)
const labelRows = [{ labels: '{"instance":"node.1","operation":"read","empty":""}' }, { labels: '{"instance":"nodeX1","operation":"write"}' }, { labels: '{"instance":"Node.1"}' }, { labels: 'null' }, { labels: 'invalid' }]
assert.equal(filterExports.filterMetricRows(labelRows, '', ''), labelRows)
assert.deepEqual(filterExports.filterMetricRows(labelRows, 'instance', 'node.1'), [labelRows[0]])
assert.deepEqual(filterExports.filterMetricRows(labelRows, 'empty', ''), [labelRows[0]])
assert.deepEqual(filterExports.filterMetricRows(labelRows, 'toString', ''), [])
assert.deepEqual(filterExports.filterMetricRows(labelRows, 'operation', 'write'), [labelRows[1]])
assert.deepEqual(filterExports.filterMetricRows(labelRows, 'instance', '.*'), [])
const presetNode = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'metricPreset')
const presetCode = ts.transpileModule(presetNode.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText
const presetExports = {}
new Function('exports', 'metricOptions', presetCode)(presetExports, options)
for (const value of [null, '', 'sum(secret)', 'unknown']) assert.equal(presetExports.metricPreset(value), 'cluster_health')
for (const option of options) assert.equal(presetExports.metricPreset(option.value), option.value)
assert.equal(options.filter(option => option.value === 'smb_request_duration_rate').length, 1)
assert.ok(options.find(option => option.value === 'smb_request_duration_rate').label.includes('µs/s'))
assert.ok(options.find(option => option.value === 'smb_request_duration_rate').description.includes('非单请求平均延迟'))
for (const id of ['smb_metrics_status', 'smb_sessions', 'smb_users', 'smb_share_activity', 'smb_in_bytes_rate', 'smb_out_bytes_rate', 'smb_request_rate']) assert.equal(options.filter(option => option.value === id).length, 1)
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
let metricParam = 'smb_metrics_status'
new Function('exports', 'require', 'useClusterContext', 'MetricContent', 'useSearchParams', 'metricPreset', ts.transpileModule(wrapper.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX } }).outputText)(ui, () => ({ jsx: (type, props, key) => ({ type, props, key }) }), () => ({ selectedClusterId: clusterId }), 'MetricContent', () => [new URLSearchParams({ metric: metricParam })], presetExports.metricPreset)
const scopes = [undefined, 1, 2].map(id => { clusterId = id; return ui.MetricPage() })
assert.equal(new Set(scopes.map(item => item.key)).size, 3)
assert.equal(scopes[2].props.selectedClusterId, 2)
assert.equal(scopes[2].props.initialMetric, 'smb_metrics_status')
metricParam = 'smb_sessions'
assert.notEqual(ui.MetricPage().key, scopes[2].key)
const content = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MetricContent')
const submit = content.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'submit')
const code = ts.transpileModule(submit.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
for (const scenario of ['late-success', 'late-failure', 'unmount', 'failure']) {
  const calls = [], updates = []
  const active = { current: true }, pending = { current: null }
  const query = (id, input, init) => new Promise((resolve, reject) => calls.push({ id, input, init, resolve, reject }))
  const env = { active, pending, selectedClusterId: 7, blocked: false, queryMetric: query, queryMetricRange: query,
    queriedMetric: { current: 'cluster_health' }, setLabelKey: value => updates.push(['labelKey', value]), setLabelValue: value => updates.push(['labelValue', value]),
    setLoading: value => updates.push(['loading', value]), setResult: value => updates.push(['result', value]), setError: value => updates.push(['error', value]) }
  const run = new Function(...Object.keys(env), `${code}; return submit`)(...Object.values(env))
  const first = run({ mode: 'instant', metric_id: 'cluster_health' })
  assert.equal(calls[0].id, 7)
  assert.equal(calls[0].init.suppressErrorNotification, true)
  if (scenario.startsWith('late')) {
    const second = run({ mode: 'range', metric_id: 'client_read_bytes', start: 'start', end: 'end', step: '30s' })
    assert.equal(env.queriedMetric.current, 'client_read_bytes')
    assert.deepEqual(updates.filter(([kind]) => kind.startsWith('label')), [['labelKey', ''], ['labelValue', '']])
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
    assert.equal(updates.some(([kind]) => kind.startsWith('label')), false)
    calls[0].reject(new Error('')); await first
    assert.deepEqual(updates, [['loading', true], ['error', ''], ['result', null], ['error', '指标查询失败'], ['loading', false]])
  }
}
assert.ok(source.includes('return () => { active.current = false; pending.current?.abort() }'))
console.log('Metric queries isolate clusters and suppress superseded completions')
