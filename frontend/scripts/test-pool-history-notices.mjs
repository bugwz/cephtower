import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/pages/cluster/PoolIOHistory.tsx', import.meta.url), 'utf8')
const tree = ts.createSourceFile('pool.tsx', source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX)
const component = tree.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'PoolIOHistory')
const queryNode = component.body.statements.find(node => ts.isFunctionDeclaration(node) && node.name.text === 'query')
const js = ts.transpileModule(queryNode.getText(tree), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText
const metrics = [{ id: 'read', title: '读取吞吐率' }, { id: 'write', title: '写入吞吐率' }]
for (const cancel of [false, true]) {
  const controller = { current: null }, updates = []
  const query = new Function('loading', 'controller', 'setLoading', 'setData', 'setError', 'metrics', 'queryMetricRange', 'clusterId', 'poolId', 'poolHistoryPoints', `${js}; return query`)(false, controller, value => updates.push(['loading', value]), value => updates.push(['data', value]), value => updates.push(['error', value]), metrics, async (id, input, options) => {
    assert.equal(id, 7)
    assert.equal(options.signal, controller.current.signal)
    assert.equal(input.step, '30s')
    if (cancel) controller.current.abort()
    return { series: [], meta: { warnings: [input.metricId] } }
  }, 7, 12, (response, poolId) => { assert.equal(poolId, 12); return response.series })
  await query()
  assert.deepEqual(updates.slice(0, 3), [['loading', true], ['data', null], ['error', '']])
  assert.deepEqual(updates.slice(3), cancel ? [] : [['data', metrics.map(metric => ({ points: [], meta: { warnings: [metric.id] } }))], ['loading', false]])
}
let data = Array.from({ length: 4 }, (_, index) => ({ points: [], meta: { infos: [`notice ${index}`] } }))
const exports = {}
new Function('exports', 'require', 'React', ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, target: ts.ScriptTarget.ES2022 } }).outputText)(exports, name => ({
  antd: { Alert: 'Alert', Button: 'Button', Card: 'Card', Empty: 'Empty', Space: 'Space', Typography: { Text: 'Text' } },
  react: { useState: value => [value === null ? data : value, () => {}], useRef: value => ({ current: value }), useEffect: () => {} },
  '../../api/client': {}, '../../api/external': {}, '../monitoring/MetricNotices': { MetricNotices: 'MetricNotices' }
}[name]), { createElement: (type, props, ...children) => ({ type, props, children }) })
const find = node => node?.type === 'MetricNotices' ? [node] : node?.children?.flat(Infinity).flatMap(find) ?? []
const notices = find(exports.PoolIOHistory({ clusterId: 7, poolId: 12 }))
assert.deepEqual(notices.map(node => node.props.meta), data.map(row => row.meta))
assert.deepEqual(notices.map(node => node.props.source), ['读取吞吐率', '写入吞吐率', '读取 IOPS', '写入 IOPS'])
data = null
assert.deepEqual(find(exports.PoolIOHistory({ clusterId: 7, poolId: 12 })), [])
console.log('Pool history preserves per-metric notices and ignores cancelled responses')
