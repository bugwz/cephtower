import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const compile = path => ts.transpileModule(readFileSync(new URL(path, import.meta.url), 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2022 } }).outputText
const definitions = {}
new Function('exports', 'require', compile('../src/pages/object/RgwSyncMetricLinks.tsx'))(definitions, () => ({}))
const metricSource = readFileSync(new URL('../src/pages/monitoring/MetricPage.tsx', import.meta.url), 'utf8')
const metricTree = ts.createSourceFile('metric.tsx', metricSource, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const helpers = metricTree.statements.filter(node => ts.isFunctionDeclaration(node) && ['filterMetricRows', 'metricLabelOptions'].includes(node.name.text))
const filters = {}
new Function('exports', ts.transpileModule(helpers.map(node => node.getText(metricTree)).join('\n'), { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText)(filters)
const states = [], refs = [], calls = [], ui = {}
let si = 0, ri = 0, deps, cleanup, pending
const jsx = (type, props) => ({ type, props })
const react = {
  useState: value => { const i = si++; if (!(i in states)) states[i] = value; return [states[i], value => { states[i] = value }] },
  useRef: value => refs[ri++] ?? (refs[ri - 1] = { current: value }),
  useEffect: (fn, next) => { if (JSON.stringify(next) !== JSON.stringify(deps)) { deps = next; pending = () => { cleanup?.(); cleanup = fn() } } }
}
new Function('exports', 'require', compile('../src/pages/object/RgwSyncMetrics.tsx'))(ui, name => ({
  react, 'react/jsx-runtime': { jsx, jsxs: jsx },
  antd: Object.fromEntries(['Alert', 'AutoComplete', 'Button', 'Card', 'Space', 'Table'].map(name => [name, name])),
  '../../api/external': { queryMetricRange: (clusterId, input, options) => new Promise((resolve, reject) => calls.push({ clusterId, input, options, resolve, reject })) },
  '../monitoring/MetricPage': { ...filters, metricSamples: series => series.values },
  '../monitoring/MetricTrend': { MetricTrend: 'MetricTrend' },
  '../monitoring/MetricNotices': { MetricNotices: 'MetricNotices' },
  './RgwSyncMetricLinks': definitions
}[name]))
const nodes = n => Array.isArray(n) ? n.flatMap(nodes) : n && typeof n === 'object' ? [n, ...nodes(n.props?.children)] : []
function render(clusterId = 7) { si = ri = 0; const result = nodes(ui.RgwSyncMetrics({ clusterId })); const effect = pending; pending = undefined; effect?.(); return result }
const tick = () => new Promise(resolve => setTimeout(resolve, 0))
render()
assert.equal(calls.length, 0, 'must not automatically query')
const button = render().find(n => n.type === 'Button')
button.props.onClick()
assert.equal(calls.length, 5)
assert.deepEqual(calls.map(call => call.input.metricId), definitions.rgwSyncMetrics.map(metric => metric.id))
assert.equal(new Set(calls.map(call => call.input.end)).size, 1)
for (const call of calls) {
  assert.equal(call.clusterId, 7)
  assert.equal(call.input.step, '60s')
  assert.equal(Date.parse(call.input.end) - Date.parse(call.input.start), 3600000)
  assert.equal(call.options.cache, 'no-store')
}
const result = { result_type: 'matrix', meta: { warnings: ['partial'] }, series: [{ metric: { source_zone: 'west' }, values: [[1, '9007199254740993']] }, { metric: { source_zone: 'east' }, values: [[1, 'NaN']] }] }
calls[0].resolve(result); calls[1].reject(new Error('secret')); calls[2].resolve({ ...result, series: [] }); calls[3].resolve({ ...result, result_type: 'vector' }); calls[4].resolve(result)
await tick()
let rendered = render()
assert.equal(rendered.filter(n => n.type === 'MetricTrend').length, 4)
assert.equal(rendered.filter(n => n.type === 'Alert' && n.props.type === 'warning').length, 2)
assert.equal(rendered.find(n => n.type === 'Table').props.dataSource[0][1], '9007199254740993')
assert.equal(rendered.find(n => n.type === 'MetricNotices').props.meta, result.meta)
assert.ok(rendered.some(n => n.type === 'pre' && n.props.children.includes('east')))
const selector = label => render().find(n => n.type === 'AutoComplete' && n.props['aria-label'] === label)
assert.deepEqual(selector('同步指标精确标签名').props.options, [{ value: 'source_zone', label: 'source_zone' }])
selector('同步指标精确标签名').props.onChange('source_zone')
assert.deepEqual(selector('同步指标精确标签值').props.options.map(option => option.value), ['east', 'west'])
selector('同步指标精确标签值').props.onChange('west')
assert.equal(render().filter(n => n.type === 'MetricTrend').length, 2)
assert.equal(render().filter(n => n.type === 'MetricNotices').length, 5, 'filters must not hide warnings')
selector('同步指标精确标签值').props.onChange('West')
assert.equal(render().filter(n => n.type === 'MetricTrend').length, 0)
assert.equal(calls.length, 5, 'filtering must not request or claim server scoping')
render().find(n => n.type === 'Button' && n.props.children === '清除标签筛选').props.onClick()
assert.equal(render().filter(n => n.type === 'MetricTrend').length, 4)
selector('同步指标精确标签名').props.onChange('source_zone')
selector('同步指标精确标签值').props.onChange('west')
button.props.onClick(); render(8)
assert.equal(states[1], '')
assert.equal(states[2], '')
assert.ok(calls.slice(5).every(call => call.options.signal.aborted))
button.props.onClick(); assert.equal(calls.length, 10, 'stale handlers must not query old clusters')
calls.slice(5).forEach(call => call.resolve(result)); await tick()
assert.equal(render(8).filter(n => n.type === 'MetricTrend').length, 0)
render(8).find(n => n.type === 'Button').props.onClick(); cleanup()
assert.ok(calls.slice(10).every(call => call.options.signal.aborted))
calls.slice(10).forEach(call => call.resolve(result)); await tick()
assert.equal(states[0].rows, undefined)
assert.ok(readFileSync(new URL('../src/pages/object/RgwSyncStatus.tsx', import.meta.url), 'utf8').includes('<RgwSyncMetrics key={clusterId'))
console.log('RGW sync overview isolates query scope and preserves per-series data and failures')
