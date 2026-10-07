import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/MonSessionHistory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('history.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const fn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'monSessionSeries')
const exports = {}
new Function('exports', ts.transpileModule(fn.getText(tree), { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const row = { metric: { cluster: 'fsid', ceph_daemon: 'mon.a', instance: 'host1' }, values: [[1, '9007199254740993'], [2, 'NaN']] }
const data = { result_type: 'matrix', meta: { cluster_fsid: 'fsid', mon_name: 'a' }, series: [row] }
assert.equal(exports.monSessionSeries(data, 'a'), data.series)
assert.deepEqual(exports.monSessionSeries({ ...data, series: [] }, 'a'), [])
for (const bad of [{ ...data, meta: {} }, { ...data, result_type: 'vector' }, { ...data, series: [null] }, { ...data, series: [{ ...row, metric: { cluster: 'other', ceph_daemon: 'mon.a' } }] }, { ...data, series: [{ ...row, values: null }] }]) assert.throws(() => exports.monSessionSeries(bad, 'a'))
assert.throws(() => exports.monSessionSeries(data, 'b'))
assert.ok(source.includes('controller.current?.abort()'))
assert.ok(source.includes('if (!active.signal.aborted) setData(result)'))
assert.ok(source.includes('metricId: \'mon_sessions\', monName'))
assert.ok(readFileSync(new URL('../src/pages/cluster/MonDetailPage.tsx', import.meta.url), 'utf8').includes('<MonSessionHistory key={JSON.stringify([selectedClusterId, monName])}'))
console.log('MON session history keeps raw samples and rejects cross-scope sequences')
const metricSource = readFileSync(new URL('../src/pages/monitoring/MetricPage.tsx', import.meta.url), 'utf8')
const metricTree = ts.createSourceFile('metric.tsx', metricSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const metricFn = metricTree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'metricSamples')
const countFn = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'monSessionSamples')
new Function('exports', ts.transpileModule(`${metricFn.getText(metricTree)}\n${countFn.getText(tree)}`, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)(exports)
const raw = ['0', '1', '9007199254740993', '1e3', '-1', '0.5', 'NaN', '+Inf']
const samples = exports.monSessionSamples({ values: raw.map((value, index) => [index, value]) })
assert.deepEqual(samples.map(sample => sample.value), raw)
assert.ok(samples.slice(0, 4).every(sample => sample.status === '有效值'))
assert.ok(samples.slice(4, 6).every(sample => sample.status === '会话数异常（应为非负整数）'))
assert.ok(samples.slice(6).every(sample => sample.status === '非有限值（不代表零）'))
assert.ok(source.includes('<MetricTrend samples={monSessionSamples(row)}'))
for (const value of ['9007199254740993.5', '1e-9999', '-1e-9999', '1.00000000000000001']) assert.equal(exports.monSessionSamples({ values: [[1, value]] })[0].status, '会话数异常（应为非负整数）')
for (const value of ['-0', '0e-9999', '100e-2', '1.20e2']) assert.equal(exports.monSessionSamples({ values: [[1, value]] })[0].status, '有效值')
const component = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'MonSessionHistory')
const renderExports = {}
let renderedData = { ...data, series: [], meta: { ...data.meta, warnings: ['partial data'], infos: ['sample omitted'] } }
new Function('exports', 'require', 'React', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText)(renderExports, name => ({
  antd: { Alert: 'Alert', Button: 'Button', Card: 'Card', Space: 'Space' },
  react: { useState: value => [value === null ? renderedData : value, () => {}], useRef: value => ({ current: value }), useEffect: () => {} },
  '../../api/external': {}, '../../components/AppTable': { AppTable: 'AppTable' },
  '../monitoring/MetricPage': { metricSamples: () => [] }, '../monitoring/MetricTrend': { MetricTrend: 'MetricTrend' },
  '../monitoring/MetricNotices': { MetricNotices: 'MetricNotices' }
}[name]), { createElement: (type, props, ...children) => ({ type, props, children }) })
const findNotices = node => node?.type === 'MetricNotices' ? [node] : node?.children?.flat(Infinity).flatMap(findNotices) ?? []
let notices = findNotices(renderExports.MonSessionHistory({ clusterId: 7, monName: 'a' }))
assert.equal(notices.length, 1)
assert.equal(notices[0].props.meta, renderedData.meta)
renderedData = null
notices = findNotices(renderExports.MonSessionHistory({ clusterId: 7, monName: 'a' }))
assert.equal(notices[0].props.meta, undefined)
const queryNode = component.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'query')
const js = ts.transpileModule(queryNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2020 } }).outputText
for (const leave of [false, true]) {
  const controller = { current: null }, updates = []
  const query = new Function('controller', 'setLoading', 'setError', 'setData', 'queryMetricRange', 'clusterId', 'monName', 'monSessionSeries', `${js}; return query`)(controller, value => updates.push(['busy', value]), value => updates.push(['error', value]), value => updates.push(['data', value]), async (id, input, init) => {
    assert.equal(id, 7); assert.equal(input.monName, 'a'); assert.equal(input.metricId, 'mon_sessions')
    assert.equal(Date.parse(input.end) - Date.parse(input.start), 3600000)
    assert.equal(init.signal, controller.current.signal)
    if (leave) controller.current.abort()
    return data
  }, 7, 'a', exports.monSessionSeries)
  await query()
  assert.deepEqual(updates.slice(0, 3), [['busy', true], ['error', ''], ['data', null]])
  assert.deepEqual(updates.slice(3), leave ? [] : [['data', data], ['busy', false]])
}
