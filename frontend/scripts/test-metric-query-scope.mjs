import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'

const source = readFileSync(new URL('../src/pages/monitoring/MetricPage.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('MetricPage.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
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
