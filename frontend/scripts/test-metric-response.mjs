import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
const source = readFileSync(new URL('../src/api/external.ts', import.meta.url), 'utf8')
const tree = ts.createSourceFile('external.ts', source, ts.ScriptTarget.Latest, true)
const code = tree.statements.filter(node => ts.isFunctionDeclaration(node) && ['queryMetric', 'queryMetricRange', 'readMetric'].includes(node.name.text)).map(node => node.getText(tree)).join('\n')
let response, lastRequest
const exports = {}
new Function('exports', 'request', 'jsonInit', ts.transpileModule(code, { compilerOptions: { target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.CommonJS } }).outputText)(exports, async (path, init) => { lastRequest = { path, init }; return response }, (method, body, init) => ({ method, body, ...init }))
const controller = new AbortController()
for (const [type, query] of [
  ['vector', () => exports.queryMetric(7, { metricId: 'cluster_health' }, { signal: controller.signal })],
  ['matrix', () => exports.queryMetricRange(7, { metricId: 'pool_read_bytes', start: 'start', end: 'end', step: '30s' }, { signal: controller.signal })]
]) {
  for (const series of [[], [{ metric: { instance: 'a' }, value: [1, '9007199254740993'], values: [[1, 'NaN'], [2, '+Inf']] }]]) {
    response = { result_type: type, series, meta: { source: 'prometheus' } }
    assert.equal(await query(), response)
    assert.equal(lastRequest.init.body.cluster_id, 7)
    assert.equal(lastRequest.init.signal, controller.signal)
    assert.ok(lastRequest.path.startsWith(type === 'vector' ? '/metric/query?' : '/metric/range?'))
  }
  for (const invalid of [null, undefined, [], {}, { result_type: type }, { result_type: type, series: null }, { result_type: type, series: [null] }, { result_type: type, series: [{ metric: {} }, 1] }, { result_type: type, series: [[]] }, { result_type: type === 'vector' ? 'matrix' : 'vector', series: [] }]) {
    response = invalid
    await assert.rejects(query, /指标响应格式异常/)
  }
}
console.log('Metric API rejects malformed envelopes without discarding native sample values')
